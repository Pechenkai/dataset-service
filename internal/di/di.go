package di

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"os"
	httpdelivery "ppo/internal/delivery/http"
	"ppo/internal/logger"
	"ppo/internal/storage"

	webui "ppo/internal/delivery/web"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"ppo/internal/config"
	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/delivery/cli"
	"ppo/internal/services"
)

type App struct {
	Config      *config.Config
	DB          *pgxpool.Pool
	HTTPHandler http.Handler
	RootCommand *cobra.Command
	Logger      *zap.Logger
	WebHandler  chi.Router
	logClose    func()
}

func Build(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	zapLogger, closeLog, err := logger.NewLogger(cfg.LogCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		zapLogger = zap.NewNop()
		closeLog = func() {}
	}
	dbPool, err := postqbuild.NewPool(ctx, cfg.Database)
	if err != nil {
		closeLog()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	zapLogger.Info("database pool created")

	s3, err := storage.NewS3Storage(cfg.Storage)
	if err != nil {
		dbPool.Close()
		closeLog()
		return nil, fmt.Errorf("init storage: %w", err)
	}

	catRepo := postqbuild.NewCategoryRepo(dbPool, zapLogger)
	dsRepo := postqbuild.NewDatasetRepo(dbPool, zapLogger)
	verRepo := postqbuild.NewVersionRepo(dbPool, zapLogger)
	mdRepo := postqbuild.NewMetadataRepo(dbPool, zapLogger)
	notifRepo := postqbuild.NewNotificationRepo(dbPool, zapLogger)
	subRepo := postqbuild.NewSubscriptionRepo(dbPool, zapLogger)
	revRepo := postqbuild.NewReviewRepo(dbPool, zapLogger)
	userRepo := postqbuild.NewUserRepo(dbPool, zapLogger)
	requestRepo := postqbuild.NewAccessRequestRepo(dbPool, zapLogger)

	catSvc := services.NewCategoryService(catRepo, zapLogger)
	dsSvc := services.NewDatasetService(dsRepo, verRepo, mdRepo, s3, zapLogger)
	notifSvc := services.NewNotificationService(notifRepo, subRepo, zapLogger)
	revSvc := services.NewReviewService(revRepo, zapLogger)
	userSvc := services.NewUserService(userRepo, zapLogger)
	subSvc := services.NewSubscriptionService(subRepo, zapLogger)
	reqSvc := services.NewAccessService(dsRepo, requestRepo, zapLogger)

	router := httpdelivery.NewRouter(
		catSvc,
		dsSvc,
		notifSvc,
		revSvc,
		userSvc,
		subSvc,
		zapLogger,
	)

	webrouter := webui.NewRouter(
		catSvc,
		dsSvc,
		notifSvc,
		revSvc,
		subSvc,
		userSvc,
		reqSvc,
		zapLogger,
	)

	rootCmd := cli.NewRootCommand(
		catSvc,
		dsSvc,
		notifSvc,
		revSvc,
		userSvc,
		subSvc,
	)

	return &App{
		Config:      cfg,
		DB:          dbPool,
		HTTPHandler: router,
		RootCommand: rootCmd,
		Logger:      zapLogger,
		WebHandler:  webrouter,
		logClose:    closeLog,
	}, nil
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.DB != nil {
		a.DB.Close()
	}
	if a.logClose != nil {
		a.logClose()
	}
	if a.Logger != nil {
		return a.Logger.Sync()
	}
	return nil
}
