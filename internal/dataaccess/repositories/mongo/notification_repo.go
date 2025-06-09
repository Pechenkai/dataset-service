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

type NotificationMongoRepo struct {
	coll        *mongo.Collection
	counterColl *mongo.Collection
	logger      *zap.Logger
}

func NewNotificationMongoRepo(db *mongo.Database, logger *zap.Logger) *NotificationMongoRepo {
	_, _ = db.Collection("notifications").Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_notif_user_created_at"),
		},
	)
	return &NotificationMongoRepo{
		coll:        db.Collection("notifications"),
		counterColl: db.Collection("counters"),
		logger:      logger,
	}
}

func (r *NotificationMongoRepo) getNextSeq(ctx context.Context, name string) (uint64, error) {
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		Seq uint64 `bson:"seq"`
	}
	if err := r.counterColl.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result); err != nil {
		r.logger.Error("counter increment failed", zap.String("name", name), zap.Error(err))
		return 0, err
	}
	return result.Seq, nil
}

func (r *NotificationMongoRepo) Create(ctx context.Context, n *entities.Notification) error {
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	seq, err := r.getNextSeq(ctx, "notifications")
	if err != nil {
		return fmt.Errorf("create notification seq: %w", err)
	}
	n.ID = seq

	doc := bson.M{
		"id":         n.ID,
		"user_id":    n.UserID,
		"dataset_id": n.DatasetID,
		"message":    n.Message,
		"is_read":    n.IsRead,
		"created_at": primitive.NewDateTimeFromTime(n.CreatedAt),
	}
	if _, err := r.coll.InsertOne(ctx, doc); err != nil {
		r.logger.Error("failed to insert notification", zap.Error(err))
		return repositories.ErrNotificationCreate
	}
	return nil
}

func (r *NotificationMongoRepo) FindByID(ctx context.Context, id uint64) (*entities.Notification, error) {
	var doc bson.M
	err := r.coll.FindOne(ctx, bson.M{"id": id}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, repositories.ErrNotificationNotFound
	}
	if err != nil {
		r.logger.Error("find notification by id failed", zap.Error(err), zap.Uint64("id", id))
		return nil, repositories.ErrNotificationScanRow
	}

	created := doc["created_at"].(primitive.DateTime)
	n := &entities.Notification{
		ID:        uint64(doc["id"].(int64)),
		UserID:    uint64(doc["user_id"].(int64)),
		DatasetID: uint64(doc["dataset_id"].(int64)),
		Message:   doc["message"].(string),
		IsRead:    doc["is_read"].(bool),
		CreatedAt: created.Time(),
	}
	return n, nil
}

func (r *NotificationMongoRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Notification, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		r.logger.Error("find by user failed", zap.Error(err), zap.Uint64("user_id", userID))
		return nil, repositories.ErrNotificationScanRow
	}
	defer cursor.Close(ctx)

	var list []*entities.Notification
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Error("decode notification failed", zap.Error(err))
			return nil, repositories.ErrNotificationScanRow
		}
		created := doc["created_at"].(primitive.DateTime)
		n := &entities.Notification{
			ID:        uint64(doc["id"].(int64)),
			UserID:    uint64(doc["user_id"].(int64)),
			DatasetID: uint64(doc["dataset_id"].(int64)),
			Message:   doc["message"].(string),
			IsRead:    doc["is_read"].(bool),
			CreatedAt: created.Time(),
		}
		list = append(list, n)
	}
	if err := cursor.Err(); err != nil {
		r.logger.Error("cursor error on find by user", zap.Error(err))
		return nil, repositories.ErrNotificationIterateRows
	}
	return list, nil
}

func (r *NotificationMongoRepo) Update(ctx context.Context, n *entities.Notification) error {
	filter := bson.M{"id": n.ID}
	update := bson.M{"$set": bson.M{"message": n.Message, "is_read": n.IsRead}}
	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("update notification failed", zap.Error(err), zap.Uint64("id", n.ID))
		return repositories.ErrNotificationUpdateFail
	}
	if res.MatchedCount == 0 {
		return repositories.ErrNotificationNotFound
	}
	return nil
}

func (r *NotificationMongoRepo) Delete(ctx context.Context, id uint64) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		r.logger.Error("delete notification failed", zap.Error(err), zap.Uint64("id", id))
		return repositories.ErrNotificationDeleteFail
	}
	if res.DeletedCount == 0 {
		return repositories.ErrNotificationNotFound
	}
	return nil
}
