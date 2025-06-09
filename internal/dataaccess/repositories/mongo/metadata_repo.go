package postqbuild

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type MetadataMongoRepo struct {
	coll        *mongo.Collection
	counterColl *mongo.Collection
	logger      *zap.Logger
}

func NewMetadataMongoRepo(db *mongo.Database, logger *zap.Logger) *MetadataMongoRepo {
	_, _ = db.Collection("metadata").Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "dataset_version_id", Value: 1}},
			Options: options.Index().SetName("idx_metadata_dataset_version"),
		},
	)
	// Индекс для счетчика
	_, _ = db.Collection("counters").Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "_id", Value: 1}},
			Options: options.Index().SetName("idx_counters_id"),
		},
	)
	return &MetadataMongoRepo{
		coll:        db.Collection("metadata"),
		counterColl: db.Collection("counters"),
		logger:      logger,
	}
}

// getNextSeq увеличивает и возвращает seq для ключа
func (r *MetadataMongoRepo) getNextSeq(ctx context.Context, name string) (uint64, error) {
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var res struct {
		Seq uint64 `bson:"seq"`
	}
	if err := r.counterColl.FindOneAndUpdate(ctx, filter, update, opts).Decode(&res); err != nil {
		r.logger.Error("counter inc failed", zap.String("name", name), zap.Error(err))
		return 0, err
	}
	return res.Seq, nil
}

func (r *MetadataMongoRepo) Create(ctx context.Context, m *entities.Metadata) error {
	id, err := r.getNextSeq(ctx, "metadata")
	if err != nil {
		return fmt.Errorf("create metadata seq: %w", err)
	}
	m.ID = id
	doc := bson.M{
		"id":                 m.ID,
		"format":             m.Format,
		"size":               m.Size,
		"tags":               m.Tags,
		"dataset_version_id": m.DatasetVersionID,
	}
	if _, err := r.coll.InsertOne(ctx, doc); err != nil {
		r.logger.Error("insert metadata failed", zap.Error(err))
		return repositories.ErrMetadataCreate
	}
	return nil
}

func (r *MetadataMongoRepo) Update(ctx context.Context, m *entities.Metadata) error {
	filter := bson.M{"id": m.ID}
	update := bson.M{"$set": bson.M{
		"format":             m.Format,
		"size":               m.Size,
		"tags":               m.Tags,
		"dataset_version_id": m.DatasetVersionID,
	}}
	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("update metadata failed", zap.Error(err), zap.Uint64("id", m.ID))
		return repositories.ErrMetadataUpdate
	}
	if res.MatchedCount == 0 {
		return repositories.ErrMetadataNotFound
	}
	return nil
}

func (r *MetadataMongoRepo) Delete(ctx context.Context, id uint64) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		r.logger.Error("delete metadata failed", zap.Error(err), zap.Uint64("id", id))
		return repositories.ErrMetadataDelete
	}
	if res.DeletedCount == 0 {
		return repositories.ErrMetadataNotFound
	}
	return nil
}

func (r *MetadataMongoRepo) FindByID(ctx context.Context, id uint64) (*entities.Metadata, error) {
	var doc bson.M
	err := r.coll.FindOne(ctx, bson.M{"id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, repositories.ErrMetadataNotFound
	}
	if err != nil {
		r.logger.Error("find metadata failed", zap.Error(err), zap.Uint64("id", id))
		return nil, repositories.ErrMetadataScan
	}
	return &entities.Metadata{
		ID:               uint64(doc["id"].(int64)),
		Format:           doc["format"].(string),
		Size:             uint64(doc["size"].(int64)),
		Tags:             doc["tags"].(string),
		DatasetVersionID: uint64(doc["dataset_version_id"].(int64)),
	}, nil
}

func (r *MetadataMongoRepo) FindByDatasetID(ctx context.Context, versionID uint64) ([]*entities.Metadata, error) {
	opts := options.Find().SetSort(bson.D{{Key: "id", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"dataset_version_id": versionID}, opts)
	if err != nil {
		r.logger.Error("list metadata failed", zap.Error(err), zap.Uint64("dataset_version_id", versionID))
		return nil, repositories.ErrMetadataList
	}
	defer cursor.Close(ctx)

	var out []*entities.Metadata
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Error("decode metadata failed", zap.Error(err))
			return nil, repositories.ErrMetadataScan
		}
		out = append(out, &entities.Metadata{
			ID:               uint64(doc["id"].(int64)),
			Format:           doc["format"].(string),
			Size:             uint64(doc["size"].(int64)),
			Tags:             doc["tags"].(string),
			DatasetVersionID: uint64(doc["dataset_version_id"].(int64)),
		})
	}
	if err := cursor.Err(); err != nil {
		r.logger.Error("cursor error metadata list", zap.Error(err))
		return nil, repositories.ErrMetadataList
	}
	return out, nil
}
