package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"ppo/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/entities"
)

func must(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	must(err)

	poolCfg, err := pgxpool.ParseConfig(cfg.Database.DSN)
	must(err)
	poolCfg.MaxConns = cfg.Database.MaxConns
	poolCfg.MinConns = cfg.Database.MinConns
	poolCfg.MaxConnIdleTime = cfg.Database.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = cfg.Database.HealthCheckPeriod
	poolCfg.ConnConfig.ConnectTimeout = cfg.Database.ConnectTimeout

	pgPool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	must(err)
	defer pgPool.Close()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI))
	must(err)
	defer mongoClient.Disconnect(ctx)
	db := mongoClient.Database(cfg.Mongo.Database)
	maxIDs := map[string]uint64{}

	insertOne := func(collName string, doc interface{}, id uint64) {
		_, err := db.Collection(collName).InsertOne(ctx, doc)
		must(err)
		if id > maxIDs[collName] {
			maxIDs[collName] = id
		}
	}

	{
		rows, err := pgPool.Query(ctx, `
            SELECT id, username, email, password,
                   registration_date, country, is_blocked, role
              FROM users
        `)
		must(err)
		defer rows.Close()

		for rows.Next() {
			var u entities.User
			must(rows.Scan(
				&u.ID, &u.Username, &u.Email, &u.Password,
				&u.RegistrationDate, &u.Country, &u.IsBlocked, &u.Role,
			))
			insertOne("users", u, u.ID)
		}
	}

	{
		rows, err := pgPool.Query(ctx, `
            SELECT id, name, description FROM categories
        `)
		must(err)
		defer rows.Close()

		for rows.Next() {
			var c entities.Category
			must(rows.Scan(&c.ID, &c.Name, &c.Description))
			insertOne("categories", c, c.ID)
		}
	}

	{
		rows, err := pgPool.Query(ctx, `
            SELECT id, name, description, owner_id, category_id, is_public, created_at
              FROM datasets
        `)
		must(err)
		defer rows.Close()

		for rows.Next() {
			var d entities.Dataset
			must(rows.Scan(
				&d.ID, &d.Name, &d.Description,
				&d.OwnerID, &d.CategoryID, &d.IsPublic, &d.CreatedAt,
			))
			insertOne("datasets", d, d.ID)
		}
	}

	{
		rows, err := pgPool.Query(ctx, `
            SELECT id, number, upload_date, filepath, dataset_id, change_log
              FROM dataset_versions
        `)
		must(err)
		defer rows.Close()

		for rows.Next() {
			var v entities.DatasetVersion
			must(rows.Scan(
				&v.ID, &v.Number, &v.UploadDate,
				&v.Filepath, &v.DatasetID, &v.ChangeLog,
			))
			insertOne("dataset_versions", v, v.ID)
		}
	}

	{
		rows, err := pgPool.Query(ctx, `
            SELECT id, format, size, tags, dataset_version_id
              FROM metadata
        `)
		must(err)
		defer rows.Close()

		for rows.Next() {
			var m entities.Metadata
			must(rows.Scan(
				&m.ID, &m.Format, &m.Size, &m.Tags, &m.DatasetVersionID,
			))
			insertOne("metadata", m, m.ID)
		}
	}

	{
		rows, err := pgPool.Query(ctx, `
            SELECT user_id, dataset_id, created_at
              FROM subscriptions
        `)
		must(err)
		defer rows.Close()

		for rows.Next() {
			var s entities.Subscription
			must(rows.Scan(&s.UserID, &s.DatasetID, &s.CreatedAt))
			insertOne("subscriptions", s, 0)
		}
	}

	{
		rows, err := pgPool.Query(ctx, `
            SELECT id, user_id, dataset_id, rating, created_at, text
              FROM reviews
        `)
		must(err)
		defer rows.Close()

		for rows.Next() {
			var r entities.Review
			must(rows.Scan(
				&r.ID, &r.UserID, &r.DatasetID,
				&r.Rating, &r.CreatedAt, &r.Text,
			))
			insertOne("reviews", r, r.ID)
		}
	}

	{
		rows, err := pgPool.Query(ctx, `
            SELECT id, user_id, dataset_id, message, is_read, created_at
              FROM notifications
        `)
		must(err)
		defer rows.Close()

		for rows.Next() {
			var n entities.Notification
			must(rows.Scan(
				&n.ID, &n.UserID, &n.DatasetID,
				&n.Message, &n.IsRead, &n.CreatedAt,
			))
			insertOne("notifications", n, n.ID)
		}
	}

	{
		rows, err := pgPool.Query(ctx, `
            SELECT id, dataset_id, user_id, status, created_at
              FROM access_requests
        `)
		must(err)
		defer rows.Close()

		for rows.Next() {
			var a entities.AccessRequest
			must(rows.Scan(
				&a.ID, &a.DatasetID, &a.UserID,
				&a.Status, &a.CreatedAt,
			))
			insertOne("access_requests", a, a.ID)
		}
	}

	counterColl := db.Collection("counters")
	for name, mx := range maxIDs {
		_, err := counterColl.UpdateOne(
			ctx,
			bson.M{"_id": name},
			bson.M{"$set": bson.M{"seq": mx}},
			options.Update().SetUpsert(true),
		)
		must(err)
	}

	fmt.Println("✅ Миграция завершена, Mongo counters обновлены.")
}
