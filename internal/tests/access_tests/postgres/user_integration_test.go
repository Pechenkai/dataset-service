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

//func TestUserRepo_CRUD(t *testing.T) {
//	ctx := context.Background()
//	logger := zap.NewNop()
//	userRepo := postqbuild.NewUserRepo(dbPool, logger)
//
//	now := time.Now().UTC()
//	uEnt, err := entities.NewUser(
//		"testuser",
//		"testuser@example.com",
//		"secretpass",
//		"NL",
//		entities.RoleUser,
//		now,
//	)
//	assert.NoError(t, err)
//	assert.Zero(t, uEnt.ID)
//
//	err = userRepo.Create(ctx, uEnt)
//	assert.NoError(t, err)
//	assert.NotZero(t, uEnt.ID)
//
//	got, err := userRepo.FindByID(ctx, uEnt.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, "testuser", got.Username)
//	assert.Equal(t, "testuser@example.com", got.Email)
//	assert.Equal(t, "NL", got.Country)
//	assert.Equal(t, entities.RoleUser, got.Role)
//
//	byEmail, err := userRepo.FindByEmail(ctx, "testuser@example.com")
//	assert.NoError(t, err)
//	assert.NotNil(t, byEmail)
//	assert.Equal(t, uEnt.ID, byEmail.ID)
//
//	all, err := userRepo.FindAll(ctx)
//	assert.NoError(t, err)
//	found := false
//	for _, u := range all {
//		if u.ID == uEnt.ID {
//			found = true
//			break
//		}
//	}
//	assert.True(t, found, "created user must be in FindAll result")
//
//	got.Username = "updateduser"
//	got.Email = "updated@example.com"
//	got.Country = "DE"
//	got.Role = entities.RoleAdmin
//
//	err = userRepo.Update(ctx, got)
//	assert.NoError(t, err)
//
//	updated, err := userRepo.FindByID(ctx, uEnt.ID)
//	assert.NoError(t, err)
//	assert.Equal(t, "updateduser", updated.Username)
//	assert.Equal(t, "updated@example.com", updated.Email)
//	assert.Equal(t, "DE", updated.Country)
//	assert.Equal(t, entities.RoleAdmin, updated.Role)
//
//	err = userRepo.Delete(ctx, uEnt.ID)
//	assert.NoError(t, err)
//
//	missing, err := userRepo.FindByID(ctx, uEnt.ID)
//	assert.ErrorIs(t, err, repositories.ErrUserNotFound)
//	assert.Nil(t, missing)
//}

type userEnv struct {
	userRepo repositories.UserRepository
}

func runUserRepoTests(t *testing.T, makeEnv func() userEnv) {
	ctx := context.Background()
	env := makeEnv()

	// 1) Создать пользователя
	now := time.Now().UTC()
	u, err := entities.NewUser(
		"testuser",
		"testuser@example.com",
		"secretpass",
		"NL",
		entities.RoleUser,
		now,
	)
	assert.NoError(t, err)
	assert.Zero(t, u.ID)

	// 2) Create
	assert.NoError(t, env.userRepo.Create(ctx, u))
	assert.NotZero(t, u.ID)

	// 3) FindByID
	got, err := env.userRepo.FindByID(ctx, u.ID)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", got.Username)
	assert.Equal(t, "testuser@example.com", got.Email)
	assert.Equal(t, "NL", got.Country)
	assert.Equal(t, entities.RoleUser, got.Role)

	// 4) FindByEmail
	byEmail, err := env.userRepo.FindByEmail(ctx, "testuser@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, byEmail)
	assert.Equal(t, u.ID, byEmail.ID)

	// 5) FindAll
	all, err := env.userRepo.FindAll(ctx)
	assert.NoError(t, err)
	found := false
	for _, x := range all {
		if x.ID == u.ID {
			found = true
			break
		}
	}
	assert.True(t, found)

	// 6) Update
	got.Username = "updateduser"
	got.Email = "updated@example.com"
	got.Country = "DE"
	got.Role = entities.RoleAdmin
	assert.NoError(t, env.userRepo.Update(ctx, got))

	updated, err := env.userRepo.FindByID(ctx, u.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updateduser", updated.Username)
	assert.Equal(t, "updated@example.com", updated.Email)
	assert.Equal(t, "DE", updated.Country)
	assert.Equal(t, entities.RoleAdmin, updated.Role)

	// 7) Delete
	assert.NoError(t, env.userRepo.Delete(ctx, u.ID))

	missing, err := env.userRepo.FindByID(ctx, u.ID)
	assert.ErrorIs(t, err, repositories.ErrUserNotFound)
	assert.Nil(t, missing)
}

func TestUserRepo_Implementations(t *testing.T) {
	tests := []struct {
		name    string
		makeEnv func() userEnv
	}{
		{
			name: "Postgres",
			makeEnv: func() userEnv {
				logger := zap.NewNop()
				return userEnv{
					userRepo: postqbuild.NewUserRepo(dbPool, logger),
				}
			},
		},
		{
			name: "Mongo",
			makeEnv: func() userEnv {
				logger := zap.NewNop()
				return userEnv{
					userRepo: mongorepo.NewUserMongoRepo(mongoDB, logger),
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runUserRepoTests(t, tc.makeEnv)
		})
	}
}
