//go:build integration
// +build integration

package postqbuild_test

import (
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"path/filepath"
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
//		fmt.Fprintf(os.Stderr, "could not connect to postgres: %v\n", err)
//		os.Exit(1)
//	}
//
//	abs, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "migrations"))
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

var (
	dbPool  *pgxpool.Pool
	mongoDB *mongo.Database
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "docker pool error: %v\n", err)
		os.Exit(1)
	}
	pool.MaxWait = 60 * time.Second
	pgRes, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_DB=testdb",
		},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "docker run postgres error: %v\n", err)
		os.Exit(1)
	}
	mongoRes, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "6",
		Env:        []string{},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "docker run mongo error: %v\n", err)
		pool.Purge(pgRes)
		os.Exit(1)
	}

	defer pool.Purge(pgRes)
	defer pool.Purge(mongoRes)

	var pg *pgxpool.Pool
	if err := pool.Retry(func() error {
		dsn := fmt.Sprintf("postgres://postgres:secret@localhost:%s/testdb?sslmode=disable", pgRes.GetPort("5432/tcp"))
		var err error
		pg, err = pgxpool.New(context.Background(), dsn)
		if err != nil {
			return err
		}
		return pg.Ping(context.Background())
	}); err != nil {
		fmt.Fprintf(os.Stderr, "could not connect to postgres: %v\n", err)
		os.Exit(1)
	}

	migrationsDir, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "migrations"))
	if err != nil {
		log.Fatalf("failed to find migrations dir: %v", err)
	}
	migrator, err := migrate.New(
		"file://"+filepath.ToSlash(migrationsDir),
		fmt.Sprintf("postgres://postgres:secret@localhost:%s/testdb?sslmode=disable", pgRes.GetPort("5432/tcp")),
	)
	if err != nil {
		log.Fatalf("migrate.New error: %v", err)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrate up error: %v", err)
	}

	dbPool = pg

	var mc *mongo.Client
	if err := pool.Retry(func() error {
		var err error
		mc, err = mongo.Connect(context.Background(), options.Client().
			ApplyURI("mongodb://localhost:"+mongoRes.GetPort("27017/tcp")))
		if err != nil {
			return err
		}
		return mc.Ping(context.Background(), nil)
	}); err != nil {
		fmt.Fprintf(os.Stderr, "could not connect to mongo: %v\n", err)
		os.Exit(1)
	}
	mongoDB = mc.Database("myappdb")

	if err := migrateMongo(context.Background(), mongoDB); err != nil {
		log.Fatalf("mongo migration error: %v", err)
	}

	code := m.Run()

	dbPool.Close()
	mc.Disconnect(context.Background())

	os.Exit(code)
}

func migrateMongo(ctx context.Context, db *mongo.Database) error {
	idxs := []struct {
		coll   string
		models []mongo.IndexModel
	}{
		{"users", []mongo.IndexModel{
			{Keys: bson.D{{"email", 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{"registration_date", -1}}},
		}},
		{"categories", []mongo.IndexModel{
			{Keys: bson.D{{"name", 1}}, Options: options.Index().SetUnique(true)},
		}},
		{"datasets", []mongo.IndexModel{
			{Keys: bson.D{{"owner_id", 1}}},
			{Keys: bson.D{{"category_id", 1}, {"is_public", 1}, {"created_at", -1}}},
		}},
		{"dataset_versions", []mongo.IndexModel{
			{Keys: bson.D{{"dataset_id", 1}, {"upload_date", -1}}},
		}},
		{"metadata", []mongo.IndexModel{
			{Keys: bson.D{{"dataset_version_id", 1}}},
		}},
		{"subscriptions", []mongo.IndexModel{
			{Keys: bson.D{{"user_id", 1}, {"dataset_id", 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{"dataset_id", 1}}},
		}},
		{"reviews", []mongo.IndexModel{
			{Keys: bson.D{{"dataset_id", 1}, {"created_at", -1}}},
			{Keys: bson.D{{"user_id", 1}, {"created_at", -1}}},
		}},
		{"notifications", []mongo.IndexModel{
			{Keys: bson.D{{"user_id", 1}, {"created_at", -1}}},
		}},
		{"access_requests", []mongo.IndexModel{
			{Keys: bson.D{{"dataset_id", 1}, {"user_id", 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{"created_at", -1}}},
		}},
	}

	for _, spec := range idxs {
		coll := db.Collection(spec.coll)
		if _, err := coll.Indexes().CreateMany(ctx, spec.models); err != nil {
			return fmt.Errorf("create indexes on %s: %w", spec.coll, err)
		}
	}

	names := []string{
		"users", "categories", "datasets", "dataset_versions",
		"metadata", "reviews", "notifications", "access_requests",
	}
	docs := make([]interface{}, len(names))
	for i, n := range names {
		docs[i] = bson.D{{Key: "_id", Value: n}, {Key: "seq", Value: 0}}
	}
	if _, err := db.Collection("counters").InsertMany(ctx, docs); err != nil {
		return fmt.Errorf("init counters: %w", err)
	}
	return nil
}
