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

//func TestSubscriptionRepo_Behavior(t *testing.T) {
//	ctx := context.Background()
//	logger := zap.NewNop()
//	subRepo := postqbuild.NewSubscriptionRepo(dbPool, logger)
//	dsRepo := postqbuild.NewDatasetRepo(dbPool, logger)
//
//	var userID uint64
//	err := dbPool.QueryRow(ctx,
//		`INSERT INTO users(username,email,password,registration_date,country,role)
//		 VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
//		"suser", "suser@example.com", "pass", time.Now(), "NL", entities.RoleUser,
//	).Scan(&userID)
//	assert.NoError(t, err)
//
//	var categoryID uint64
//	err = dbPool.QueryRow(ctx,
//		`INSERT INTO categories(name,description) VALUES($1,$2) RETURNING id`,
//		"SubCat", "sub desc",
//	).Scan(&categoryID)
//	assert.NoError(t, err)
//
//	dsEnt, err := entities.NewDataset("SubDS", "desc", userID, categoryID, false, time.Now())
//	assert.NoError(t, err)
//	err = dsRepo.Create(ctx, dsEnt)
//	assert.NoError(t, err)
//	assert.NotZero(t, dsEnt.ID)
//
//	sub, err := entities.NewSubscription(userID, dsEnt.ID, time.Now())
//	assert.NoError(t, err)
//
//	err = subRepo.Create(ctx, sub)
//	assert.NoError(t, err)
//
//	subscribed, err := subRepo.IsSubscribed(ctx, userID, dsEnt.ID)
//	assert.NoError(t, err)
//	assert.True(t, subscribed)
//
//	subs, err := subRepo.GetSubscribers(ctx, dsEnt.ID)
//	assert.NoError(t, err)
//	assert.Contains(t, subs, userID)
//
//	err = subRepo.Create(ctx, sub)
//	assert.Error(t, err)
//
//	ok, err := subRepo.IsSubscribed(ctx, userID+999, dsEnt.ID)
//	assert.NoError(t, err)
//	assert.False(t, ok)
//}

type subscriptionEnv struct {
	userRepo repositories.UserRepository
	catRepo  repositories.CategoryRepository
	dsRepo   repositories.DatasetRepository
	subRepo  repositories.SubscriptionRepository
}

func runSubscriptionRepoTests(t *testing.T, makeEnv func() subscriptionEnv) {
	ctx := context.Background()
	env := makeEnv()

	// 1) Создать пользователя
	u := &entities.User{
		Username: "suser",
		Email:    "suser@example.com",
		Password: "pass",
		Country:  "NL",
		Role:     entities.RoleUser,
	}
	assert.NoError(t, env.userRepo.Create(ctx, u))
	assert.NotZero(t, u.ID)

	// 2) Создать категорию
	c := &entities.Category{Name: "SubCat", Description: "sub desc"}
	assert.NoError(t, env.catRepo.Create(ctx, c))
	assert.NotZero(t, c.ID)

	// 3) Создать датасет
	ds, err := entities.NewDataset("SubDS", "desc", u.ID, c.ID, false, time.Now().UTC())
	assert.NoError(t, err)
	assert.NoError(t, env.dsRepo.Create(ctx, ds))
	assert.NotZero(t, ds.ID)

	// 4) Create Subscription
	sub, err := entities.NewSubscription(u.ID, ds.ID, time.Now().UTC())
	assert.NoError(t, err)
	//assert.Zero(t, sub.UserID) // zero before Create
	sub.UserID, sub.DatasetID = u.ID, ds.ID

	assert.NoError(t, env.subRepo.Create(ctx, sub))

	// 5) IsSubscribed true
	ok, err := env.subRepo.IsSubscribed(ctx, u.ID, ds.ID)
	assert.NoError(t, err)
	assert.True(t, ok)

	// 6) GetSubscribers contains userID
	subs, err := env.subRepo.GetSubscribers(ctx, ds.ID)
	assert.NoError(t, err)
	assert.Contains(t, subs, u.ID)

	// 7) Duplicate Create returns ErrAlreadySubscribed
	err = env.subRepo.Create(ctx, sub)
	assert.ErrorIs(t, err, repositories.ErrAlreadySubscribed)

	// 8) IsSubscribed false for non-existent
	ok, err = env.subRepo.IsSubscribed(ctx, u.ID+999, ds.ID)
	assert.NoError(t, err)
	assert.False(t, ok)

	// 9) GetByUser returns dataset ID
	dsids, err := env.subRepo.GetByUser(ctx, u.ID)
	assert.NoError(t, err)
	assert.Contains(t, dsids, ds.ID)

	// 10) Unsubscribe
	assert.NoError(t, env.subRepo.Unsubscribe(ctx, u.ID, ds.ID))
	ok, err = env.subRepo.IsSubscribed(ctx, u.ID, ds.ID)
	assert.NoError(t, err)
	assert.False(t, ok)

	// 11) Unsubscribe non-existent returns ErrSubscriptionNotFound
	err = env.subRepo.Unsubscribe(ctx, u.ID, ds.ID)
	assert.ErrorIs(t, err, repositories.ErrSubscriptionNotFound)
}

func TestSubscriptionRepo_Implementations(t *testing.T) {
	tests := []struct {
		name    string
		makeEnv func() subscriptionEnv
	}{
		{
			name: "Postgres",
			makeEnv: func() subscriptionEnv {
				logger := zap.NewNop()
				return subscriptionEnv{
					userRepo: postqbuild.NewUserRepo(dbPool, logger),
					catRepo:  postqbuild.NewCategoryRepo(dbPool, logger),
					dsRepo:   postqbuild.NewDatasetRepo(dbPool, logger),
					subRepo:  postqbuild.NewSubscriptionRepo(dbPool, logger),
				}
			},
		},
		{
			name: "Mongo",
			makeEnv: func() subscriptionEnv {
				logger := zap.NewNop()
				return subscriptionEnv{
					userRepo: mongorepo.NewUserMongoRepo(mongoDB, logger),
					catRepo:  mongorepo.NewCategoryMongoRepo(mongoDB, logger),
					dsRepo:   mongorepo.NewDatasetMongoRepo(mongoDB, logger),
					subRepo:  mongorepo.NewSubscriptionMongoRepo(mongoDB, logger),
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

			runSubscriptionRepoTests(t, tc.makeEnv)
		})
	}
}
