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

type DatasetMongoRepo struct {
	coll        *mongo.Collection
	counterColl *mongo.Collection
	logger      *zap.Logger
}

func NewDatasetMongoRepo(db *mongo.Database, logger *zap.Logger) *DatasetMongoRepo {
	_, _ = db.Collection("datasets").Indexes().CreateMany(
		context.Background(),
		[]mongo.IndexModel{
			{Keys: bson.D{{Key: "owner_id", Value: 1}}},
			{Keys: bson.D{{Key: "category_id", Value: 1}}},
			{Keys: bson.D{{Key: "is_public", Value: 1}}},
			{Keys: bson.D{{Key: "created_at", Value: -1}}},
		},
	)
	return &DatasetMongoRepo{
		coll:        db.Collection("datasets"),
		counterColl: db.Collection("counters"),
		logger:      logger,
	}
}

func (r *DatasetMongoRepo) getNextSeq(ctx context.Context, name string) (uint64, error) {
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var doc struct {
		Seq uint64 `bson:"seq"`
	}
	if err := r.counterColl.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc); err != nil {
		r.logger.Error("counter increment failed", zap.String("name", name), zap.Error(err))
		return 0, err
	}
	return doc.Seq, nil
}

func (r *DatasetMongoRepo) Create(ctx context.Context, d *entities.Dataset) error {
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}
	r.logger.Debug("Create Dataset called",
		zap.String("name", d.Name),
		zap.Uint64("owner_id", d.OwnerID),
	)

	seq, err := r.getNextSeq(ctx, "datasets")
	if err != nil {
		return fmt.Errorf("create dataset seq: %w", err)
	}
	d.ID = seq

	doc := bson.M{
		"id":          d.ID,
		"name":        d.Name,
		"description": d.Description,
		"owner_id":    d.OwnerID,
		"category_id": d.CategoryID,
		"is_public":   d.IsPublic,
		"created_at":  primitive.NewDateTimeFromTime(d.CreatedAt),
	}
	if _, err := r.coll.InsertOne(ctx, doc); err != nil {
		r.logger.Error("failed to insert dataset", zap.Error(err))
		return repositories.ErrDatasetCreate
	}
	r.logger.Info("dataset created", zap.Uint64("id", d.ID))
	return nil
}

func (r *DatasetMongoRepo) FindByID(ctx context.Context, id uint64) (*entities.Dataset, error) {
	r.logger.Debug("FindByID Dataset called", zap.Uint64("id", id))
	var doc bson.M
	err := r.coll.FindOne(ctx, bson.M{"id": id}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, repositories.ErrDatasetNotFound
	}
	if err != nil {
		r.logger.Error("find by id error", zap.Error(err))
		return nil, repositories.ErrDatasetScan
	}

	created, ok := doc["created_at"].(primitive.DateTime)
	if !ok {
		return nil, fmt.Errorf("unexpected type for created_at: %T", doc["created_at"])
	}

	return &entities.Dataset{
		ID:          uint64(doc["id"].(int64)),
		Name:        doc["name"].(string),
		Description: doc["description"].(string),
		OwnerID:     uint64(doc["owner_id"].(int64)),
		CategoryID:  uint64(doc["category_id"].(int64)),
		IsPublic:    doc["is_public"].(bool),
		CreatedAt:   created.Time(),
	}, nil
}

func (r *DatasetMongoRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Dataset, error) {
	r.logger.Debug("FindByUserID Datasets called", zap.Uint64("user_id", userID))
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"owner_id": userID}, opts)
	if err != nil {
		r.logger.Error("find by user error", zap.Error(err))
		return nil, repositories.ErrDatasetScan
	}
	defer cursor.Close(ctx)

	var list []*entities.Dataset
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Error("decode dataset error", zap.Error(err))
			return nil, repositories.ErrDatasetScan
		}
		created := doc["created_at"].(primitive.DateTime)
		list = append(list, &entities.Dataset{
			ID:          uint64(doc["id"].(int64)),
			Name:        doc["name"].(string),
			Description: doc["description"].(string),
			OwnerID:     uint64(doc["owner_id"].(int64)),
			CategoryID:  uint64(doc["category_id"].(int64)),
			IsPublic:    doc["is_public"].(bool),
			CreatedAt:   created.Time(),
		})
	}
	return list, nil
}

func (r *DatasetMongoRepo) FindAll(ctx context.Context) ([]*entities.Dataset, error) {
	r.logger.Debug("FindAll Datasets called")
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		r.logger.Error("find all error", zap.Error(err))
		return nil, repositories.ErrDatasetList
	}
	defer cursor.Close(ctx)

	var list []*entities.Dataset
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Error("decode dataset error", zap.Error(err))
			return nil, repositories.ErrDatasetScan
		}
		created := doc["created_at"].(primitive.DateTime)
		list = append(list, &entities.Dataset{
			ID:          uint64(doc["id"].(int64)),
			Name:        doc["name"].(string),
			Description: doc["description"].(string),
			OwnerID:     uint64(doc["owner_id"].(int64)),
			CategoryID:  uint64(doc["category_id"].(int64)),
			IsPublic:    doc["is_public"].(bool),
			CreatedAt:   created.Time(),
		})
	}
	return list, nil
}

func (r *DatasetMongoRepo) FindPublic(ctx context.Context) ([]*entities.Dataset, error) {
	r.logger.Debug("FindPublic Datasets called")
	cursor, err := r.coll.Find(ctx, bson.M{"is_public": true})
	if err != nil {
		r.logger.Error("find public error", zap.Error(err))
		return nil, repositories.ErrDatasetScan
	}
	defer cursor.Close(ctx)

	var list []*entities.Dataset
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Error("decode dataset error", zap.Error(err))
			return nil, repositories.ErrDatasetScan
		}
		created := doc["created_at"].(primitive.DateTime)
		list = append(list, &entities.Dataset{
			ID:          uint64(doc["id"].(int64)),
			Name:        doc["name"].(string),
			Description: doc["description"].(string),
			OwnerID:     uint64(doc["owner_id"].(int64)),
			CategoryID:  uint64(doc["category_id"].(int64)),
			IsPublic:    doc["is_public"].(bool),
			CreatedAt:   created.Time(),
		})
	}
	return list, nil
}

func (r *DatasetMongoRepo) FindByCategoryID(ctx context.Context, categoryID uint64) ([]*entities.Dataset, error) {
	r.logger.Debug("FindByCategoryID Datasets called", zap.Uint64("category_id", categoryID))
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"category_id": categoryID}, opts)
	if err != nil {
		r.logger.Error("find by category error", zap.Error(err))
		return nil, repositories.ErrDatasetScan
	}
	defer cursor.Close(ctx)

	var list []*entities.Dataset
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Error("decode dataset error", zap.Error(err))
			return nil, repositories.ErrDatasetScan
		}
		created := doc["created_at"].(primitive.DateTime)
		list = append(list, &entities.Dataset{
			ID:          uint64(doc["id"].(int64)),
			Name:        doc["name"].(string),
			Description: doc["description"].(string),
			OwnerID:     uint64(doc["owner_id"].(int64)),
			CategoryID:  uint64(doc["category_id"].(int64)),
			IsPublic:    doc["is_public"].(bool),
			CreatedAt:   created.Time(),
		})
	}
	return list, nil
}

func (r *DatasetMongoRepo) Update(ctx context.Context, d *entities.Dataset) error {
	r.logger.Debug("Update Dataset called", zap.Uint64("id", d.ID))
	update := bson.M{"$set": bson.M{
		"name":        d.Name,
		"description": d.Description,
		"category_id": d.CategoryID,
		"is_public":   d.IsPublic,
	}}
	res, err := r.coll.UpdateOne(ctx, bson.M{"id": d.ID}, update)
	if err != nil {
		r.logger.Error("update error", zap.Error(err))
		return repositories.ErrDatasetUpdate
	}
	if res.MatchedCount == 0 {
		return repositories.ErrDatasetNotFound
	}
	return nil
}

func (r *DatasetMongoRepo) Delete(ctx context.Context, id uint64) error {
	r.logger.Debug("Delete Dataset called", zap.Uint64("id", id))
	res, err := r.coll.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		r.logger.Error("delete error", zap.Error(err))
		return repositories.ErrDatasetDelete
	}
	if res.DeletedCount == 0 {
		return repositories.ErrDatasetNotFound
	}
	return nil
}
