package postqbuild_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"

	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
	"ppo/internal/repositories"
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

func TestDatasetRepo_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := postqbuild.NewDatasetRepo(dbPool)

	var userID uint64
	err := dbPool.QueryRow(ctx,
		`INSERT INTO users(username,email,password,registration_date,country,role)
		    VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
		"u1", "u1@example.com", "hash", time.Now(), "NL", entities.RoleUser,
	).Scan(&userID)
	assert.NoError(t, err)

	var categoryID uint64
	err = dbPool.QueryRow(ctx,
		`INSERT INTO categories(name,description) VALUES($1,$2) RETURNING id`,
		"Cat1", "desc").Scan(&categoryID)
	assert.NoError(t, err)

	ds, err := entities.NewDataset("DS1", "Dataset 1", userID, categoryID, true, time.Now())
	assert.NoError(t, err)
	assert.Zero(t, ds.ID)

	err = repo.Create(ctx, ds)
	assert.NoError(t, err)
	assert.NotZero(t, ds.ID)

	fetched, err := repo.FindByID(ctx, ds.ID)
	assert.NoError(t, err)
	assert.Equal(t, ds.Name, fetched.Name)
	assert.Equal(t, ds.OwnerID, fetched.OwnerID)

	byUser, err := repo.FindByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, byUser, 1)
	assert.Equal(t, ds.ID, byUser[0].ID)

	all, err := repo.FindAll(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(all), 1)

	fetched.Name = "DS1-renamed"
	fetched.Description = "updated"
	err = repo.Update(ctx, fetched)
	assert.NoError(t, err)

	updated, err := repo.FindByID(ctx, ds.ID)
	assert.NoError(t, err)
	assert.Equal(t, "DS1-renamed", updated.Name)
	assert.Equal(t, "updated", updated.Description)

	err = repo.Delete(ctx, ds.ID)
	assert.NoError(t, err)

	missing, err := repo.FindByID(ctx, ds.ID)
	assert.ErrorIs(t, err, repositories.ErrDatasetNotFound)
	assert.Nil(t, missing)
}
