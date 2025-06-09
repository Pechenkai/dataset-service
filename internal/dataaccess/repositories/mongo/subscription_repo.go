package postqbuild

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type SubscriptionMongoRepo struct {
	coll   *mongo.Collection
	logger *zap.Logger
}

func NewSubscriptionMongoRepo(db *mongo.Database, logger *zap.Logger) *SubscriptionMongoRepo {
	// Уникальный индекс по паре user_id+dataset_id
	idx := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "dataset_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("idx_sub_user_dataset"),
	}
	if _, err := db.Collection("subscriptions").Indexes().CreateOne(context.Background(), idx); err != nil {
		logger.Warn("could not ensure unique index on subscriptions", zap.Error(err))
	}
	return &SubscriptionMongoRepo{
		coll:   db.Collection("subscriptions"),
		logger: logger,
	}
}

func (r *SubscriptionMongoRepo) Create(ctx context.Context, s *entities.Subscription) error {
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}
	r.logger.Debug("Create Subscription called",
		zap.Uint64("user_id", s.UserID),
		zap.Uint64("dataset_id", s.DatasetID),
	)

	doc := bson.M{
		"user_id":    s.UserID,
		"dataset_id": s.DatasetID,
		"created_at": primitive.NewDateTimeFromTime(s.CreatedAt),
	}
	_, err := r.coll.InsertOne(ctx, doc)
	if err != nil {
		var we mongo.WriteException
		if errors.As(err, &we) {
			for _, e := range we.WriteErrors {
				if e.Code == 11000 {
					r.logger.Warn("duplicate subscription on create",
						zap.Uint64("user_id", s.UserID),
						zap.Uint64("dataset_id", s.DatasetID),
					)
					return repositories.ErrAlreadySubscribed
				}
			}
		}
		r.logger.Error("failed to insert subscription", zap.Error(err))
		return fmt.Errorf("subscribe: %w", err)
	}
	r.logger.Info("subscription created successfully",
		zap.Uint64("user_id", s.UserID),
		zap.Uint64("dataset_id", s.DatasetID),
	)
	return nil
}

func (r *SubscriptionMongoRepo) IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error) {
	r.logger.Debug("IsSubscribed called",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
	)

	err := r.coll.FindOne(ctx, bson.M{
		"user_id":    userID,
		"dataset_id": datasetID,
	}).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		r.logger.Error("error executing IsSubscribed", zap.Error(err),
			zap.Uint64("user_id", userID), zap.Uint64("dataset_id", datasetID))
		return false, fmt.Errorf("is subscribed: %w", err)
	}
	return true, nil
}

func (r *SubscriptionMongoRepo) GetSubscribers(ctx context.Context, datasetID uint64) ([]uint64, error) {
	r.logger.Debug("GetSubscribers called", zap.Uint64("dataset_id", datasetID))

	cursor, err := r.coll.Find(ctx, bson.M{"dataset_id": datasetID}, options.Find())
	if err != nil {
		r.logger.Error("failed to execute GetSubscribers query", zap.Error(err), zap.Uint64("dataset_id", datasetID))
		return nil, fmt.Errorf("get subscribers: %w", err)
	}
	defer cursor.Close(ctx)

	var ids []uint64
	for cursor.Next(ctx) {
		var doc struct {
			UserID uint64 `bson:"user_id"`
		}
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Error("failed to decode subscriber row", zap.Error(err))
			return nil, fmt.Errorf("scan subscriber row: %w", err)
		}
		ids = append(ids, doc.UserID)
	}
	if err := cursor.Err(); err != nil {
		r.logger.Error("error iterating subscriber rows", zap.Error(err))
		return nil, fmt.Errorf("iterate subscriber rows: %w", err)
	}
	r.logger.Info("subscribers fetched successfully",
		zap.Uint64("dataset_id", datasetID), zap.Int("count", len(ids)))
	return ids, nil
}

func (r *SubscriptionMongoRepo) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	r.logger.Debug("Unsubscribe called",
		zap.Uint64("user_id", userID), zap.Uint64("dataset_id", datasetID))

	res, err := r.coll.DeleteOne(ctx, bson.M{"user_id": userID, "dataset_id": datasetID})
	if err != nil {
		r.logger.Error("failed to execute Unsubscribe query", zap.Error(err),
			zap.Uint64("user_id", userID), zap.Uint64("dataset_id", datasetID))
		return fmt.Errorf("unsubscribe: %w", err)
	}
	if res.DeletedCount == 0 {
		return repositories.ErrSubscriptionNotFound
	}
	return nil
}

func (r *SubscriptionMongoRepo) GetByUser(ctx context.Context, userID uint64) ([]uint64, error) {
	r.logger.Debug("GetByUser called", zap.Uint64("user_id", userID))

	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID}, options.Find())
	if err != nil {
		r.logger.Error("failed to execute GetByUser query", zap.Error(err), zap.Uint64("user_id", userID))
		return nil, fmt.Errorf("get subscriptions by user: %w", err)
	}
	defer cursor.Close(ctx)

	var dsids []uint64
	for cursor.Next(ctx) {
		var doc struct {
			DatasetID uint64 `bson:"dataset_id"`
		}
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Error("failed to decode subscription row", zap.Error(err))
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		dsids = append(dsids, doc.DatasetID)
	}
	if err := cursor.Err(); err != nil {
		r.logger.Error("error iterating subscription rows", zap.Error(err))
		return nil, fmt.Errorf("iterate subscription rows: %w", err)
	}
	r.logger.Info("subscriptions fetched by user successfully",
		zap.Uint64("user_id", userID), zap.Int("count", len(dsids)))
	return dsids, nil
}
