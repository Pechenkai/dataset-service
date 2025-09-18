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

//func TestDatasetRepo_CRUD(t *testing.T) {
//	ctx := context.Background()
//	logger := zap.NewNop()
//	repo := postqbuild.NewDatasetRepo(dbPool, logger)
//
//	var userID uint64
//	err := dbPool.QueryRow(ctx,
//		`INSERT INTO users(username,email,password,registration_date,country,role)
//		    VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
//		"u1", "u1@example.com", "hash", time.Now(), "NL", entities.RoleUser,
//	).Scan(&userID)
//	assert.NoError(t, err)
//
//	var categoryID uint64
//	err = dbPool.QueryRow(ctx,
//		`INSERT INTO categories(name,description) VALUES($1,$2) RETURNING id`,
//		"Cat1", "desc").Scan(&categoryID)
//	assert.NoError(t, err)
//
//	ds, err := entities.NewDataset("DS1", "Dataset 1", userID, categoryID, true, time.Now())
//	assert.NoError(t, err)
//	assert.Zero(t, ds.ID)
//
//	err = repo.Create(ctx, ds)
//	assert.NoError(t, err)
//	assert.NotZero(t, ds.ID)
//
//	fetched, err := repo.FindByID(ctx, ds.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, ds.Name, fetched.Name)
//	assert.Equal(t, ds.OwnerID, fetched.OwnerID)
//
//	byUser, err := repo.FindByUserID(ctx, userID)
//	assert.NoError(t, err)
//	assert.Len(t, byUser, 1)
//	assert.Equal(t, ds.ID, byUser[0].ID)
//
//	all, err := repo.FindAll(ctx)
//	assert.NoError(t, err)
//	assert.GreaterOrEqual(t, len(all), 1)
//
//	fetched.Name = "DS1-renamed"
//	fetched.Description = "updated"
//	err = repo.Update(ctx, fetched)
//	assert.NoError(t, err)
//
//	updated, err := repo.FindByID(ctx, ds.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, "DS1-renamed", updated.Name)
//	assert.Equal(t, "updated", updated.Description)
//
//	err = repo.Delete(ctx, ds.ID)
//	assert.NoError(t, err)
//
//	missing, err := repo.FindByID(ctx, ds.ID)
//	assert.ErrorIs(t, err, repositories.ErrDatasetNotFound)
//	assert.Nil(t, missing)
//}

type dsEnv struct {
	userRepo repositories.UserRepository
	catRepo  repositories.CategoryRepository
	dsRepo   repositories.DatasetRepository
}

func runDatasetRepoTests(t *testing.T, makeEnv func() dsEnv) {
	ctx := context.Background()
	env := makeEnv()

	u := &entities.User{
		Username: "u1", Email: "u1@example.com",
		Password: "hash", Country: "NL", Role: entities.RoleUser,
	}
	assert.NoError(t, env.userRepo.Create(ctx, u))
	assert.NotZero(t, u.ID)

	c := &entities.Category{Name: "Cat1", Description: "desc"}
	assert.NoError(t, env.catRepo.Create(ctx, c))
	assert.NotZero(t, c.ID)

	// 2) Create Dataset
	ds, err := entities.NewDataset("DS1", "Dataset 1", u.ID, c.ID, true, time.Now().UTC())
	assert.NoError(t, err)
	assert.Zero(t, ds.ID)

	assert.NoError(t, env.dsRepo.Create(ctx, ds))
	assert.NotZero(t, ds.ID)

	// 3) FindByID
	fetched, err := env.dsRepo.FindByID(ctx, ds.ID)
	assert.NoError(t, err)
	assert.Equal(t, ds.Name, fetched.Name)
	assert.Equal(t, ds.OwnerID, fetched.OwnerID)

	// 4) FindByUserID
	byUser, err := env.dsRepo.FindByUserID(ctx, u.ID)
	assert.NoError(t, err)
	assert.Len(t, byUser, 1)
	assert.Equal(t, ds.ID, byUser[0].ID)

	// 5) FindAll
	all, err := env.dsRepo.FindAll(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(all), 1)

	// 6) Update
	fetched.Name = "DS1-renamed"
	fetched.Description = "updated"
	assert.NoError(t, env.dsRepo.Update(ctx, fetched))

	updated, err := env.dsRepo.FindByID(ctx, ds.ID)
	assert.NoError(t, err)
	assert.Equal(t, "DS1-renamed", updated.Name)
	assert.Equal(t, "updated", updated.Description)

	// 7) Delete
	assert.NoError(t, env.dsRepo.Delete(ctx, ds.ID))
	missing, err := env.dsRepo.FindByID(ctx, ds.ID)
	assert.ErrorIs(t, err, repositories.ErrDatasetNotFound)
	assert.Nil(t, missing)
}

func TestDatasetRepo_Implementations(t *testing.T) {
	tests := []struct {
		name    string
		makeEnv func() dsEnv
	}{
		{
			name: "Postgres",
			makeEnv: func() dsEnv {
				logger := zap.NewNop()
				return dsEnv{
					userRepo: postqbuild.NewUserRepo(dbPool, logger),
					catRepo:  postqbuild.NewCategoryRepo(dbPool, logger),
					dsRepo:   postqbuild.NewDatasetRepo(dbPool, logger),
				}
			},
		},
		{
			name: "Mongo",
			makeEnv: func() dsEnv {
				logger := zap.NewNop()
				return dsEnv{
					userRepo: mongorepo.NewUserMongoRepo(mongoDB, logger),
					catRepo:  mongorepo.NewCategoryMongoRepo(mongoDB, logger),
					dsRepo:   mongorepo.NewDatasetMongoRepo(mongoDB, logger),
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

			runDatasetRepoTests(t, tc.makeEnv)
		})
	}
}
