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

//func TestDatasetVersionRepo_CRUD(t *testing.T) {
//	ctx := context.Background()
//
//	var userID uint64
//	err := dbPool.QueryRow(ctx,
//		`INSERT INTO users(username,email,password,registration_date,country,role)
//		 VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
//		"vuser", "vuser@example.com", "pass", time.Now(), "NL", entities.RoleUser,
//	).Scan(&userID)
//	assert.NoError(t, err)
//
//	var categoryID uint64
//	err = dbPool.QueryRow(ctx,
//		`INSERT INTO categories(name,description) VALUES($1,$2) RETURNING id`,
//		"VerCat", "desc",
//	).Scan(&categoryID)
//	assert.NoError(t, err)
//
//	logger := zap.NewNop()
//	dsRepo := postqbuild.NewDatasetRepo(dbPool, logger)
//	dsEnt, err := entities.NewDataset("TestDS", "desc", userID, categoryID, true, time.Now())
//	assert.NoError(t, err)
//	assert.Zero(t, dsEnt.ID)
//
//	err = dsRepo.Create(ctx, dsEnt)
//	assert.NoError(t, err)
//	assert.NotZero(t, dsEnt.ID)
//
//	verRepo := postqbuild.NewVersionRepo(dbPool, logger)
//	verEnt, err := entities.NewDatasetVersion("v1.0", "/tmp/file1", "initial upload", dsEnt.ID, time.Now())
//	assert.NoError(t, err)
//	assert.Zero(t, verEnt.ID)
//
//	err = verRepo.Create(ctx, verEnt)
//	assert.NoError(t, err)
//	assert.NotZero(t, verEnt.ID)
//
//	fetched, err := verRepo.FindByID(ctx, verEnt.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, verEnt.Number, fetched.Number)
//	assert.Equal(t, verEnt.DatasetID, fetched.DatasetID)
//
//	versions, err := verRepo.FindByDatasetID(ctx, dsEnt.ID)
//	assert.NoError(t, err)
//	assert.Len(t, versions, 1)
//	assert.Equal(t, verEnt.ID, versions[0].ID)
//
//	fetched.ChangeLog = "fixed changelog"
//	err = verRepo.Update(ctx, fetched)
//	assert.NoError(t, err)
//
//	updated, err := verRepo.FindByID(ctx, verEnt.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, "fixed changelog", updated.ChangeLog)
//
//	err = verRepo.Delete(ctx, verEnt.ID)
//	assert.NoError(t, err)
//
//	missing, err := verRepo.FindByID(ctx, verEnt.ID)
//	assert.ErrorIs(t, err, repositories.ErrVersionNotFound)
//	assert.Nil(t, missing)
//}

type versionEnv struct {
	userRepo repositories.UserRepository
	catRepo  repositories.CategoryRepository
	dsRepo   repositories.DatasetRepository
	verRepo  repositories.DatasetVersionRepository
}

func runVersionRepoTests(t *testing.T, makeEnv func() versionEnv) {
	ctx := context.Background()
	env := makeEnv()

	// 1) Создаём пользователя
	u := &entities.User{
		Username: "vuser",
		Email:    "vuser@example.com",
		Password: "pass",
		Country:  "NL",
		Role:     entities.RoleUser,
	}
	assert.NoError(t, env.userRepo.Create(ctx, u))
	assert.NotZero(t, u.ID)

	// 2) Создаём категорию
	c := &entities.Category{Name: "VerCat", Description: "desc"}
	assert.NoError(t, env.catRepo.Create(ctx, c))
	assert.NotZero(t, c.ID)

	// 3) Создаём датасет
	ds, err := entities.NewDataset("TestDS", "desc", u.ID, c.ID, true, time.Now().UTC())
	assert.NoError(t, err)
	assert.Zero(t, ds.ID)
	assert.NoError(t, env.dsRepo.Create(ctx, ds))
	assert.NotZero(t, ds.ID)

	// 4) Create Version
	ver, err := entities.NewDatasetVersion("v1.0", "/tmp/file1", "initial upload", ds.ID, time.Now().UTC())
	assert.NoError(t, err)
	assert.Zero(t, ver.ID)
	assert.NoError(t, env.verRepo.Create(ctx, ver))
	assert.NotZero(t, ver.ID)

	// 5) FindByID
	fetched, err := env.verRepo.FindByID(ctx, ver.ID)
	assert.NoError(t, err)
	assert.Equal(t, ver.Number, fetched.Number)
	assert.Equal(t, ver.DatasetID, fetched.DatasetID)

	// 6) FindByDatasetID
	versions, err := env.verRepo.FindByDatasetID(ctx, ds.ID)
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, ver.ID, versions[0].ID)

	// 7) Update
	fetched.ChangeLog = "fixed changelog"
	assert.NoError(t, env.verRepo.Update(ctx, fetched))
	updated, err := env.verRepo.FindByID(ctx, ver.ID)
	assert.NoError(t, err)
	assert.Equal(t, "fixed changelog", updated.ChangeLog)

	// 8) Delete
	assert.NoError(t, env.verRepo.Delete(ctx, ver.ID))
	missing, err := env.verRepo.FindByID(ctx, ver.ID)
	assert.ErrorIs(t, err, repositories.ErrVersionNotFound)
	assert.Nil(t, missing)
}

func TestVersionRepo_Implementations(t *testing.T) {
	tests := []struct {
		name    string
		makeEnv func() versionEnv
	}{
		{
			name: "Postgres",
			makeEnv: func() versionEnv {
				logger := zap.NewNop()
				return versionEnv{
					userRepo: postqbuild.NewUserRepo(dbPool, logger),
					catRepo:  postqbuild.NewCategoryRepo(dbPool, logger),
					dsRepo:   postqbuild.NewDatasetRepo(dbPool, logger),
					verRepo:  postqbuild.NewVersionRepo(dbPool, logger),
				}
			},
		},
		{
			name: "Mongo",
			makeEnv: func() versionEnv {
				logger := zap.NewNop()
				return versionEnv{
					userRepo: mongorepo.NewUserMongoRepo(mongoDB, logger),
					catRepo:  mongorepo.NewCategoryMongoRepo(mongoDB, logger),
					dsRepo:   mongorepo.NewDatasetMongoRepo(mongoDB, logger),
					verRepo:  mongorepo.NewVersionMongoRepo(mongoDB, logger),
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

			runVersionRepoTests(t, tc.makeEnv)
		})
	}
}
