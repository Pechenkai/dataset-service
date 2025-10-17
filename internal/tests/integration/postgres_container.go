//go:build integration || e2e

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"ppo/internal/config"
	"ppo/internal/dataaccess/repositories/postqbuild"
)

type PostgresContainer struct {
	Pool     *dockertest.Pool
	Resource *dockertest.Resource
	DB       *pgxpool.Pool
	Config   config.Database
	DSN      string
}

func StartPostgres(t *testing.T) *PostgresContainer {
	t.Helper()

	pool, err := dockertest.NewPool("")
	require.NoError(t, err, "failed to connect to Docker daemon")

	dbName := fmt.Sprintf("ppo_test_%d", time.Now().UnixNano())
	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "15",
			Env: []string{
				"POSTGRES_PASSWORD=pass",
				"POSTGRES_USER=postgres",
				fmt.Sprintf("POSTGRES_DB=%s", dbName),
			},
		},
		func(h *docker.HostConfig) {
			h.AutoRemove = true
		},
	)
	require.NoError(t, err, "failed to start postgres container")

	t.Cleanup(func() {
		_ = pool.Purge(resource)
	})

	port := resource.GetPort("5432/tcp")
	dsn := fmt.Sprintf("postgres://postgres:pass@localhost:%s/%s?sslmode=disable", port, dbName)

	cfg := config.Database{
		DSN:               dsn,
		MaxConns:          4,
		MinConns:          1,
		MaxConnIdleTime:   time.Minute,
		HealthCheckPeriod: time.Minute,
		ConnectTimeout:    5 * time.Second,
	}

	var db *pgxpool.Pool
	require.NoError(t, pool.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var err error
		db, err = postqbuild.NewPool(ctx, cfg)
		if err != nil {
			return err
		}

		if err := db.Ping(ctx); err != nil {
			db.Close()
			return err
		}
		return nil
	}), "postgres container did not become ready in time")

	t.Cleanup(func() {
		if db != nil {
			db.Close()
		}
	})

	m, err := migrate.New("file://"+MigrationsDir(), dsn)
	require.NoError(t, err, "failed to init migrations")

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err, "failed to apply migrations")
	}
	if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
		require.FailNow(t, "failed to close migrations", "sourceErr=%v dbErr=%v", sourceErr, dbErr)
	}

	return &PostgresContainer{
		Pool:     pool,
		Resource: resource,
		DB:       db,
		Config:   cfg,
		DSN:      dsn,
	}
}
