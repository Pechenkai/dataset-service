package postgres

import (
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"

	"ppo/internal/dataaccess/repositories/postgres"
	"ppo/internal/entities"
)

var dbPool *pgxpool.Pool

func TestMain(m *testing.M) {
	p, err := dockertest.NewPool("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "docker pool error: %v\n", err)
		os.Exit(1)
	}
	resource, err := p.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_DB=testdb",
		},
	}, func(cfg *docker.HostConfig) {
		cfg.AutoRemove = true
		cfg.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "docker run error: %v\n", err)
		os.Exit(1)
	}
	defer p.Purge(resource)

	var pool *pgxpool.Pool
	if err := p.Retry(func() error {
		dsn := fmt.Sprintf(
			"postgres://postgres:secret@localhost:%s/testdb?sslmode=disable",
			resource.GetPort("5432/tcp"),
		)
		pool, err = pgxpool.New(context.Background(), dsn)
		if err != nil {
			return err
		}
		return pool.Ping(context.Background())
	}); err != nil {
		fmt.Fprintf(os.Stderr, "could not connect to Postgres: %v\n", err)
		os.Exit(1)
	}

	abs, err := filepath.Abs(filepath.Join("..", "..", "..", "migrations"))
	if err != nil {
		panic(err)
	}
	migrationsURL := "file://" + filepath.ToSlash(abs)
	fmt.Println("Using migrationsURL:", migrationsURL)

	pgURL := fmt.Sprintf(
		"postgres://postgres:secret@localhost:%s/testdb?sslmode=disable",
		resource.GetPort("5432/tcp"),
	)

	migrator, err := migrate.New(migrationsURL, pgURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate.New error: %v\n", err)
		os.Exit(1)
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		fmt.Fprintf(os.Stderr, "migrate up error: %v\n", err)
		os.Exit(1)
	}

	dbPool = pool

	code := m.Run()

	dbPool.Close()
	os.Exit(code)
}

func TestSubscriptionRepo_Behavior(t *testing.T) {
	ctx := context.Background()
	subRepo := postgres.NewSubscriptionRepo(dbPool)
	dsRepo := postgres.NewDatasetRepo(dbPool)

	var userID uint64
	err := dbPool.QueryRow(ctx,
		`INSERT INTO users(username,email,password,registration_date,country,role)
		 VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
		"suser", "suser@example.com", "pass", time.Now(), "NL", entities.RoleUser,
	).Scan(&userID)
	assert.NoError(t, err)

	var categoryID uint64
	err = dbPool.QueryRow(ctx,
		`INSERT INTO categories(name,description) VALUES($1,$2) RETURNING id`,
		"SubCat", "sub desc",
	).Scan(&categoryID)
	assert.NoError(t, err)

	dsEnt, err := entities.NewDataset("SubDS", "desc", userID, categoryID, false, time.Now())
	assert.NoError(t, err)
	err = dsRepo.Create(ctx, dsEnt)
	assert.NoError(t, err)
	assert.NotZero(t, dsEnt.ID)

	err = subRepo.Subscribe(ctx, userID, dsEnt.ID)
	assert.NoError(t, err)

	subscribed, err := subRepo.IsSubscribed(ctx, userID, dsEnt.ID)
	assert.NoError(t, err)
	assert.True(t, subscribed)

	subs, err := subRepo.GetSubscribers(ctx, dsEnt.ID)
	assert.NoError(t, err)
	assert.Contains(t, subs, userID)

	err = subRepo.Subscribe(ctx, userID, dsEnt.ID)
	assert.Error(t, err)

	ok, err := subRepo.IsSubscribed(ctx, userID+999, dsEnt.ID)
	assert.NoError(t, err)
	assert.False(t, ok)
}
