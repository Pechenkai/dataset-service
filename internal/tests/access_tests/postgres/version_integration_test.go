package postgres_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"

	"ppo/internal/dataaccess/repositories/postgres"
	"ppo/internal/entities"
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
//		Repository: "postgres",
//		Tag:        "15-alpine",
//		Env: []string{
//			"POSTGRES_USER=postgres",
//			"POSTGRES_PASSWORD=secret",
//			"POSTGRES_DB=testdb",
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
//			"postgres://postgres:secret@localhost:%s/testdb?sslmode=disable",
//			resource.GetPort("5432/tcp"),
//		)
//		pool, err = pgxpool.New(context.Background(), dsn)
//		if err != nil {
//			return err
//		}
//		return pool.Ping(context.Background())
//	}); err != nil {
//		fmt.Fprintf(os.Stderr, "could not connect to Postgres: %v\n", err)
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
//		"postgres://postgres:secret@localhost:%s/testdb?sslmode=disable",
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

func TestDatasetVersionRepo_CRUD(t *testing.T) {
	ctx := context.Background()

	var userID uint64
	err := dbPool.QueryRow(ctx,
		`INSERT INTO users(username,email,password,registration_date,country,role)
		 VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
		"vuser", "vuser@example.com", "pass", time.Now(), "NL", entities.RoleUser,
	).Scan(&userID)
	assert.NoError(t, err)

	var categoryID uint64
	err = dbPool.QueryRow(ctx,
		`INSERT INTO categories(name,description) VALUES($1,$2) RETURNING id`,
		"VerCat", "desc",
	).Scan(&categoryID)
	assert.NoError(t, err)

	dsRepo := postgres.NewDatasetRepo(dbPool)
	dsEnt, err := entities.NewDataset("TestDS", "desc", userID, categoryID, true, time.Now())
	assert.NoError(t, err)
	assert.Zero(t, dsEnt.ID)

	err = dsRepo.Create(ctx, dsEnt)
	assert.NoError(t, err)
	assert.NotZero(t, dsEnt.ID)

	verRepo := postgres.NewVersionRepo(dbPool)
	verEnt, err := entities.NewDatasetVersion("v1.0", "/tmp/file1", "initial upload", dsEnt.ID, time.Now())
	assert.NoError(t, err)
	assert.Zero(t, verEnt.ID)

	err = verRepo.Create(ctx, verEnt)
	assert.NoError(t, err)
	assert.NotZero(t, verEnt.ID)

	fetched, err := verRepo.FindByID(ctx, verEnt.ID)
	assert.NoError(t, err)
	assert.Equal(t, verEnt.Number, fetched.Number)
	assert.Equal(t, verEnt.DatasetID, fetched.DatasetID)

	versions, err := verRepo.FindByDatasetID(ctx, dsEnt.ID)
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, verEnt.ID, versions[0].ID)

	fetched.ChangeLog = "fixed changelog"
	err = verRepo.Update(ctx, fetched)
	assert.NoError(t, err)

	updated, err := verRepo.FindByID(ctx, verEnt.ID)
	assert.NoError(t, err)
	assert.Equal(t, "fixed changelog", updated.ChangeLog)

	err = verRepo.Delete(ctx, verEnt.ID)
	assert.NoError(t, err)

	missing, err := verRepo.FindByID(ctx, verEnt.ID)
	assert.ErrorIs(t, err, postgres.ErrVersionNotFound)
	assert.Nil(t, missing)
}
