package di

import (
	"context"
	"errors"
	"fmt"
	chi "github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"net/http"
	"os"
	"ppo/internal/config"
	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/delivery/cli"
	cliapi "ppo/internal/delivery/cli/api"
	httpdelivery "ppo/internal/delivery/http"
	"ppo/internal/entities"
	"ppo/internal/integrations/catfacts"
	"ppo/internal/integrations/openai"
	"ppo/internal/logger"
	"ppo/internal/services"
	"ppo/internal/storage"
	"strings"
	"time"

	webui "ppo/internal/delivery/web"
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
	catfactsClient := catfacts.NewHTTPClient(cfg.External.CatFacts, zapLogger)
	factSvc := services.NewDatasetFactService(dsRepo, catfactsClient, zapLogger)
	openaiClient := openai.NewHTTPClient(cfg.External.OpenAI, zapLogger)
	summarySvc := services.NewDatasetSummaryService(dsRepo, openaiClient, zapLogger)
	tokenTTL := cfg.Auth.AccessTokenTTL
	if tokenTTL <= 0 {
		tokenTTL = 24 * time.Hour
	}
	tokenSvc := services.NewHMACTokenService(cfg.Auth.Secret, tokenTTL)
	twofaSvc := services.NewTwoFAService(services.TwoFAConfig{
		CodeTTL:       cfg.Auth.TwoFA.CodeTTL,
		MaxAttempts:   cfg.Auth.TwoFA.MaxAttempts,
		BlockDuration: cfg.Auth.TwoFA.BlockDuration,
		Delivery:      cfg.Auth.TwoFA.Delivery,
		DebugSecret:   cfg.Auth.TwoFA.DebugSecret,
	}, nil, zapLogger)

	if err := ensureAdminUser(ctx, userSvc, cfg.Admin, zapLogger); err != nil {
		dbPool.Close()
		closeLog()
		return nil, fmt.Errorf("seed admin user: %w", err)
	}

	router := httpdelivery.NewRouter(
		catSvc,
		dsSvc,
		factSvc,
		summarySvc,
		notifSvc,
		revSvc,
		userSvc,
		subSvc,
		reqSvc,
		tokenSvc,
		tokenTTL,
		twofaSvc,
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

	tokenStore, err := cliapi.NewFileTokenStore(cfg.CLI.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("init token store: %w", err)
	}
	apiClient, err := cliapi.NewClient(cfg.CLI.APIBaseURL, tokenStore)
	if err != nil {
		return nil, fmt.Errorf("init api client: %w", err)
	}
	rootCmd := cli.NewRootCommand(apiClient)

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

func ensureAdminUser(ctx context.Context, userSvc services.UserService, creds config.AdminAccount, logger *zap.Logger) error {
	email := strings.TrimSpace(creds.Email)
	password := strings.TrimSpace(creds.Password)
	if email == "" || password == "" {
		logger.Debug("admin credentials incomplete; skipping seed", zap.String("email", creds.Email))
		return nil
	}

	_, err := userSvc.GetUserByEmail(ctx, email)
	if err == nil {
		logger.Debug("admin user already exists", zap.String("email", email))
		return nil
	}
	if !errors.Is(err, services.ErrUserNotFound) {
		return fmt.Errorf("check admin existence: %w", err)
	}

	username := strings.TrimSpace(creds.Username)
	if username == "" {
		username = "admin"
	}
	country := strings.TrimSpace(creds.Country)
	if country == "" {
		country = "RU"
	}

	if _, err := userSvc.Register(ctx, services.RegisterUserCmd{
		Username: username,
		Email:    email,
		Password: password,
		Country:  country,
		Role:     entities.RoleAdmin,
	}); err != nil {
		if errors.Is(err, services.ErrUserExists) {
			logger.Info("admin user already exists", zap.String("email", email))
			return nil
		}
		return fmt.Errorf("create admin user: %w", err)
	}

	logger.Info("default admin user seeded", zap.String("email", email))
	return nil
}
