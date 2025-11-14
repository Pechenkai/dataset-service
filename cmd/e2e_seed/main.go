package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"

	"ppo/internal/config"
	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/services"
	"ppo/internal/storage"
	"ppo/internal/tests/testdata"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := zap.NewNop()

	dbPool, err := postqbuild.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer dbPool.Close()

	if err := ensureBucket(ctx, cfg.Storage); err != nil {
		return fmt.Errorf("ensure bucket: %w", err)
	}

	s3, err := storage.NewS3Storage(cfg.Storage)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}

	catRepo := postqbuild.NewCategoryRepo(dbPool, logger)
	dsRepo := postqbuild.NewDatasetRepo(dbPool, logger)
	verRepo := postqbuild.NewVersionRepo(dbPool, logger)
	mdRepo := postqbuild.NewMetadataRepo(dbPool, logger)
	userRepo := postqbuild.NewUserRepo(dbPool, logger)

	catSvc := services.NewCategoryService(catRepo, logger)
	dsSvc := services.NewDatasetService(dsRepo, verRepo, mdRepo, s3, logger)

	fabric := testdata.NewFabric()
	seed := time.Now().UnixNano()

	user := fabric.RegularUser()
	user.Email = fmt.Sprintf("capture-user-%d@example.com", seed)
	user.Username = fmt.Sprintf("capture_user_%d", seed)
	if err := userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	catCmd := fabric.CreateCategoryCommand()
	catCmd.Name = fmt.Sprintf("Capture Category %d", seed)
	categoryID, err := catSvc.CreateCategory(ctx, catCmd)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}

	category, err := catSvc.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("fetch category: %w", err)
	}

	dsCmd := fabric.CreateDatasetCommand(category, user)
	dsCmd.ActorID = user.ID
	dsCmd.Name = fmt.Sprintf("Capture Dataset %d", seed)
	dsCmd.FileName = fmt.Sprintf("capture-%d.bin", seed)
	payload := fmt.Sprintf("capture payload %d", seed)
	reader := strings.NewReader(payload)

	datasetID, err := dsSvc.CreateDataset(ctx, dsCmd, reader, int64(len(payload)))
	if err != nil {
		return fmt.Errorf("create dataset: %w", err)
	}

	fmt.Printf("USER_ID=%d\n", user.ID)
	fmt.Printf("DATASET_ID=%d\n", datasetID)
	fmt.Printf("CATEGORY_ID=%d\n", categoryID)
	return nil
}

func ensureBucket(ctx context.Context, cfg config.Storage) error {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
		Region: cfg.Region,
	})
	if err != nil {
		return err
	}

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region})
}
