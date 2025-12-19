package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ppo/internal/config"
	"ppo/internal/contracts"
	"ppo/internal/entities"
	"ppo/internal/logger"
	"ppo/internal/messaging"
	"ppo/internal/messaging/rabbit"
	"ppo/internal/metrics"
)

type coreApp struct {
	cfg            *config.Config
	bus            messaging.Bus
	logger         *zap.Logger
	logClose       func()
	dataReplyCh    sync.Map
	dataReplyQueue string
	toDataQueue    string
	toGatewayQueue string
	fromGateway    string
	processed      sync.Map
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

	bus, err := rabbit.New(cfg.Broker, zapLogger)
	if err != nil {
		log.Fatalf("init broker: %v", err)
	}
	defer bus.Close()

	app := &coreApp{
		cfg:            cfg,
		bus:            bus,
		logger:         zapLogger,
		logClose:       closeLog,
		dataReplyQueue: cfg.Broker.DataToCoreQueue,
		toDataQueue:    cfg.Broker.CoreToDataQueue,
		toGatewayQueue: cfg.Broker.CoreToGatewayQueue,
		fromGateway:    cfg.Broker.GatewayToCoreQueue,
	}

	if err := app.startDataReplies(ctx); err != nil {
		log.Fatalf("data reply listener: %v", err)
	}
	if err := bus.Consume(ctx, app.fromGateway, app.handleGatewayMessage); err != nil {
		log.Fatalf("gateway consumer: %v", err)
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

	zapLogger.Info("core service started", zap.String("addr", server.Addr))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("core http server error: %v", err)
	}
}

func (a *coreApp) router() http.Handler {
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

func (a *coreApp) handleGatewayMessage(ctx context.Context, msg contracts.Envelope) error {
	if msg.ID != "" {
		if _, exists := a.processed.Load(msg.ID); exists {
			if a.logger != nil {
				a.logger.Debug("duplicate message ignored", zap.String("id", msg.ID))
			}
			return nil
		}
	}
	switch msg.Type {
	case contracts.MsgTypeHealthRequest:
		resp := contracts.HealthResponse{
			Service:   contracts.DefaultServiceNameCore,
			Status:    "ok",
			Timestamp: time.Now(),
		}
		payload, _ := json.Marshal(resp)
		return a.bus.Publish(ctx, a.toGatewayQueue, contracts.Envelope{
			Type:          contracts.MsgTypeHealthResponse,
			Source:        contracts.DefaultServiceNameCore,
			CorrelationID: msg.CorrelationID,
			OccurredAt:    time.Now(),
			Payload:       payload,
		})
	case contracts.MsgTypeCategoryListRequest,
		contracts.MsgTypeDatasetListRequest,
		contracts.MsgTypeDatasetGetRequest,
		contracts.MsgTypeVersionsListRequest,
		contracts.MsgTypeVersionGetRequest,
		contracts.MsgTypeUserGetRequest,
		contracts.MsgTypeNotificationsRequest,
		contracts.MsgTypeAccessPendingRequest,
		contracts.MsgTypeAccessFindRequest,
		contracts.MsgTypeAccessUpdateRequest,
		contracts.MsgTypeNotificationCreateRequest,
		contracts.MsgTypeNotificationMarkRequest,
		contracts.MsgTypeDatasetCreateRequest,
		contracts.MsgTypeDatasetUpdateRequest,
		contracts.MsgTypeDatasetDeleteRequest,
		contracts.MsgTypeVersionAddRequest,
		contracts.MsgTypeVersionDeleteRequest,
		contracts.MsgTypeReviewCreateRequest,
		contracts.MsgTypeReviewUpdateRequest,
		contracts.MsgTypeReviewDeleteRequest,
		contracts.MsgTypeSubscriptionCreateRequest,
		contracts.MsgTypeSubscriptionDeleteRequest,
		contracts.MsgTypeCategoryCreateRequest,
		contracts.MsgTypeCategoryUpdateRequest,
		contracts.MsgTypeCategoryDeleteRequest,
		contracts.MsgTypeUserCreateRequest,
		contracts.MsgTypeUserUpdateRequest,
		contracts.MsgTypeUserDeleteRequest:
		if err := a.enforceAccess(ctx, msg); err != nil {
			return a.sendError(ctx, msg, contracts.MessageErrorCodeProcessError, err.Error())
		}
		resp, err := a.forwardToData(ctx, msg)
		if err != nil {
			return a.sendError(ctx, msg, contracts.MessageErrorCodeProcessError, err.Error())
		}
		if msg.ID != "" {
			a.processed.Store(msg.ID, struct{}{})
		}
		return a.bus.Publish(ctx, a.toGatewayQueue, resp)
	default:
		return a.sendError(ctx, msg, contracts.MessageErrorCodeUnknownType, "unsupported message type")
	}
}

func (a *coreApp) forwardToData(ctx context.Context, msg contracts.Envelope) (contracts.Envelope, error) {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.Broker.MessageProcessTimeout)
	defer cancel()

	waitCh := make(chan contracts.Envelope, 1)
	corrID := msg.CorrelationID
	if corrID == "" {
		corrID = uuid.NewString()
	}
	msg.CorrelationID = corrID

	a.dataReplyCh.Store(corrID, waitCh)
	defer a.dataReplyCh.Delete(corrID)

	if err := a.bus.Publish(ctx, a.toDataQueue, msg); err != nil {
		return contracts.Envelope{}, err
	}

	select {
	case resp := <-waitCh:
		resp.CorrelationID = msg.CorrelationID
		return resp, nil
	case <-ctx.Done():
		return contracts.Envelope{}, ctx.Err()
	}
}

func (a *coreApp) startDataReplies(ctx context.Context) error {
	return a.bus.Consume(ctx, a.dataReplyQueue, func(ctx context.Context, msg contracts.Envelope) error {
		if msg.CorrelationID == "" {
			return nil
		}
		if ch, ok := a.dataReplyCh.Load(msg.CorrelationID); ok {
			select {
			case ch.(chan contracts.Envelope) <- msg:
			default:
			}
		}
		return nil
	})
}

func (a *coreApp) sendError(ctx context.Context, msg contracts.Envelope, code, text string) error {
	resp := contracts.Envelope{
		Type:          msg.Type,
		Source:        contracts.DefaultServiceNameCore,
		CorrelationID: msg.CorrelationID,
		TraceID:       msg.TraceID,
		OccurredAt:    time.Now(),
		Error: &contracts.ErrorPayload{
			Code:    code,
			Message: text,
		},
	}
	return a.bus.Publish(ctx, a.toGatewayQueue, resp)
}

func requireActor(msg contracts.Envelope) error {
	if msg.ActorID == 0 {
		return fmt.Errorf("actor required")
	}
	if strings.ToLower(strings.TrimSpace(msg.ActorRole)) == "guest" {
		return fmt.Errorf("insufficient role")
	}
	return nil
}

func (a *coreApp) enforceAccess(ctx context.Context, msg contracts.Envelope) error {
	if err := requireActor(msg); err != nil {
		return err
	}
	role := strings.ToLower(strings.TrimSpace(msg.ActorRole))
	switch msg.Type {
	case contracts.MsgTypeCategoryCreateRequest,
		contracts.MsgTypeCategoryUpdateRequest,
		contracts.MsgTypeCategoryDeleteRequest,
		contracts.MsgTypeUserCreateRequest,
		contracts.MsgTypeUserUpdateRequest,
		contracts.MsgTypeUserDeleteRequest:
		if role != entities.RoleAdmin {
			return fmt.Errorf("admin role required")
		}
	case contracts.MsgTypeDatasetUpdateRequest,
		contracts.MsgTypeDatasetDeleteRequest:
		var req contracts.DatasetUpdateRequest
		if err := json.Unmarshal(msg.Payload, &req); err != nil || req.ID == 0 {
			return fmt.Errorf("invalid dataset payload")
		}
		if err := a.ensureDatasetOwner(ctx, msg, req.ID); err != nil {
			return err
		}
	case contracts.MsgTypeDatasetCreateRequest:
		return nil
	case contracts.MsgTypeVersionAddRequest:
		var req contracts.VersionAddRequest
		if err := json.Unmarshal(msg.Payload, &req); err != nil || req.DatasetID == 0 {
			return fmt.Errorf("invalid version payload")
		}
		if err := a.ensureDatasetOwner(ctx, msg, req.DatasetID); err != nil {
			return err
		}
	case contracts.MsgTypeVersionDeleteRequest:
		var req contracts.VersionDeleteRequest
		if err := json.Unmarshal(msg.Payload, &req); err != nil || req.VersionID == 0 {
			return fmt.Errorf("invalid version payload")
		}
		dsID, err := a.fetchDatasetIDByVersion(ctx, msg, req.VersionID)
		if err != nil {
			return err
		}
		if err := a.ensureDatasetOwner(ctx, msg, dsID); err != nil {
			return err
		}
	case contracts.MsgTypeSubscriptionCreateRequest,
		contracts.MsgTypeSubscriptionDeleteRequest:
		var req contracts.SubscriptionDeleteRequest
		_ = json.Unmarshal(msg.Payload, &req)
		if req.UserID != 0 && req.UserID != msg.ActorID && role != entities.RoleAdmin {
			return fmt.Errorf("actor must match user")
		}
	case contracts.MsgTypeAccessUpdateRequest:
		if role != entities.RoleAdmin {
			return fmt.Errorf("admin role required")
		}
	}
	return nil
}

func (a *coreApp) ensureDatasetOwner(ctx context.Context, msg contracts.Envelope, datasetID uint64) error {
	ds, err := a.fetchDataset(ctx, msg, datasetID)
	if err != nil {
		return err
	}
	if ds == nil {
		return fmt.Errorf("dataset not found")
	}
	if ds.OwnerID != msg.ActorID && strings.ToLower(strings.TrimSpace(msg.ActorRole)) != entities.RoleAdmin {
		return fmt.Errorf("forbidden: not owner")
	}
	return nil
}

func (a *coreApp) fetchDatasetIDByVersion(ctx context.Context, msg contracts.Envelope, versionID uint64) (uint64, error) {
	payload, _ := json.Marshal(contracts.VersionGetRequest{VersionID: versionID})
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeVersionGetRequest,
		Source:        contracts.DefaultServiceNameCore,
		CorrelationID: uuid.NewString(),
		TraceID:       msg.TraceID,
		Payload:       payload,
	}
	resp, err := a.forwardToData(ctx, env)
	if err != nil {
		return 0, err
	}
	if resp.Error != nil {
		return 0, fmt.Errorf("%s", resp.Error.Message)
	}
	var vr contracts.VersionGetResponse
	if err := json.Unmarshal(resp.Payload, &vr); err != nil {
		return 0, err
	}
	if vr.Version == nil {
		return 0, fmt.Errorf("version not found")
	}
	return vr.Version.DatasetID, nil
}

func (a *coreApp) fetchDataset(ctx context.Context, msg contracts.Envelope, datasetID uint64) (*entities.Dataset, error) {
	payload, _ := json.Marshal(contracts.DatasetGetRequest{ID: datasetID})
	env := contracts.Envelope{
		ID:            uuid.NewString(),
		Type:          contracts.MsgTypeDatasetGetRequest,
		Source:        contracts.DefaultServiceNameCore,
		CorrelationID: uuid.NewString(),
		TraceID:       msg.TraceID,
		Payload:       payload,
	}
	resp, err := a.forwardToData(ctx, env)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("%s", resp.Error.Message)
	}
	var dr contracts.DatasetGetResponse
	if err := json.Unmarshal(resp.Payload, &dr); err != nil {
		return nil, err
	}
	return dr.Dataset, nil
}
