package postqbuild

import (
	"context"
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

type VersionMongoRepo struct {
	coll        *mongo.Collection
	counterColl *mongo.Collection
	logger      *zap.Logger
}

func NewVersionMongoRepo(db *mongo.Database, logger *zap.Logger) *VersionMongoRepo {
	_, _ = db.Collection("dataset_versions").Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{Keys: bson.D{{Key: "upload_date", Value: -1}}},
	)
	return &VersionMongoRepo{
		coll:        db.Collection("dataset_versions"),
		counterColl: db.Collection("counters"),
		logger:      logger,
	}
}

func (r *VersionMongoRepo) getNextSeq(ctx context.Context, name string) (uint64, error) {
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

func (r *VersionMongoRepo) Create(ctx context.Context, v *entities.DatasetVersion) error {
	if v.UploadDate.IsZero() {
		v.UploadDate = time.Now().UTC()
	}
	seq, err := r.getNextSeq(ctx, "dataset_versions")
	if err != nil {
		return fmt.Errorf("create version seq: %w", err)
	}
	v.ID = seq
	// Вставляем документ вручную
	doc := bson.M{
		"id":          v.ID,
		"number":      v.Number,
		"upload_date": primitive.NewDateTimeFromTime(v.UploadDate),
		"filepath":    v.Filepath,
		"dataset_id":  v.DatasetID,
		"change_log":  v.ChangeLog,
	}
	if _, err := r.coll.InsertOne(ctx, doc); err != nil {
		r.logger.Error("insert version failed", zap.Error(err))
		return repositories.ErrVersionCreate
	}
	return nil
}

func (r *VersionMongoRepo) Update(ctx context.Context, v *entities.DatasetVersion) error {
	if v.UploadDate.IsZero() {
		v.UploadDate = time.Now().UTC()
	}
	filter := bson.M{"id": v.ID}
	update := bson.M{"$set": bson.M{
		"number":      v.Number,
		"upload_date": primitive.NewDateTimeFromTime(v.UploadDate),
		"filepath":    v.Filepath,
		"change_log":  v.ChangeLog,
	}}
	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("update version failed", zap.Error(err), zap.Uint64("id", v.ID))
		return repositories.ErrVersionUpdate
	}
	if res.MatchedCount == 0 {
		return repositories.ErrVersionNotFound
	}
	return nil
}

func (r *VersionMongoRepo) Delete(ctx context.Context, id uint64) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		r.logger.Error("delete version failed", zap.Error(err), zap.Uint64("id", id))
		return repositories.ErrVersionDelete
	}
	if res.DeletedCount == 0 {
		return repositories.ErrVersionNotFound
	}
	return nil
}

func (r *VersionMongoRepo) FindByID(ctx context.Context, id uint64) (*entities.DatasetVersion, error) {
	var doc bson.M
	err := r.coll.FindOne(ctx, bson.M{"id": id}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, repositories.ErrVersionNotFound
	}
	if err != nil {
		r.logger.Error("find version by id failed", zap.Error(err), zap.Uint64("id", id))
		return nil, repositories.ErrVersionScan
	}
	created := doc["upload_date"].(primitive.DateTime)
	v := &entities.DatasetVersion{
		ID:         uint64(doc["id"].(int64)),
		Number:     doc["number"].(string),
		UploadDate: created.Time(),
		Filepath:   doc["filepath"].(string),
		DatasetID:  uint64(doc["dataset_id"].(int64)),
		ChangeLog:  doc["change_log"].(string),
	}
	return v, nil
}

func (r *VersionMongoRepo) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.DatasetVersion, error) {
	opts := options.Find().SetSort(bson.D{{Key: "upload_date", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"dataset_id": datasetID}, opts)
	if err != nil {
		r.logger.Error("find versions by dataset failed", zap.Error(err), zap.Uint64("dataset_id", datasetID))
		return nil, repositories.ErrVersionList
	}
	defer cursor.Close(ctx)
	var list []*entities.DatasetVersion
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			r.logger.Error("decode version failed", zap.Error(err))
			return nil, repositories.ErrVersionScan
		}
		created := doc["upload_date"].(primitive.DateTime)
		v := &entities.DatasetVersion{
			ID:         uint64(doc["id"].(int64)),
			Number:     doc["number"].(string),
			UploadDate: created.Time(),
			Filepath:   doc["filepath"].(string),
			DatasetID:  uint64(doc["dataset_id"].(int64)),
			ChangeLog:  doc["change_log"].(string),
		}
		list = append(list, v)
	}
	if err := cursor.Err(); err != nil {
		r.logger.Error("cursor error on find versions", zap.Error(err), zap.Uint64("dataset_id", datasetID))
		return nil, repositories.ErrVersionList
	}
	return list, nil
}
