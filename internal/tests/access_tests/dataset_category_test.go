package access_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"ppo/internal/dataaccess/repositories/postgres"
	"ppo/internal/entities"
)

func setupPostgresContainer(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		Env:          map[string]string{"POSTGRES_PASSWORD": "secret", "POSTGRES_DB": "testdb", "POSTGRES_USER": "test"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}
	cont, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := cont.Host(ctx)
	require.NoError(t, err)
	port, err := cont.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://test:secret@%s:%s/testdb?sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	// Apply migrations (assumes migrate CLI or library invoked externally)
	// Alternatively, exec SQL files here before tests

	return pool, func() {
		pool.Close()
		cont.Terminate(ctx)
	}
}

func TestDatasetRepo_CRUD(t *testing.T) {
	pool, teardown := setupPostgresContainer(t)
	defer teardown()
	ctx := context.Background()

	// assume migrations applied externally
	dsRepo := postgres.NewDatasetRepo(pool)

	// prepare user and category (FK constraints)
	// using test SQL or separate repos
	_, err := pool.Exec(ctx, `INSERT INTO users(username,email,password,registration_date,country,role) VALUES('u','u@e','p',now(),'c','user')`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO categories(name,description) VALUES('cat','desc')`)
	require.NoError(t, err)

	// 1. Create & FindByID
	d := &entities.Dataset{
		Name:        "test-ds",
		Description: "desc",
		OwnerID:     1,
		CategoryID:  1,
		IsPublic:    true,
	}
	require.NoError(t, dsRepo.Create(ctx, d))
	require.NotZero(t, d.ID)

	got, err := dsRepo.FindByID(ctx, d.ID)
	require.NoError(t, err)
	require.Equal(t, "test-ds", got.Name)

	// 2. Update
	d.Name = "updated"
	require.NoError(t, dsRepo.Update(ctx, d))
	got2, err := dsRepo.FindByID(ctx, d.ID)
	require.NoError(t, err)
	require.Equal(t, "updated", got2.Name)

	// 3. Delete
	require.NoError(t, dsRepo.Delete(ctx, d.ID))
	_, err = dsRepo.FindByID(ctx, d.ID)
	require.Equal(t, postgres.ErrDatasetNotFound, err)
}

func TestCategoryRepo_UniqueConstraint(t *testing.T) {
	pool, teardown := setupPostgresContainer(t)
	defer teardown()
	ctx := context.Background()

	catRepo := postgres.NewCategoryRepo(pool)
	c := &entities.Category{Name: "X", Description: "d"}
	require.NoError(t, catRepo.Create(ctx, c))

	dup := &entities.Category{Name: "X", Description: "d2"}
	err := catRepo.Create(ctx, dup)
	require.Equal(t, postgres.ErrCategoryAlreadyExists, err)
}
