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

type ReviewMongoRepo struct {
	coll        *mongo.Collection
	counterColl *mongo.Collection
	logger      *zap.Logger
}

func NewReviewMongoRepo(db *mongo.Database, logger *zap.Logger) *ReviewMongoRepo {
	_, _ = db.Collection("reviews").Indexes().CreateMany(
		context.Background(),
		[]mongo.IndexModel{
			{Keys: bson.D{{Key: "dataset_id", Value: 1}, {Key: "created_at", Value: -1}}},
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		},
	)
	return &ReviewMongoRepo{
		coll:        db.Collection("reviews"),
		counterColl: db.Collection("counters"),
		logger:      logger,
	}
}

func (r *ReviewMongoRepo) getNextSeq(ctx context.Context, name string) (uint64, error) {
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var res struct {
		Seq uint64 `bson:"seq"`
	}
	if err := r.counterColl.FindOneAndUpdate(ctx, filter, update, opts).Decode(&res); err != nil {
		r.logger.Error("counter increment failed", zap.String("name", name), zap.Error(err))
		return 0, err
	}
	return res.Seq, nil
}

func (r *ReviewMongoRepo) Create(ctx context.Context, rv *entities.Review) error {
	if rv.CreatedAt.IsZero() {
		rv.CreatedAt = time.Now().UTC()
	}
	seq, err := r.getNextSeq(ctx, "reviews")
	if err != nil {
		return fmt.Errorf("create review seq: %w", err)
	}
	rv.ID = seq

	doc := bson.M{
		"id":         rv.ID,
		"user_id":    rv.UserID,
		"dataset_id": rv.DatasetID,
		"rating":     rv.Rating,
		"text":       rv.Text,
		"created_at": primitive.NewDateTimeFromTime(rv.CreatedAt),
	}
	if _, err := r.coll.InsertOne(ctx, doc); err != nil {
		r.logger.Error("insert review failed", zap.Error(err))
		return repositories.ErrReviewNotFound // or appropriate ErrReviewCreate? Use creation error?
	}
	return nil
}

func (r *ReviewMongoRepo) Update(ctx context.Context, rv *entities.Review) error {
	filter := bson.M{"id": rv.ID}
	update := bson.M{"$set": bson.M{"rating": rv.Rating, "text": rv.Text}}
	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("update review failed", zap.Error(err), zap.Uint64("id", rv.ID))
		return repositories.ErrReviewNotFound
	}
	if res.MatchedCount == 0 {
		return repositories.ErrReviewNotFound
	}
	return nil
}

func (r *ReviewMongoRepo) Delete(ctx context.Context, id uint64) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		r.logger.Error("delete review failed", zap.Error(err), zap.Uint64("id", id))
		return repositories.ErrReviewNotFound
	}
	if res.DeletedCount == 0 {
		return repositories.ErrReviewNotFound
	}
	return nil
}

func (r *ReviewMongoRepo) FindByID(ctx context.Context, id uint64) (*entities.Review, error) {
	var doc bson.M
	err := r.coll.FindOne(ctx, bson.M{"id": id}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, repositories.ErrReviewNotFound
	}
	if err != nil {
		r.logger.Error("find review by id failed", zap.Error(err), zap.Uint64("id", id))
		return nil, repositories.ErrReviewNotFound
	}
	created := doc["created_at"].(primitive.DateTime)
	rv := &entities.Review{
		ID:        uint64(doc["id"].(int64)),
		UserID:    uint64(doc["user_id"].(int64)),
		DatasetID: uint64(doc["dataset_id"].(int64)),
		Rating:    entities.Rating(doc["rating"].(int32)),
		Text:      doc["text"].(string),
		CreatedAt: created.Time(),
	}
	return rv, nil
}

func (r *ReviewMongoRepo) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Review, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"dataset_id": datasetID}, opts)
	if err != nil {
		r.logger.Error("find reviews by dataset failed", zap.Error(err), zap.Uint64("dataset_id", datasetID))
		return nil, repositories.ErrReviewNotFound
	}
	defer cursor.Close(ctx)
	var list []*entities.Review
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Error("decode review failed", zap.Error(err))
			return nil, repositories.ErrReviewNotFound
		}
		created := doc["created_at"].(primitive.DateTime)
		rv := &entities.Review{
			ID:        uint64(doc["id"].(int64)),
			UserID:    uint64(doc["user_id"].(int64)),
			DatasetID: uint64(doc["dataset_id"].(int64)),
			Rating:    entities.Rating(doc["rating"].(int32)),
			Text:      doc["text"].(string),
			CreatedAt: created.Time(),
		}
		list = append(list, rv)
	}
	if err := cursor.Err(); err != nil {
		r.logger.Error("cursor error on find by dataset", zap.Error(err))
		return nil, repositories.ErrReviewNotFound
	}
	return list, nil
}

func (r *ReviewMongoRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Review, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		r.logger.Error("find reviews by user failed", zap.Error(err), zap.Uint64("user_id", userID))
		return nil, repositories.ErrReviewNotFound
	}
	defer cursor.Close(ctx)
	var list []*entities.Review
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Error("decode review failed", zap.Error(err))
			return nil, repositories.ErrReviewNotFound
		}
		created := doc["created_at"].(primitive.DateTime)
		rv := &entities.Review{
			ID:        uint64(doc["id"].(int64)),
			UserID:    uint64(doc["user_id"].(int64)),
			DatasetID: uint64(doc["dataset_id"].(int64)),
			Rating:    entities.Rating(doc["rating"].(int32)),
			Text:      doc["text"].(string),
			CreatedAt: created.Time(),
		}
		list = append(list, rv)
	}
	if err := cursor.Err(); err != nil {
		r.logger.Error("cursor error on find by user", zap.Error(err))
		return nil, repositories.ErrReviewNotFound
	}
	return list, nil
}
