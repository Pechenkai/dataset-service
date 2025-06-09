package postqbuild

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type CategoryMongoRepo struct {
	coll        *mongo.Collection
	counterColl *mongo.Collection
	datasetColl *mongo.Collection
	logger      *zap.Logger
}

func NewCategoryMongoRepo(db *mongo.Database, logger *zap.Logger) *CategoryMongoRepo {
	_, _ = db.Collection("categories").Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &CategoryMongoRepo{
		coll:        db.Collection("categories"),
		counterColl: db.Collection("counters"),
		datasetColl: db.Collection("datasets"),
		logger:      logger,
	}
}

func (r *CategoryMongoRepo) getNextSeq(ctx context.Context, name string) (uint64, error) {
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var res struct {
		Seq uint64 `bson:"seq"`
	}
	if err := r.counterColl.FindOneAndUpdate(ctx, filter, update, opts).Decode(&res); err != nil {
		r.logger.Error("counter increment failed", zap.String("name", name), zap.Error(err))
		return 0, err
	}
	return res.Seq, nil
}

func (r *CategoryMongoRepo) Create(ctx context.Context, c *entities.Category) error {
	r.logger.Debug("Create Category", zap.String("name", c.Name))

	id, err := r.getNextSeq(ctx, "categories")
	if err != nil {
		return fmt.Errorf("create category seq: %w", err)
	}
	c.ID = id

	_, err = r.coll.InsertOne(ctx, c)
	if err != nil {
		var we mongo.WriteException
		if errors.As(err, &we) {
			for _, e := range we.WriteErrors {
				if e.Code == 11000 {
					r.logger.Warn("duplicate category name", zap.String("name", c.Name))
					return repositories.ErrCategoryAlreadyExists
				}
			}
		}
		r.logger.Error("insert failed", zap.Error(err))
		return fmt.Errorf("%w: %v", repositories.ErrCategoryCreate, err)
	}
	r.logger.Info("category created", zap.Uint64("id", c.ID))
	return nil
}

func (r *CategoryMongoRepo) Update(ctx context.Context, c *entities.Category) error {
	r.logger.Debug("Update Category", zap.Uint64("id", c.ID))

	res, err := r.coll.UpdateOne(
		ctx,
		bson.M{"id": c.ID},
		bson.M{"$set": bson.M{"name": c.Name, "description": c.Description}},
	)
	if err != nil {
		var we mongo.WriteException
		if errors.As(err, &we) {
			for _, e := range we.WriteErrors {
				if e.Code == 11000 {
					r.logger.Warn("duplicate category name on update", zap.String("name", c.Name))
					return repositories.ErrCategoryAlreadyExists
				}
			}
		}
		r.logger.Error("update failed", zap.Error(err))
		return fmt.Errorf("%w: %v", repositories.ErrCategoryUpdate, err)
	}
	if res.MatchedCount == 0 {
		r.logger.Warn("no category found to update", zap.Uint64("id", c.ID))
		return repositories.ErrCategoryNotFound
	}
	r.logger.Info("category updated", zap.Uint64("id", c.ID))
	return nil
}

func (r *CategoryMongoRepo) Delete(ctx context.Context, id uint64) error {
	r.logger.Debug("Delete Category called", zap.Uint64("id", id))
	// проверяем “непустоту”: есть ли данные, ссылающиеся на эту категорию
	cnt, err := r.datasetColl.CountDocuments(ctx, bson.M{"category_id": id})
	if err != nil {
		r.logger.Error("failed to count datasets for category", zap.Uint64("id", id), zap.Error(err))
		return fmt.Errorf("%w: %v", repositories.ErrCategoryDelete, err)
	}
	if cnt > 0 {
		r.logger.Warn("category not empty, cannot delete", zap.Uint64("id", id))
		return repositories.ErrCategoryNotEmpty
	}
	// собственно удаление
	res, err := r.coll.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		r.logger.Error("failed to delete category", zap.Uint64("id", id), zap.Error(err))
		return fmt.Errorf("%w: %v", repositories.ErrCategoryDelete, err)
	}
	if res.DeletedCount == 0 {
		r.logger.Warn("no category found to delete", zap.Uint64("id", id))
		return repositories.ErrCategoryNotFound
	}
	r.logger.Info("category deleted", zap.Uint64("id", id))
	return nil
}

func (r *CategoryMongoRepo) FindByID(ctx context.Context, id uint64) (*entities.Category, error) {
	r.logger.Debug("FindByID Category called", zap.Uint64("id", id))
	var c entities.Category
	err := r.coll.FindOne(ctx, bson.M{"id": id}).Decode(&c)
	if err == mongo.ErrNoDocuments {
		r.logger.Warn("category not found", zap.Uint64("id", id))
		return nil, repositories.ErrCategoryNotFound
	}
	if err != nil {
		r.logger.Error("failed to find category by ID", zap.Uint64("id", id), zap.Error(err))
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryFind, err)
	}
	r.logger.Info("category fetched", zap.Uint64("id", c.ID))
	return &c, nil
}

func (r *CategoryMongoRepo) FindAll(ctx context.Context) ([]*entities.Category, error) {
	r.logger.Debug("FindAll Categories called")
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		r.logger.Error("failed to list categories", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryFind, err)
	}
	defer cursor.Close(ctx)

	var list []*entities.Category
	for cursor.Next(ctx) {
		var c entities.Category
		if err := cursor.Decode(&c); err != nil {
			r.logger.Error("failed to decode category", zap.Error(err))
			return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryScan, err)
		}
		list = append(list, &c)
	}
	if err := cursor.Err(); err != nil {
		r.logger.Error("cursor error listing categories", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryIterate, err)
	}
	r.logger.Info("categories listed", zap.Int("count", len(list)))
	return list, nil
}
