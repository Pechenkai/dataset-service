package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ppo/internal/config"
	"ppo/internal/contracts"
	"ppo/internal/logger"
	"ppo/internal/messaging"
	"ppo/internal/messaging/rabbit"
	"ppo/internal/metrics"
	"ppo/internal/storage"
)

type gatewayApp struct {
	cfg          *config.Config
	bus          messaging.Bus
	logger       *zap.Logger
	logClose     func()
	pending      sync.Map
	replyQueue   string
	requestQueue string
	storage      storage.Storage
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

	s3, err := storage.NewS3Storage(cfg.Storage)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	bus, err := rabbit.New(cfg.Broker, zapLogger)
	if err != nil {
		log.Fatalf("init broker: %v", err)
	}
	defer bus.Close()

	app := &gatewayApp{
		cfg:          cfg,
		bus:          bus,
		logger:       zapLogger,
		logClose:     closeLog,
		replyQueue:   cfg.Broker.CoreToGatewayQueue,
		requestQueue: cfg.Broker.GatewayToCoreQueue,
		storage:      s3,
	}

	if err := app.startReplyListener(ctx); err != nil {
		log.Fatalf("start reply listener: %v", err)
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

	zapLogger.Info("gateway service started", zap.String("addr", server.Addr))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("gateway http server error: %v", err)
	}
}

func (a *gatewayApp) router() http.Handler {
	r := chi.NewRouter()
	r.Use(metrics.Middleware)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.Handler().ServeHTTP(w, r)
	})

	r.Get("/async/categories", a.handleListCategories)
	r.Get("/async/datasets", a.handleListDatasets)
	r.Get("/async/datasets/{id}", a.handleGetDataset)
	r.Get("/async/datasets/{id}/versions", a.handleListVersions)
	r.Get("/async/versions/{id}", a.handleGetVersion)
	r.Get("/async/notifications/{userID}", a.handleListNotifications)
	r.Get("/async/access/owner/{ownerID}/pending", a.handlePendingAccess)
	r.Get("/async/access", a.handleFindAccess)
	r.Post("/async/access/{id}/status", a.handleUpdateAccessStatus)
	r.Post("/async/notifications", a.handleCreateNotification)
	r.Post("/async/notifications/{id}/read", a.handleMarkNotification)
	r.Post("/async/categories", a.handleCreateCategory)
	r.Put("/async/categories/{id}", a.handleUpdateCategory)
	r.Delete("/async/categories/{id}", a.handleDeleteCategory)
	r.Post("/async/users", a.handleCreateUser)
	r.Put("/async/users/{id}", a.handleUpdateUser)
	r.Delete("/async/users/{id}", a.handleDeleteUser)
	r.Get("/async/users/{id}", a.handleGetUser)
	r.Post("/async/datasets", a.handleCreateDataset)
	r.Put("/async/datasets/{id}", a.handleUpdateDataset)
	r.Delete("/async/datasets/{id}", a.handleDeleteDataset)
	r.Post("/async/datasets/{id}/versions", a.handleAddVersion)
	r.Delete("/async/versions/{id}", a.handleDeleteVersion)
	r.Post("/async/reviews", a.handleCreateReview)
	r.Put("/async/reviews/{id}", a.handleUpdateReview)
	r.Delete("/async/reviews/{id}", a.handleDeleteReview)
	r.Post("/async/subscriptions", a.handleCreateSubscription)
	r.Delete("/async/subscriptions", a.handleDeleteSubscription)

	return r
}

func (a *gatewayApp) handleListCategories(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(withTrace(r.Context(), r), a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	req := contracts.ListCategoriesRequest{}
	body, err := json.Marshal(req)
	if err != nil {
		http.Error(w, "encode request", http.StatusInternalServerError)
		return
	}

	corrID := uuid.NewString()
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeCategoryListRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: corrID,
		OccurredAt:    time.Now(),
		Payload:       body,
	}

	respEnv, err := a.rpc(ctx, env)
	if err != nil {
		a.logger.Error("category rpc failed", zap.Error(err))
		http.Error(w, "core unavailable", http.StatusBadGateway)
		return
	}
	if respEnv.Error != nil {
		http.Error(w, respEnv.Error.Message, http.StatusBadGateway)
		return
	}

	var resp contracts.ListCategoriesResponse
	if err := json.Unmarshal(respEnv.Payload, &resp); err != nil {
		a.logger.Error("decode categories response", zap.Error(err))
		http.Error(w, "decode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		a.logger.Error("encode categories response", zap.Error(err))
	}
}

func (a *gatewayApp) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var req contracts.CategoryCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeCategoryCreateRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.CategoryCreateResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req contracts.CategoryUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	req.ID = id
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeCategoryUpdateRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.CategoryUpdateResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	req := contracts.CategoryDeleteRequest{ID: id}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeCategoryDeleteRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.CategoryDeleteResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req contracts.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeUserCreateRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.UserCreateResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req contracts.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	req.ID = id
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeUserUpdateRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.UserUpdateResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	req := contracts.UserDeleteRequest{ID: id}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeUserDeleteRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.UserDeleteResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleGetUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(withTrace(r.Context(), r), a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	req := contracts.UserGetRequest{ID: id}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeUserGetRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.UserGetResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleListDatasets(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(withTrace(r.Context(), r), a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	onlyPublic := r.URL.Query().Get("only_public") == "true"
	var ownerID *uint64
	if v := strings.TrimSpace(r.URL.Query().Get("owner_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			http.Error(w, "invalid owner_id", http.StatusBadRequest)
			return
		}
		ownerID = &id
	}

	req := contracts.DatasetListRequest{OnlyPublic: onlyPublic, OwnerID: ownerID}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeDatasetListRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}

	var resp contracts.DatasetListResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleGetDataset(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(withTrace(r.Context(), r), a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	req := contracts.DatasetGetRequest{ID: id}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeDatasetGetRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.DatasetGetResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleListVersions(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(withTrace(r.Context(), r), a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	req := contracts.VersionsListRequest{DatasetID: id}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeVersionsListRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.VersionsListResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleGetVersion(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(withTrace(r.Context(), r), a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	req := contracts.VersionGetRequest{VersionID: id}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeVersionGetRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.VersionGetResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUintParam(chi.URLParam(r, "userID"), "userID")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(withTrace(r.Context(), r), a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	req := contracts.NotificationsRequest{UserID: userID}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeNotificationsRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.NotificationsResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handlePendingAccess(w http.ResponseWriter, r *http.Request) {
	ownerID, err := parseUintParam(chi.URLParam(r, "ownerID"), "ownerID")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(withTrace(r.Context(), r), a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	req := contracts.AccessPendingRequest{OwnerID: ownerID}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeAccessPendingRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.AccessPendingResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleFindAccess(w http.ResponseWriter, r *http.Request) {
	datasetID, err := parseUintParam(r.URL.Query().Get("dataset_id"), "dataset_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, err := parseUintParam(r.URL.Query().Get("user_id"), "user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(withTrace(r.Context(), r), a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	req := contracts.AccessFindRequest{DatasetID: datasetID, UserID: userID}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeAccessFindRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.AccessFindResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleUpdateAccessStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var req contracts.AccessUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	req.RequestID = id
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeAccessUpdateRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.AccessUpdateResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleCreateNotification(w http.ResponseWriter, r *http.Request) {
	var req contracts.NotificationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeNotificationCreateRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.NotificationCreateResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleMarkNotification(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	req := contracts.NotificationMarkRequest{
		NotificationID: id,
		IsRead:         true,
	}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeNotificationMarkRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.NotificationMarkResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleCreateDataset(w http.ResponseWriter, r *http.Request) {
	var req contracts.DatasetCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctxBase, uid, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	req.OwnerID = uid

	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeDatasetCreateRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.DatasetCreateResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleUpdateDataset(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var req contracts.DatasetUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	req.ID = id
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeDatasetUpdateRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.DatasetUpdateResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleDeleteDataset(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	req := contracts.DatasetDeleteRequest{ID: id}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeDatasetDeleteRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.DatasetDeleteResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleAddVersion(w http.ResponseWriter, r *http.Request) {
	datasetID, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	req, err := a.parseVersionAddRequest(ctx, r, datasetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeVersionAddRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.VersionAddResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleDeleteVersion(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	req := contracts.VersionDeleteRequest{VersionID: id}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeVersionDeleteRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.VersionDeleteResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleCreateReview(w http.ResponseWriter, r *http.Request) {
	var req contracts.ReviewCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctxBase, uid, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	req.UserID = uid

	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeReviewCreateRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.ReviewCreateResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleUpdateReview(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var req contracts.ReviewUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	req.ID = id
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeReviewUpdateRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.ReviewUpdateResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleDeleteReview(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"), "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxBase, _, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	req := contracts.ReviewDeleteRequest{ID: id}
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeReviewDeleteRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.ReviewDeleteResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req contracts.SubscriptionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctxBase, uid, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	req.UserID = uid
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeSubscriptionCreateRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.SubscriptionCreateResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) handleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	dsID, err := parseUintParam(r.URL.Query().Get("dataset_id"), "dataset_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxBase, uid, _, err := a.actorContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// enforce actor user
	req := contracts.SubscriptionDeleteRequest{UserID: uid, DatasetID: dsID}
	ctx, cancel := context.WithTimeout(ctxBase, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()
	body, _ := json.Marshal(req)
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeSubscriptionDeleteRequest,
		Source:        contracts.DefaultServiceNameGateway,
		CorrelationID: uuid.NewString(),
		OccurredAt:    time.Now(),
		Payload:       body,
	}
	var resp contracts.SubscriptionDeleteResponse
	if err := a.rpcAndDecode(ctx, env, &resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, resp)
}

func (a *gatewayApp) rpc(ctx context.Context, msg contracts.Envelope) (contracts.Envelope, error) {
	if msg.TraceID == "" {
		if trace := ctx.Value(traceKey{}); trace != nil {
			if s, ok := trace.(string); ok {
				msg.TraceID = s
			}
		}
	}
	if actor := ctx.Value(actorKey{}); actor != nil {
		if id, ok := actor.(uint64); ok {
			msg.ActorID = id
		}
	}
	if role := ctx.Value(roleKey{}); role != nil {
		if s, ok := role.(string); ok {
			msg.ActorRole = s
		}
	}
	waitCh := make(chan contracts.Envelope, 1)
	if msg.CorrelationID == "" {
		msg.CorrelationID = uuid.NewString()
	}
	a.pending.Store(msg.CorrelationID, waitCh)
	defer a.pending.Delete(msg.CorrelationID)

	if err := a.bus.Publish(ctx, a.requestQueue, msg); err != nil {
		return contracts.Envelope{}, err
	}

	select {
	case resp := <-waitCh:
		return resp, nil
	case <-ctx.Done():
		return contracts.Envelope{}, ctx.Err()
	}
}

func (a *gatewayApp) startReplyListener(ctx context.Context) error {
	return a.bus.Consume(ctx, a.replyQueue, func(ctx context.Context, msg contracts.Envelope) error {
		if msg.CorrelationID == "" {
			return nil
		}
		if ch, ok := a.pending.Load(msg.CorrelationID); ok {
			select {
			case ch.(chan contracts.Envelope) <- msg:
			default:
			}
		} else if a.logger != nil {
			a.logger.Debug("no pending requester for message",
				zap.String("corr_id", msg.CorrelationID),
				zap.String("type", msg.Type))
		}
		return nil
	})
}

func (a *gatewayApp) rpcAndDecode(ctx context.Context, env contracts.Envelope, out any) error {
	respEnv, err := a.rpc(ctx, env)
	if err != nil {
		return err
	}
	if respEnv.Error != nil {
		return fmt.Errorf("upstream error: %s", respEnv.Error.Message)
	}
	if out == nil {
		return nil
	}
	if len(respEnv.Payload) == 0 {
		return fmt.Errorf("empty payload")
	}
	return json.Unmarshal(respEnv.Payload, out)
}

func parseUintParam(value, name string) (uint64, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return 0, fmt.Errorf("%s is required", name)
	}
	id, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return id, nil
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

type traceKey struct{}
type actorKey struct{}
type roleKey struct{}

func withTrace(ctx context.Context, r *http.Request) context.Context {
	if t := r.Header.Get("X-Request-ID"); t != "" {
		return context.WithValue(ctx, traceKey{}, t)
	}
	return context.WithValue(ctx, traceKey{}, uuid.NewString())
}

func userFromHeader(r *http.Request) (uint64, error) {
	idStr := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if idStr == "" {
		return 0, fmt.Errorf("missing X-User-ID")
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid X-User-ID")
	}
	return id, nil
}

func withActor(ctx context.Context, id uint64, role string) context.Context {
	ctx = context.WithValue(ctx, actorKey{}, id)
	ctx = context.WithValue(ctx, roleKey{}, role)
	return ctx
}

func roleFromHeader(r *http.Request) string {
	role := strings.ToLower(strings.TrimSpace(r.Header.Get("X-User-Role")))
	if role == "" {
		return "user"
	}
	return role
}

func (a *gatewayApp) actorContext(r *http.Request) (context.Context, uint64, string, error) {
	uid, err := userFromHeader(r)
	if err != nil {
		return nil, 0, "", err
	}
	role := roleFromHeader(r)
	ctx := withActor(withTrace(r.Context(), r), uid, role)
	return ctx, uid, role, nil
}

func (a *gatewayApp) parseVersionAddRequest(ctx context.Context, r *http.Request, datasetID uint64) (contracts.VersionAddRequest, error) {
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(50 << 20); err != nil {
			return contracts.VersionAddRequest{}, fmt.Errorf("parse multipart: %w", err)
		}
		file, fh, err := r.FormFile("file")
		if err != nil {
			return contracts.VersionAddRequest{}, fmt.Errorf("file is required")
		}
		defer file.Close()

		number := strings.TrimSpace(r.FormValue("number"))
		if number == "" {
			number = "v0.1"
		}
		changeLog := r.FormValue("change_log")
		objectKey := buildObjectKey(datasetID, number, fh.Filename)

		if a.storage == nil {
			return contracts.VersionAddRequest{}, fmt.Errorf("storage is not configured")
		}
		if _, err := a.storage.Upload(ctx, objectKey, file, fh.Size); err != nil {
			return contracts.VersionAddRequest{}, fmt.Errorf("upload file: %w", err)
		}

		return contracts.VersionAddRequest{
			DatasetID: datasetID,
			Number:    number,
			ChangeLog: changeLog,
			Filepath:  objectKey,
		}, nil
	}

	var req contracts.VersionAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return contracts.VersionAddRequest{}, fmt.Errorf("invalid body")
	}
	req.DatasetID = datasetID
	if strings.TrimSpace(req.Filepath) == "" {
		return contracts.VersionAddRequest{}, fmt.Errorf("filepath is required")
	}
	return req, nil
}

func buildObjectKey(datasetID uint64, version string, filename string) string {
	name := filepath.Base(filename)
	if name == "." || name == "" {
		name = "file.bin"
	}
	safeVersion := strings.ReplaceAll(strings.TrimSpace(version), "/", "_")
	return fmt.Sprintf("datasets/%d/%s/%s", datasetID, safeVersion, name)
}
