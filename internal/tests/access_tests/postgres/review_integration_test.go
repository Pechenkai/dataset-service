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

//func TestReviewRepo_CRUD(t *testing.T) {
//	ctx := context.Background()
//	logger := zap.NewNop()
//	revRepo := postqbuild.NewReviewRepo(dbPool, logger)
//	dsRepo := postqbuild.NewDatasetRepo(dbPool, logger)
//
//	var userID uint64
//	err := dbPool.QueryRow(ctx,
//		`INSERT INTO users(username,email,password,registration_date,country,role)
//		 VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
//		"ruser", "ruser@example.com", "pass", time.Now(), "NL", entities.RoleUser,
//	).Scan(&userID)
//	assert.NoError(t, err)
//
//	var categoryID uint64
//	err = dbPool.QueryRow(ctx,
//		`INSERT INTO categories(name,description) VALUES($1,$2) RETURNING id`,
//		"RevCat", "rev desc",
//	).Scan(&categoryID)
//	assert.NoError(t, err)
//
//	dsEnt, err := entities.NewDataset("RevDS", "desc", userID, categoryID, true, time.Now())
//	assert.NoError(t, err)
//	err = dsRepo.Create(ctx, dsEnt)
//	assert.NoError(t, err)
//	assert.NotZero(t, dsEnt.ID)
//
//	rvEnt, err := entities.NewReview(userID, dsEnt.ID, entities.Rating4, time.Now(), "Great data")
//	assert.NoError(t, err)
//	assert.Zero(t, rvEnt.ID)
//
//	err = revRepo.Create(ctx, rvEnt)
//	assert.NoError(t, err)
//	assert.NotZero(t, rvEnt.ID)
//
//	fetched, err := revRepo.FindByID(ctx, rvEnt.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, entities.Rating4, fetched.Rating)
//	assert.Equal(t, "Great data", fetched.Text)
//
//	byDS, err := revRepo.FindByDatasetID(ctx, dsEnt.ID)
//	assert.NoError(t, err)
//	assert.Len(t, byDS, 1)
//	assert.Equal(t, rvEnt.ID, byDS[0].ID)
//
//	byUser, err := revRepo.FindByUserID(ctx, userID)
//	assert.NoError(t, err)
//	assert.Len(t, byUser, 1)
//	assert.Equal(t, rvEnt.ID, byUser[0].ID)
//
//	fetched.Rating = entities.Rating5
//	fetched.Text = "Updated text"
//	err = revRepo.Update(ctx, fetched)
//	assert.NoError(t, err)
//
//	updated, err := revRepo.FindByID(ctx, rvEnt.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, entities.Rating5, updated.Rating)
//	assert.Equal(t, "Updated text", updated.Text)
//
//	err = revRepo.Delete(ctx, rvEnt.ID)
//	assert.NoError(t, err)
//
//	missing, err := revRepo.FindByID(ctx, rvEnt.ID)
//	assert.ErrorIs(t, err, repositories.ErrReviewNotFound)
//	assert.Nil(t, missing)
//}

type reviewEnv struct {
	userRepo repositories.UserRepository
	catRepo  repositories.CategoryRepository
	dsRepo   repositories.DatasetRepository
	revRepo  repositories.ReviewRepository
}

func runReviewRepoTests(t *testing.T, makeEnv func() reviewEnv) {
	ctx := context.Background()
	env := makeEnv()

	u := &entities.User{
		Username: "ruser",
		Email:    "ruser@example.com",
		Password: "pass",
		Country:  "NL",
		Role:     entities.RoleUser,
	}
	assert.NoError(t, env.userRepo.Create(ctx, u))
	assert.NotZero(t, u.ID)

	// 2) Создаём категорию
	c := &entities.Category{Name: "RevCat", Description: "rev desc"}
	assert.NoError(t, env.catRepo.Create(ctx, c))
	assert.NotZero(t, c.ID)

	// 3) Создаём датасет
	ds, err := entities.NewDataset("RevDS", "desc", u.ID, c.ID, true, time.Now().UTC())
	assert.NoError(t, err)
	assert.Zero(t, ds.ID)
	assert.NoError(t, env.dsRepo.Create(ctx, ds))
	assert.NotZero(t, ds.ID)

	// 4) Create Review
	rv, err := entities.NewReview(u.ID, ds.ID, entities.Rating4, time.Now().UTC(), "Great data")
	assert.NoError(t, err)
	assert.Zero(t, rv.ID)
	assert.NoError(t, env.revRepo.Create(ctx, rv))
	assert.NotZero(t, rv.ID)

	// 5) FindByID
	fetched, err := env.revRepo.FindByID(ctx, rv.ID)
	assert.NoError(t, err)
	assert.Equal(t, entities.Rating4, fetched.Rating)
	assert.Equal(t, "Great data", fetched.Text)

	// 6) FindByDatasetID
	byDS, err := env.revRepo.FindByDatasetID(ctx, ds.ID)
	assert.NoError(t, err)
	assert.Len(t, byDS, 1)
	assert.Equal(t, rv.ID, byDS[0].ID)

	// 7) FindByUserID
	byUser, err := env.revRepo.FindByUserID(ctx, u.ID)
	assert.NoError(t, err)
	assert.Len(t, byUser, 1)
	assert.Equal(t, rv.ID, byUser[0].ID)

	// 8) Update
	fetched.Rating = entities.Rating5
	fetched.Text = "Updated text"
	assert.NoError(t, env.revRepo.Update(ctx, fetched))

	updated, err := env.revRepo.FindByID(ctx, rv.ID)
	assert.NoError(t, err)
	assert.Equal(t, entities.Rating5, updated.Rating)
	assert.Equal(t, "Updated text", updated.Text)

	// 9) Delete
	assert.NoError(t, env.revRepo.Delete(ctx, rv.ID))
	missing, err := env.revRepo.FindByID(ctx, rv.ID)
	assert.ErrorIs(t, err, repositories.ErrReviewNotFound)
	assert.Nil(t, missing)
}

func TestReviewRepo_Implementations(t *testing.T) {
	tests := []struct {
		name    string
		makeEnv func() reviewEnv
	}{
		{
			name: "Postgres",
			makeEnv: func() reviewEnv {
				logger := zap.NewNop()
				return reviewEnv{
					userRepo: postqbuild.NewUserRepo(dbPool, logger),
					catRepo:  postqbuild.NewCategoryRepo(dbPool, logger),
					dsRepo:   postqbuild.NewDatasetRepo(dbPool, logger),
					revRepo:  postqbuild.NewReviewRepo(dbPool, logger),
				}
			},
		},
		{
			name: "Mongo",
			makeEnv: func() reviewEnv {
				logger := zap.NewNop()
				return reviewEnv{
					userRepo: mongorepo.NewUserMongoRepo(mongoDB, logger),
					catRepo:  mongorepo.NewCategoryMongoRepo(mongoDB, logger),
					dsRepo:   mongorepo.NewDatasetMongoRepo(mongoDB, logger),
					revRepo:  mongorepo.NewReviewMongoRepo(mongoDB, logger),
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

			runReviewRepoTests(t, tc.makeEnv)
		})
	}
}
