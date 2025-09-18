//go:build integration
// +build integration

package postqbuild_test

import (
	"context"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	mongorepo "ppo/internal/dataaccess/repositories/mongo"
	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
	"ppo/internal/repositories"
	"testing"
	"time"
)

//var dbPool *pgxpool.Pool
//
//func TestMain(m *testing.M) {
//	p, err := dockertest.NewPool("")
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "docker pool error: %v\n", err)
//		os.Exit(1)
//	}
//	resource, err := p.RunWithOptions(&dockertest.RunOptions{
//		Repository: "postqbuild",
//		Tag:        "15-alpine",
//		Env: []string{
//			"postqbuild_USER=postqbuild",
//			"postqbuild_PASSWORD=secret",
//			"postqbuild_DB=testdb",
//		},
//	}, func(cfg *docker.HostConfig) {
//		cfg.AutoRemove = true
//		cfg.RestartPolicy = docker.RestartPolicy{Name: "no"}
//	})
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "docker run error: %v\n", err)
//		os.Exit(1)
//	}
//	defer p.Purge(resource)
//
//	var pool *pgxpool.Pool
//	if err := p.Retry(func() error {
//		dsn := fmt.Sprintf(
//			"postqbuild://postqbuild:secret@localhost:%s/testdb?sslmode=disable",
//			resource.GetPort("5432/tcp"),
//		)
//		pool, err = pgxpool.New(context.Background(), dsn)
//		if err != nil {
//			return err
//		}
//		return pool.Ping(context.Background())
//	}); err != nil {
//		fmt.Fprintf(os.Stderr, "could not connect to postqbuild: %v\n", err)
//		os.Exit(1)
//	}
//
//	abs, err := filepath.Abs(filepath.Join("..", "..", "..", "migrations"))
//	if err != nil {
//		panic(err)
//	}
//	migrationsURL := "file://" + filepath.ToSlash(abs)
//	fmt.Println("Using migrationsURL:", migrationsURL)
//
//	pgURL := fmt.Sprintf(
//		"postqbuild://postqbuild:secret@localhost:%s/testdb?sslmode=disable",
//		resource.GetPort("5432/tcp"),
//	)
//
//	migrator, err := migrate.New(migrationsURL, pgURL)
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "migrate.New error: %v\n", err)
//		os.Exit(1)
//	}
//
//	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
//		fmt.Fprintf(os.Stderr, "migrate up error: %v\n", err)
//		os.Exit(1)
//	}
//
//	dbPool = pool
//
//	code := m.Run()
//
//	dbPool.Close()
//	os.Exit(code)
//}

//func TestMetadataRepo_CRUD(t *testing.T) {
//	ctx := context.Background()
//	logger := zap.NewNop()
//	metaRepo := postqbuild.NewMetadataRepo(dbPool, logger)
//	dsRepo := postqbuild.NewDatasetRepo(dbPool, logger)
//	verRepo := postqbuild.NewVersionRepo(dbPool, logger)
//
//	var userID uint64
//	err := dbPool.QueryRow(ctx,
//		`INSERT INTO users(username,email,password,registration_date,country,role)
//		 VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
//		"muser", "muser@example.com", "pass", time.Now(), "NL", entities.RoleUser,
//	).Scan(&userID)
//	assert.NoError(t, err)
//
//	var categoryID uint64
//	err = dbPool.QueryRow(ctx,
//		`INSERT INTO categories(name,description) VALUES($1,$2) RETURNING id`,
//		"MetaCat", "meta desc",
//	).Scan(&categoryID)
//	assert.NoError(t, err)
//
//	dsEnt, err := entities.NewDataset("MDataset", "desc", userID, categoryID, false, time.Now())
//	assert.NoError(t, err)
//	err = dsRepo.Create(ctx, dsEnt)
//	assert.NoError(t, err)
//	assert.NotZero(t, dsEnt.ID)
//
//	verEnt, err := entities.NewDatasetVersion("v1.0", "/tmp/f.mp4", "initial", dsEnt.ID, time.Now())
//	assert.NoError(t, err)
//	err = verRepo.Create(ctx, verEnt)
//	assert.NoError(t, err)
//	assert.NotZero(t, verEnt.ID)
//
//	mdEnt, err := entities.NewMetadata("mp4", "tag1,tag2", 12345, verEnt.ID)
//	assert.NoError(t, err)
//	assert.Zero(t, mdEnt.ID)
//
//	err = metaRepo.Create(ctx, mdEnt)
//	assert.NoError(t, err)
//	assert.NotZero(t, mdEnt.ID)
//
//	got, err := metaRepo.FindByID(ctx, mdEnt.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, mdEnt.Format, got.Format)
//	assert.Equal(t, mdEnt.Size, got.Size)
//	assert.Equal(t, mdEnt.DatasetVersionID, got.DatasetVersionID)
//
//	list, err := metaRepo.FindByDatasetID(ctx, verEnt.ID)
//	assert.NoError(t, err)
//	assert.Len(t, list, 1)
//	assert.Equal(t, mdEnt.ID, list[0].ID)
//
//	got.Tags = "tagA,tagB"
//	got.Size = 54321
//	err = metaRepo.Update(ctx, got)
//	assert.NoError(t, err)
//
//	updated, err := metaRepo.FindByID(ctx, mdEnt.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, "tagA,tagB", updated.Tags)
//	assert.Equal(t, uint64(54321), updated.Size)
//
//	err = metaRepo.Delete(ctx, mdEnt.ID)
//	assert.NoError(t, err)
//
//	missing, err := metaRepo.FindByID(ctx, mdEnt.ID)
//	assert.ErrorIs(t, err, repositories.ErrMetadataNotFound)
//	assert.Nil(t, missing)
//}

type metadataEnv struct {
	userRepo repositories.UserRepository
	catRepo  repositories.CategoryRepository
	dsRepo   repositories.DatasetRepository
	verRepo  repositories.DatasetVersionRepository
	mdRepo   repositories.MetadataRepository
}

func runMetadataRepoTests(t *testing.T, makeEnv func() metadataEnv) {
	ctx := context.Background()
	env := makeEnv()

	// 1) Создаём пользователя
	u := &entities.User{
		Username: "u1",
		Email:    "u1@example.com",
		Password: "hash",
		Country:  "NL",
		Role:     entities.RoleUser,
	}
	assert.NoError(t, env.userRepo.Create(ctx, u))
	assert.NotZero(t, u.ID)

	// 2) Создаём категорию
	c := &entities.Category{Name: "MetaCat", Description: "meta desc"}
	assert.NoError(t, env.catRepo.Create(ctx, c))
	assert.NotZero(t, c.ID)

	// 3) Создаём датасет
	ds, err := entities.NewDataset("DS", "desc", u.ID, c.ID, false, time.Now().UTC())
	assert.NoError(t, err)
	assert.Zero(t, ds.ID)
	assert.NoError(t, env.dsRepo.Create(ctx, ds))
	assert.NotZero(t, ds.ID)

	// 4) Создаём версию датасета
	ver, err := entities.NewDatasetVersion("v1.0", "/tmp/f.mp4", "initial", ds.ID, time.Now().UTC())
	assert.NoError(t, err)
	assert.Zero(t, ver.ID)
	assert.NoError(t, env.verRepo.Create(ctx, ver))
	assert.NotZero(t, ver.ID)

	// 5) Create Metadata
	md, err := entities.NewMetadata("mp4", "tag1,tag2", 12345, ver.ID)
	assert.NoError(t, err)
	assert.Zero(t, md.ID)
	assert.NoError(t, env.mdRepo.Create(ctx, md))
	assert.NotZero(t, md.ID)

	// 6) FindByID
	got, err := env.mdRepo.FindByID(ctx, md.ID)
	assert.NoError(t, err)
	assert.Equal(t, md.Format, got.Format)
	assert.Equal(t, md.Size, got.Size)
	assert.Equal(t, md.DatasetVersionID, got.DatasetVersionID)

	// 7) FindByDatasetID
	list, err := env.mdRepo.FindByDatasetID(ctx, ver.ID)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, md.ID, list[0].ID)

	// 8) Update
	got.Tags = "tagA,tagB"
	got.Size = 54321
	assert.NoError(t, env.mdRepo.Update(ctx, got))

	updated, err := env.mdRepo.FindByID(ctx, md.ID)
	assert.NoError(t, err)
	assert.Equal(t, "tagA,tagB", updated.Tags)
	assert.Equal(t, uint64(54321), updated.Size)

	// 9) Delete
	assert.NoError(t, env.mdRepo.Delete(ctx, md.ID))
	missing, err := env.mdRepo.FindByID(ctx, md.ID)
	assert.ErrorIs(t, err, repositories.ErrMetadataNotFound)
	assert.Nil(t, missing)
}

func TestMetadataRepo_Implementations(t *testing.T) {
	tests := []struct {
		name    string
		makeEnv func() metadataEnv
	}{
		{
			name: "Postgres",
			makeEnv: func() metadataEnv {
				logger := zap.NewNop()
				return metadataEnv{
					userRepo: postqbuild.NewUserRepo(dbPool, logger),
					catRepo:  postqbuild.NewCategoryRepo(dbPool, logger),
					dsRepo:   postqbuild.NewDatasetRepo(dbPool, logger),
					verRepo:  postqbuild.NewVersionRepo(dbPool, logger),
					mdRepo:   postqbuild.NewMetadataRepo(dbPool, logger),
				}
			},
		},
		{
			name: "Mongo",
			makeEnv: func() metadataEnv {
				logger := zap.NewNop()
				return metadataEnv{
					userRepo: mongorepo.NewUserMongoRepo(mongoDB, logger),
					catRepo:  mongorepo.NewCategoryMongoRepo(mongoDB, logger),
					dsRepo:   mongorepo.NewDatasetMongoRepo(mongoDB, logger),
					verRepo:  mongorepo.NewVersionMongoRepo(mongoDB, logger),
					mdRepo:   mongorepo.NewMetadataMongoRepo(mongoDB, logger),
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			for _, tbl := range []string{
				"access_requests", "notifications", "reviews",
				"subscriptions", "metadata", "dataset_versions",
				"datasets", "categories", "users",
			} {
				if _, err := dbPool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tbl)); err != nil {
					t.Fatalf("failed to truncate %s: %v", tbl, err)
				}
			}
			for _, coll := range []string{
				"access_requests", "notifications", "reviews",
				"subscriptions", "metadata", "dataset_versions",
				"datasets", "categories", "users", "counters",
			} {
				_ = mongoDB.Collection(coll).Drop(ctx)
			}

			runMetadataRepoTests(t, tc.makeEnv)
		})
	}
}
