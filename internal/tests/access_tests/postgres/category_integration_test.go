//go:build integration
// +build integration

package postqbuild_test

import (
	"context"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	mongorepo "ppo/internal/dataaccess/repositories/mongo"
	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
	"ppo/internal/repositories"
	"testing"
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

//func TestCategoryRepo_CRUD_Integration(t *testing.T) {
//	logger := zap.NewNop()
//	repo := postqbuild.NewCategoryRepo(dbPool, logger)
//	ctx := context.Background()
//
//	c := &entities.Category{Name: "Dogs", Description: "All about dogs"}
//	err := repo.Create(ctx, c)
//	assert.NoError(t, err)
//	assert.NotZero(t, c.ID)
//
//	fetched, err := repo.FindByID(ctx, c.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, "Dogs", fetched.Name)
//	assert.Equal(t, "All about dogs", fetched.Description)
//
//	fetched.Description = "Dogs & Puppies"
//	err = repo.Update(ctx, fetched)
//	assert.NoError(t, err)
//
//	updated, err := repo.FindByID(ctx, c.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, "Dogs & Puppies", updated.Description)
//
//	all, err := repo.FindAll(ctx)
//	assert.NoError(t, err)
//	found := false
//	for _, cat := range all {
//		if cat.ID == c.ID {
//			found = true
//		}
//	}
//	assert.True(t, found)
//
//	err = repo.Delete(ctx, c.ID)
//	assert.NoError(t, err)
//
//	_, err = repo.FindByID(ctx, c.ID)
//	assert.ErrorIs(t, err, repositories.ErrCategoryNotFound)
//}
//
//func TestCreateDuplicateCategory_Integration(t *testing.T) {
//	logger := zap.NewNop()
//	repo := postqbuild.NewCategoryRepo(dbPool, logger)
//	ctx := context.Background()
//
//	c1 := &entities.Category{Name: "Unique", Description: ""}
//	assert.NoError(t, repo.Create(ctx, c1))
//
//	c2 := &entities.Category{Name: "Unique", Description: "dup"}
//	err := repo.Create(ctx, c2)
//	assert.ErrorIs(t, err, repositories.ErrCategoryAlreadyExists)
//
//	assert.NoError(t, repo.Delete(ctx, c1.ID))
//}

func runCategoryRepoTests(t *testing.T, factory func() repositories.CategoryRepository) {
	repo := factory()
	ctx := context.Background()

	// 1) Create
	c := &entities.Category{Name: "Dogs", Description: "All about dogs"}
	assert.NoError(t, repo.Create(ctx, c))
	assert.NotZero(t, c.ID)

	// 2) FindByID
	f, err := repo.FindByID(ctx, c.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Dogs", f.Name)

	// 3) Update
	f.Description = "Dogs & Puppies"
	assert.NoError(t, repo.Update(ctx, f))
	u, err := repo.FindByID(ctx, c.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Dogs & Puppies", u.Description)

	// 4) FindAll
	all, err := repo.FindAll(ctx)
	assert.NoError(t, err)
	found := false
	for _, x := range all {
		if x.ID == c.ID {
			found = true
		}
	}
	assert.True(t, found)

	// 5) Duplicate insert
	dup := &entities.Category{Name: "Dogs"}
	err = repo.Create(ctx, dup)
	assert.ErrorIs(t, err, repositories.ErrCategoryAlreadyExists)

	// 6) Delete
	assert.NoError(t, repo.Delete(ctx, c.ID))
	_, err = repo.FindByID(ctx, c.ID)
	assert.ErrorIs(t, err, repositories.ErrCategoryNotFound)
}

func TestCategoryRepo_Implementations(t *testing.T) {
	tests := []struct {
		name    string
		factory func() repositories.CategoryRepository
	}{
		{
			name: "Postgres",
			factory: func() repositories.CategoryRepository {
				return postqbuild.NewCategoryRepo(dbPool, zap.NewNop())
			},
		},
		{
			name: "Mongo",
			factory: func() repositories.CategoryRepository {
				return mongorepo.NewCategoryMongoRepo(mongoDB, zap.NewNop())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runCategoryRepoTests(t, tc.factory)
		})
	}
}
