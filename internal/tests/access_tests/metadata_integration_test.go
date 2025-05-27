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

func TestMetadataRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	metaRepo := postgres.NewMetadataRepo(dbPool)
	dsRepo := postgres.NewDatasetRepo(dbPool)
	verRepo := postgres.NewVersionRepo(dbPool)

	var userID uint64
	err := dbPool.QueryRow(ctx,
		`INSERT INTO users(username,email,password,registration_date,country,role)
		 VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
		"muser", "muser@example.com", "pass", time.Now(), "NL", entities.RoleUser,
	).Scan(&userID)
	assert.NoError(t, err)

	var categoryID uint64
	err = dbPool.QueryRow(ctx,
		`INSERT INTO categories(name,description) VALUES($1,$2) RETURNING id`,
		"MetaCat", "meta desc",
	).Scan(&categoryID)
	assert.NoError(t, err)

	dsEnt, err := entities.NewDataset("MDataset", "desc", userID, categoryID, false, time.Now())
	assert.NoError(t, err)
	err = dsRepo.Create(ctx, dsEnt)
	assert.NoError(t, err)
	assert.NotZero(t, dsEnt.ID)

	verEnt, err := entities.NewDatasetVersion("v1.0", "/tmp/f.mp4", "initial", dsEnt.ID, time.Now())
	assert.NoError(t, err)
	err = verRepo.Create(ctx, verEnt)
	assert.NoError(t, err)
	assert.NotZero(t, verEnt.ID)

	mdEnt, err := entities.NewMetadata("mp4", "tag1,tag2", 12345, verEnt.ID)
	assert.NoError(t, err)
	assert.Zero(t, mdEnt.ID)

	err = metaRepo.Create(ctx, mdEnt)
	assert.NoError(t, err)
	assert.NotZero(t, mdEnt.ID)

	got, err := metaRepo.FindByID(ctx, mdEnt.ID)
	assert.NoError(t, err)
	assert.Equal(t, mdEnt.Format, got.Format)
	assert.Equal(t, mdEnt.Size, got.Size)
	assert.Equal(t, mdEnt.DatasetVersionID, got.DatasetVersionID)

	list, err := metaRepo.FindByDatasetID(ctx, verEnt.ID)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, mdEnt.ID, list[0].ID)

	got.Tags = "tagA,tagB"
	got.Size = 54321
	err = metaRepo.Update(ctx, got)
	assert.NoError(t, err)

	updated, err := metaRepo.FindByID(ctx, mdEnt.ID)
	assert.NoError(t, err)
	assert.Equal(t, "tagA,tagB", updated.Tags)
	assert.Equal(t, uint64(54321), updated.Size)

	err = metaRepo.Delete(ctx, mdEnt.ID)
	assert.NoError(t, err)

	missing, err := metaRepo.FindByID(ctx, mdEnt.ID)
	assert.ErrorIs(t, err, postgres.ErrMetadataNotFound)
	assert.Nil(t, missing)
}
