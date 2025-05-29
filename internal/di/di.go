package di

import (
	"context"
	"ppo/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"ppo/internal/config"
	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/delivery/cli"
	//httpdelivery "ppo/internal/delivery/http"
	"ppo/internal/services"
)

type App struct {
	Config *config.Config
	DB     *pgxpool.Pool
	//HTTPHandler httpdelivery.Router
	RootCommand *cobra.Command
}

func Build(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	dbPool, err := postqbuild.NewPool(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	s3, err := storage.NewS3Storage(cfg.Storage)
	if err != nil {
		return nil, err
	}

	catRepo := postqbuild.NewCategoryRepo(dbPool)
	dsRepo := postqbuild.NewDatasetRepo(dbPool)
	verRepo := postqbuild.NewVersionRepo(dbPool)
	mdRepo := postqbuild.NewMetadataRepo(dbPool)
	notifRepo := postqbuild.NewNotificationRepo(dbPool)
	subRepo := postqbuild.NewSubscriptionRepo(dbPool)
	revRepo := postqbuild.NewReviewRepo(dbPool)
	userRepo := postqbuild.NewUserRepo(dbPool)

	catSvc := services.NewCategoryService(catRepo)
	dsSvc := services.NewDatasetService(dsRepo, verRepo, mdRepo, s3)
	notifSvc := services.NewNotificationService(notifRepo, subRepo)
	revSvc := services.NewReviewService(revRepo)
	userSvc := services.NewUserService(userRepo)
	subSvc := services.NewSubscriptionService(subRepo)

	//router := httpdelivery.NewRouter(
	//	catSvc,
	//	dsSvc,
	//	notifSvc,
	//	revSvc,
	//	userSvc,
	//)

	rootCmd := cli.NewRootCommand(
		catSvc,
		dsSvc,
		notifSvc,
		revSvc,
		userSvc,
		subSvc,
	)

	return &App{
		Config: cfg,
		DB:     dbPool,
		//HTTPHandler: router,
		RootCommand: rootCmd,
	}, nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.DB.Close()
	return nil
}
