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
}

func Build(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	zapLogger, closeLog, err := logger.NewLogger(cfg.LogCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %w\n", err)
		zapLogger = zap.NewNop()
		closeLog = func() {}
	}

	_ = closeLog

	dbPool, err := postqbuild.NewPool(ctx, cfg.Database)
	if err != nil {
		zapLogger.Fatal("failed to connect to database", zap.Error(err))
	}

	zapLogger.Info("database pool created")

	s3, err := storage.NewS3Storage(cfg.Storage)
	if err != nil {
		return nil, err
	}

	catRepo := postqbuild.NewCategoryRepo(dbPool, zapLogger)
	dsRepo := postqbuild.NewDatasetRepo(dbPool, zapLogger)
	verRepo := postqbuild.NewVersionRepo(dbPool, zapLogger)
	mdRepo := postqbuild.NewMetadataRepo(dbPool, zapLogger)
	notifRepo := postqbuild.NewNotificationRepo(dbPool, zapLogger)
	subRepo := postqbuild.NewSubscriptionRepo(dbPool, zapLogger)
	revRepo := postqbuild.NewReviewRepo(dbPool, zapLogger)
	userRepo := postqbuild.NewUserRepo(dbPool, zapLogger)

	catSvc := services.NewCategoryService(catRepo, zapLogger)
	dsSvc := services.NewDatasetService(dsRepo, verRepo, mdRepo, s3, zapLogger)
	notifSvc := services.NewNotificationService(notifRepo, subRepo, zapLogger)
	revSvc := services.NewReviewService(revRepo, zapLogger)
	userSvc := services.NewUserService(userRepo, zapLogger)
	subSvc := services.NewSubscriptionService(subRepo, zapLogger)

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
		WebHandler:  webrouter,
	}, nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.DB.Close()
	return a.Logger.Sync()
}
