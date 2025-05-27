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

func TestUserRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	userRepo := postgres.NewUserRepo(dbPool)

	now := time.Now().UTC()
	uEnt, err := entities.NewUser(
		"testuser",
		"testuser@example.com",
		"secretpass",
		"NL",
		entities.RoleUser,
		now,
	)
	assert.NoError(t, err)
	assert.Zero(t, uEnt.ID)

	err = userRepo.Create(ctx, uEnt)
	assert.NoError(t, err)
	assert.NotZero(t, uEnt.ID)

	got, err := userRepo.FindByID(ctx, uEnt.ID)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", got.Username)
	assert.Equal(t, "testuser@example.com", got.Email)
	assert.Equal(t, "NL", got.Country)
	assert.Equal(t, entities.RoleUser, got.Role)

	byEmail, err := userRepo.FindByEmail(ctx, "testuser@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, byEmail)
	assert.Equal(t, uEnt.ID, byEmail.ID)

	all, err := userRepo.FindAll(ctx)
	assert.NoError(t, err)
	found := false
	for _, u := range all {
		if u.ID == uEnt.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "created user must be in FindAll result")

	got.Username = "updateduser"
	got.Email = "updated@example.com"
	got.Country = "DE"
	got.Role = entities.RoleAdmin

	err = userRepo.Update(ctx, got)
	assert.NoError(t, err)

	updated, err := userRepo.FindByID(ctx, uEnt.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updateduser", updated.Username)
	assert.Equal(t, "updated@example.com", updated.Email)
	assert.Equal(t, "DE", updated.Country)
	assert.Equal(t, entities.RoleAdmin, updated.Role)

	err = userRepo.Delete(ctx, uEnt.ID)
	assert.NoError(t, err)

	missing, err := userRepo.FindByID(ctx, uEnt.ID)
	assert.ErrorIs(t, err, postgres.ErrUserNotFound)
	assert.Nil(t, missing)
}
