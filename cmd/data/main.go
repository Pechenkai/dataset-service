package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ppo/internal/config"
	"ppo/internal/contracts"
	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
	"ppo/internal/logger"
	"ppo/internal/messaging"
	"ppo/internal/messaging/rabbit"
	"ppo/internal/metrics"
	"strings"
)

type dataApp struct {
	cfg         *config.Config
	logger      *zap.Logger
	logClose    func()
	bus         messaging.Bus
	catRepo     *postqbuild.CategoryRepo
	dsRepo      *postqbuild.DatasetRepo
	verRepo     *postqbuild.VersionRepo
	notifRepo   *postqbuild.NotificationRepo
	accessRepo  *postqbuild.AccessRequestRepo
	reviewRepo  *postqbuild.ReviewRepo
	subRepo     *postqbuild.SubscriptionRepo
	userRepo    *postqbuild.UserRepo
	replyTo     string
	consumeFrom string
	processed   sync.Map
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	zapLogger, closeLog, err := logger.NewLogger(cfg.LogCfg)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer closeLog()

	dbPool, err := postqbuild.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer dbPool.Close()

	bus, err := rabbit.New(cfg.Broker, zapLogger)
	if err != nil {
		log.Fatalf("init broker: %v", err)
	}
	defer bus.Close()

	app := &dataApp{
		cfg:         cfg,
		logger:      zapLogger,
		logClose:    closeLog,
		bus:         bus,
		dsRepo:      postqbuild.NewDatasetRepo(dbPool, zapLogger),
		verRepo:     postqbuild.NewVersionRepo(dbPool, zapLogger),
		catRepo:     postqbuild.NewCategoryRepo(dbPool, zapLogger),
		notifRepo:   postqbuild.NewNotificationRepo(dbPool, zapLogger),
		accessRepo:  postqbuild.NewAccessRequestRepo(dbPool, zapLogger),
		reviewRepo:  postqbuild.NewReviewRepo(dbPool, zapLogger),
		subRepo:     postqbuild.NewSubscriptionRepo(dbPool, zapLogger),
		userRepo:    postqbuild.NewUserRepo(dbPool, zapLogger),
		replyTo:     cfg.Broker.DataToCoreQueue,
		consumeFrom: cfg.Broker.CoreToDataQueue,
	}

	if err := bus.Consume(ctx, app.consumeFrom, app.handleCoreMessage); err != nil {
		log.Fatalf("consume commands: %v", err)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler: app.router(),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	zapLogger.Info("data service started", zap.String("addr", server.Addr))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("data http server error: %v", err)
	}
}

func (a *dataApp) router() http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.Handler().ServeHTTP(w, r)
	})

	return r
}

func (a *dataApp) handleCoreMessage(ctx context.Context, msg contracts.Envelope) (err error) {
	if msg.ID != "" {
		if _, exists := a.processed.Load(msg.ID); exists {
			if a.logger != nil {
				a.logger.Debug("duplicate message ignored", zap.String("id", msg.ID))
			}
			return nil
		}
	}
	defer func() {
		if err == nil && msg.ID != "" {
			a.processed.Store(msg.ID, struct{}{})
		}
	}()
	switch msg.Type {
	case contracts.MsgTypeHealthRequest:
		resp := contracts.HealthResponse{
			Service:   contracts.DefaultServiceNameData,
			Status:    "ok",
			Timestamp: time.Now(),
		}
		payload, _ := json.Marshal(resp)
		return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
			Type:          contracts.MsgTypeHealthResponse,
			Source:        contracts.DefaultServiceNameData,
			CorrelationID: msg.CorrelationID,
			OccurredAt:    time.Now(),
			Payload:       payload,
		})
	case contracts.MsgTypeCategoryListRequest:
		return a.handleListCategories(ctx, msg)
	case contracts.MsgTypeDatasetListRequest:
		return a.handleDatasetList(ctx, msg)
	case contracts.MsgTypeDatasetGetRequest:
		return a.handleDatasetGet(ctx, msg)
	case contracts.MsgTypeDatasetCreateRequest:
		return a.handleDatasetCreate(ctx, msg)
	case contracts.MsgTypeDatasetUpdateRequest:
		return a.handleDatasetUpdate(ctx, msg)
	case contracts.MsgTypeDatasetDeleteRequest:
		return a.handleDatasetDelete(ctx, msg)
	case contracts.MsgTypeVersionsListRequest:
		return a.handleVersionList(ctx, msg)
	case contracts.MsgTypeVersionGetRequest:
		return a.handleVersionGet(ctx, msg)
	case contracts.MsgTypeVersionAddRequest:
		return a.handleVersionAdd(ctx, msg)
	case contracts.MsgTypeVersionDeleteRequest:
		return a.handleVersionDelete(ctx, msg)
	case contracts.MsgTypeNotificationsRequest:
		return a.handleNotifications(ctx, msg)
	case contracts.MsgTypeAccessPendingRequest:
		return a.handleAccessPending(ctx, msg)
	case contracts.MsgTypeAccessFindRequest:
		return a.handleAccessFind(ctx, msg)
	case contracts.MsgTypeAccessUpdateRequest:
		return a.handleAccessUpdate(ctx, msg)
	case contracts.MsgTypeNotificationCreateRequest:
		return a.handleNotificationCreate(ctx, msg)
	case contracts.MsgTypeNotificationMarkRequest:
		return a.handleNotificationMark(ctx, msg)
	case contracts.MsgTypeReviewCreateRequest:
		return a.handleReviewCreate(ctx, msg)
	case contracts.MsgTypeReviewUpdateRequest:
		return a.handleReviewUpdate(ctx, msg)
	case contracts.MsgTypeReviewDeleteRequest:
		return a.handleReviewDelete(ctx, msg)
	case contracts.MsgTypeSubscriptionCreateRequest:
		return a.handleSubscriptionCreate(ctx, msg)
	case contracts.MsgTypeSubscriptionDeleteRequest:
		return a.handleSubscriptionDelete(ctx, msg)
	case contracts.MsgTypeCategoryCreateRequest:
		return a.handleCategoryCreate(ctx, msg)
	case contracts.MsgTypeCategoryUpdateRequest:
		return a.handleCategoryUpdate(ctx, msg)
	case contracts.MsgTypeCategoryDeleteRequest:
		return a.handleCategoryDelete(ctx, msg)
	case contracts.MsgTypeUserCreateRequest:
		return a.handleUserCreate(ctx, msg)
	case contracts.MsgTypeUserUpdateRequest:
		return a.handleUserUpdate(ctx, msg)
	case contracts.MsgTypeUserDeleteRequest:
		return a.handleUserDelete(ctx, msg)
	case contracts.MsgTypeUserGetRequest:
		return a.handleUserGet(ctx, msg)
	default:
		return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
			Type:          msg.Type,
			Source:        contracts.DefaultServiceNameData,
			CorrelationID: msg.CorrelationID,
			OccurredAt:    time.Now(),
			Error: &contracts.ErrorPayload{
				Code:    contracts.MessageErrorCodeUnknownType,
				Message: "unsupported message type",
			},
		})
	}
}

func (a *dataApp) handleListCategories(ctx context.Context, msg contracts.Envelope) error {
	list, err := a.catRepo.FindAll(ctx)
	if err != nil {
		if a.logger != nil {
			a.logger.Error("list categories failed", zap.Error(err))
		}
		return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
			Type:          contracts.MsgTypeCategoryListResponse,
			Source:        contracts.DefaultServiceNameData,
			CorrelationID: msg.CorrelationID,
			OccurredAt:    time.Now(),
			Error: &contracts.ErrorPayload{
				Code:    contracts.MessageErrorCodeProcessError,
				Message: err.Error(),
			},
		})
	}

	resp := contracts.ListCategoriesResponse{Categories: list}
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeCategoryListResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleCategoryCreate(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.CategoryCreateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeCategoryCreateResponse, fmt.Errorf("invalid category"))
	}
	cat, err := entities.NewCategory(req.Name, req.Description)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeCategoryCreateResponse, err)
	}
	if err := a.catRepo.Create(ctx, cat); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeCategoryCreateResponse, err)
	}

	resp := contracts.CategoryCreateResponse{ID: cat.ID}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeCategoryCreateResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleCategoryUpdate(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.CategoryUpdateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.ID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeCategoryUpdateResponse, fmt.Errorf("invalid category"))
	}
	cat, err := entities.NewCategory(req.Name, req.Description)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeCategoryUpdateResponse, err)
	}
	cat.ID = req.ID
	if err := a.catRepo.Update(ctx, cat); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeCategoryUpdateResponse, err)
	}
	resp := contracts.CategoryUpdateResponse{Updated: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeCategoryUpdateResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleCategoryDelete(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.CategoryDeleteRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.ID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeCategoryDeleteResponse, fmt.Errorf("invalid category id"))
	}
	if err := a.catRepo.Delete(ctx, req.ID); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeCategoryDeleteResponse, err)
	}
	resp := contracts.CategoryDeleteResponse{Deleted: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeCategoryDeleteResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleUserCreate(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.UserCreateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeUserCreateResponse, fmt.Errorf("invalid user"))
	}
	u, err := entities.NewUser(req.Username, req.Email, req.Password, req.Country, req.Role, time.Now())
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeUserCreateResponse, err)
	}
	if err := a.userRepo.Create(ctx, u); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeUserCreateResponse, err)
	}
	resp := contracts.UserCreateResponse{ID: u.ID}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeUserCreateResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleUserUpdate(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.UserUpdateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.ID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeUserUpdateResponse, fmt.Errorf("invalid user"))
	}
	user, err := a.userRepo.FindByID(ctx, req.ID)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeUserUpdateResponse, err)
	}
	if user == nil {
		return a.publishError(ctx, msg, contracts.MsgTypeUserUpdateResponse, fmt.Errorf("user not found"))
	}
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Country != "" {
		user.Country = req.Country
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Block != nil {
		user.IsBlocked = *req.Block
	}
	if err := a.userRepo.Update(ctx, user); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeUserUpdateResponse, err)
	}
	resp := contracts.UserUpdateResponse{Updated: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeUserUpdateResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleUserDelete(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.UserDeleteRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.ID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeUserDeleteResponse, fmt.Errorf("invalid user id"))
	}
	if err := a.userRepo.Delete(ctx, req.ID); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeUserDeleteResponse, err)
	}
	resp := contracts.UserDeleteResponse{Deleted: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeUserDeleteResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleUserGet(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.UserGetRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.ID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeUserGetResponse, fmt.Errorf("invalid user id"))
	}
	user, err := a.userRepo.FindByID(ctx, req.ID)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeUserGetResponse, err)
	}
	resp := contracts.UserGetResponse{User: user}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeUserGetResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleDatasetList(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.DatasetListRequest
	if len(msg.Payload) > 0 {
		_ = json.Unmarshal(msg.Payload, &req)
	}

	var (
		list []*entities.Dataset
		err  error
	)
	switch {
	case req.OwnerID != nil:
		list, err = a.dsRepo.FindByUserID(ctx, *req.OwnerID)
	case req.OnlyPublic:
		list, err = a.dsRepo.FindPublic(ctx)
	default:
		list, err = a.dsRepo.FindAll(ctx)
	}
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetListResponse, err)
	}

	resp := contracts.DatasetListResponse{Datasets: list}
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeDatasetListResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleDatasetCreate(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.DatasetCreateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetCreateResponse, fmt.Errorf("invalid dataset payload"))
	}
	ds, err := entities.NewDataset(req.Name, req.Description, req.OwnerID, req.CategoryID, req.IsPublic, time.Now())
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetCreateResponse, err)
	}
	if err := a.dsRepo.Create(ctx, ds); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetCreateResponse, err)
	}
	resp := contracts.DatasetCreateResponse{ID: ds.ID}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeDatasetCreateResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleDatasetUpdate(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.DatasetUpdateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.ID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetUpdateResponse, fmt.Errorf("invalid dataset payload"))
	}
	existing, err := a.dsRepo.FindByID(ctx, req.ID)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetUpdateResponse, err)
	}
	if existing == nil {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetUpdateResponse, fmt.Errorf("dataset not found"))
	}
	existing.Name = req.Name
	existing.Description = req.Description
	existing.CategoryID = req.CategoryID
	existing.IsPublic = req.IsPublic
	if err := a.dsRepo.Update(ctx, existing); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetUpdateResponse, err)
	}
	resp := contracts.DatasetUpdateResponse{Updated: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeDatasetUpdateResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleDatasetDelete(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.DatasetDeleteRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.ID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetDeleteResponse, fmt.Errorf("invalid dataset id"))
	}
	if err := a.dsRepo.Delete(ctx, req.ID); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetDeleteResponse, err)
	}
	resp := contracts.DatasetDeleteResponse{Deleted: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeDatasetDeleteResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleVersionAdd(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.VersionAddRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.DatasetID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeVersionAddResponse, fmt.Errorf("invalid version payload"))
	}
	ver, err := entities.NewDatasetVersion(req.Number, req.Filepath, req.ChangeLog, req.DatasetID, time.Now())
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeVersionAddResponse, err)
	}
	if err := a.verRepo.Create(ctx, ver); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeVersionAddResponse, err)
	}
	resp := contracts.VersionAddResponse{ID: ver.ID}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeVersionAddResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleVersionDelete(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.VersionDeleteRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.VersionID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeVersionDeleteResponse, fmt.Errorf("invalid version id"))
	}
	if err := a.verRepo.Delete(ctx, req.VersionID); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeVersionDeleteResponse, err)
	}
	resp := contracts.VersionDeleteResponse{Deleted: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeVersionDeleteResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleReviewCreate(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.ReviewCreateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeReviewCreateResponse, fmt.Errorf("invalid review"))
	}
	rv, err := entities.NewReview(req.UserID, req.DatasetID, req.Rating, time.Now(), req.Text)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeReviewCreateResponse, err)
	}
	if err := a.reviewRepo.Create(ctx, rv); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeReviewCreateResponse, err)
	}
	resp := contracts.ReviewCreateResponse{ID: rv.ID}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeReviewCreateResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleReviewUpdate(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.ReviewUpdateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.ID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeReviewUpdateResponse, fmt.Errorf("invalid review"))
	}
	rv := &entities.Review{
		ID:     req.ID,
		Rating: req.Rating,
		Text:   req.Text,
	}
	if err := a.reviewRepo.Update(ctx, rv); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeReviewUpdateResponse, err)
	}
	resp := contracts.ReviewUpdateResponse{Updated: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeReviewUpdateResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleReviewDelete(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.ReviewDeleteRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.ID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeReviewDeleteResponse, fmt.Errorf("invalid review id"))
	}
	if err := a.reviewRepo.Delete(ctx, req.ID); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeReviewDeleteResponse, err)
	}
	resp := contracts.ReviewDeleteResponse{Deleted: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeReviewDeleteResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleSubscriptionCreate(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.SubscriptionCreateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeSubscriptionCreateResponse, fmt.Errorf("invalid subscription"))
	}
	sub, err := entities.NewSubscription(req.UserID, req.DatasetID, time.Now())
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeSubscriptionCreateResponse, err)
	}
	if err := a.subRepo.Create(ctx, sub); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeSubscriptionCreateResponse, err)
	}
	resp := contracts.SubscriptionCreateResponse{Created: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeSubscriptionCreateResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleSubscriptionDelete(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.SubscriptionDeleteRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.UserID == 0 || req.DatasetID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeSubscriptionDeleteResponse, fmt.Errorf("invalid subscription ids"))
	}
	if err := a.subRepo.Unsubscribe(ctx, req.UserID, req.DatasetID); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeSubscriptionDeleteResponse, err)
	}
	resp := contracts.SubscriptionDeleteResponse{Deleted: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeSubscriptionDeleteResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}
func (a *dataApp) handleDatasetGet(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.DatasetGetRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.ID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetGetResponse, fmt.Errorf("invalid dataset id"))
	}

	ds, err := a.dsRepo.FindByID(ctx, req.ID)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeDatasetGetResponse, err)
	}
	resp := contracts.DatasetGetResponse{Dataset: ds}
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeDatasetGetResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleVersionList(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.VersionsListRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.DatasetID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeVersionsListResponse, fmt.Errorf("invalid dataset id"))
	}

	vers, err := a.verRepo.FindByDatasetID(ctx, req.DatasetID)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeVersionsListResponse, err)
	}
	resp := contracts.VersionsListResponse{Versions: vers}
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeVersionsListResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleVersionGet(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.VersionGetRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.VersionID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeVersionGetResponse, fmt.Errorf("invalid version id"))
	}
	ver, err := a.verRepo.FindByID(ctx, req.VersionID)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeVersionGetResponse, err)
	}
	resp := contracts.VersionGetResponse{Version: ver}
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeVersionGetResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleNotifications(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.NotificationsRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.UserID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeNotificationsResponse, fmt.Errorf("invalid user id"))
	}

	items, err := a.notifRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeNotificationsResponse, err)
	}

	resp := contracts.NotificationsResponse{Notifications: items}
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeNotificationsResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleAccessPending(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.AccessPendingRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.OwnerID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeAccessPendingResponse, fmt.Errorf("invalid owner id"))
	}
	items, err := a.accessRepo.ListPendingByOwner(ctx, req.OwnerID)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeAccessPendingResponse, err)
	}
	resp := contracts.AccessPendingResponse{Requests: items}
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeAccessPendingResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleAccessFind(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.AccessFindRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.DatasetID == 0 || req.UserID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeAccessFindResponse, fmt.Errorf("invalid ids"))
	}
	item, err := a.accessRepo.Find(ctx, req.DatasetID, req.UserID)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeAccessFindResponse, err)
	}
	resp := contracts.AccessFindResponse{Request: item}
	payload, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeAccessFindResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleAccessUpdate(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.AccessUpdateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.RequestID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeAccessUpdateResponse, fmt.Errorf("invalid request"))
	}
	if _, ok := entities.ValidAccessStatuses[req.Status]; !ok {
		return a.publishError(ctx, msg, contracts.MsgTypeAccessUpdateResponse, fmt.Errorf("invalid status"))
	}
	if err := a.accessRepo.UpdateStatus(ctx, req.RequestID, string(req.Status)); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeAccessUpdateResponse, err)
	}
	resp := contracts.AccessUpdateResponse{Updated: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeAccessUpdateResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleNotificationCreate(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.NotificationCreateRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.UserID == 0 || req.DatasetID == 0 || strings.TrimSpace(req.Message) == "" {
		return a.publishError(ctx, msg, contracts.MsgTypeNotificationCreateResponse, fmt.Errorf("invalid notification"))
	}
	n := &entities.Notification{
		UserID:    req.UserID,
		DatasetID: req.DatasetID,
		Message:   req.Message,
		IsRead:    false,
		CreatedAt: time.Now().UTC(),
	}
	if err := a.notifRepo.Create(ctx, n); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeNotificationCreateResponse, err)
	}
	resp := contracts.NotificationCreateResponse{ID: n.ID}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeNotificationCreateResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) handleNotificationMark(ctx context.Context, msg contracts.Envelope) error {
	var req contracts.NotificationMarkRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil || req.NotificationID == 0 {
		return a.publishError(ctx, msg, contracts.MsgTypeNotificationMarkResponse, fmt.Errorf("invalid notification id"))
	}
	item, err := a.notifRepo.FindByID(ctx, req.NotificationID)
	if err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeNotificationMarkResponse, err)
	}
	if item == nil {
		return a.publishError(ctx, msg, contracts.MsgTypeNotificationMarkResponse, fmt.Errorf("not found"))
	}
	item.IsRead = req.IsRead
	if err := a.notifRepo.Update(ctx, item); err != nil {
		return a.publishError(ctx, msg, contracts.MsgTypeNotificationMarkResponse, err)
	}
	resp := contracts.NotificationMarkResponse{Updated: true}
	payload, _ := json.Marshal(resp)
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          contracts.MsgTypeNotificationMarkResponse,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Payload:       payload,
	})
}

func (a *dataApp) publishError(ctx context.Context, msg contracts.Envelope, respType string, cause error) error {
	return a.bus.Publish(ctx, a.replyTo, contracts.Envelope{
		Type:          respType,
		Source:        contracts.DefaultServiceNameData,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Error: &contracts.ErrorPayload{
			Code:    contracts.MessageErrorCodeProcessError,
			Message: cause.Error(),
		},
	})
}
