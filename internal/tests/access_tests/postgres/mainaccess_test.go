package postqbuild_test

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
		fmt.Fprintf(os.Stderr, "could not connect to postgres: %v\n", err)
		os.Exit(1)
	}

	abs, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "migrations"))
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
