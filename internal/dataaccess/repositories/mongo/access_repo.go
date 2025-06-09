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

type AccessRequestMongoRepo struct {
	coll        *mongo.Collection
	counterColl *mongo.Collection
	logger      *zap.Logger
}

func NewAccessRequestMongoRepo(db *mongo.Database, logger *zap.Logger) *AccessRequestMongoRepo {
	logger.Debug("NewAccessRequestMongoRepo initialized")
	return &AccessRequestMongoRepo{
		coll:        db.Collection("access_requests"),
		counterColl: db.Collection("counters"),
		logger:      logger,
	}
}

func (r *AccessRequestMongoRepo) getNextSequence(ctx context.Context, name string) (uint64, error) {
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result struct {
		Seq uint64 `bson:"seq"`
	}
	if err := r.counterColl.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result); err != nil {
		r.logger.Error("failed to get next sequence", zap.String("name", name), zap.Error(err))
		return 0, err
	}
	return result.Seq, nil
}

func (r *AccessRequestMongoRepo) Create(ctx context.Context, ar *entities.AccessRequest) error {
	r.logger.Debug("Create AccessRequest called",
		zap.Uint64("dataset_id", ar.DatasetID),
		zap.Uint64("user_id", ar.UserID),
		zap.String("status", string(ar.Status)),
	)

	id, err := r.getNextSequence(ctx, "access_requests")
	if err != nil {
		return fmt.Errorf("create access_request: seq error: %w", err)
	}
	ar.ID = id

	if _, err := r.coll.InsertOne(ctx, ar); err != nil {
		r.logger.Error("failed to insert access_request", zap.Error(err))
		return fmt.Errorf("create access_request: %w", err)
	}
	r.logger.Info("AccessRequest created", zap.Uint64("id", ar.ID))
	return nil
}

func (r *AccessRequestMongoRepo) Find(ctx context.Context, datasetID, userID uint64) (*entities.AccessRequest, error) {
	r.logger.Debug("Find AccessRequest called",
		zap.Uint64("dataset_id", datasetID),
		zap.Uint64("user_id", userID),
	)

	filter := bson.M{"dataset_id": datasetID, "user_id": userID}
	var ar entities.AccessRequest
	err := r.coll.FindOne(ctx, filter).Decode(&ar)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Debug("AccessRequest not found", zap.Uint64("dataset_id", datasetID), zap.Uint64("user_id", userID))
			return nil, nil
		}
		r.logger.Error("failed to execute Find", zap.Error(err))
		return nil, fmt.Errorf("find access_request: %w", err)
	}
	r.logger.Info("AccessRequest fetched", zap.Uint64("id", ar.ID))
	return &ar, nil
}

func (r *AccessRequestMongoRepo) ListPendingByOwner(ctx context.Context, ownerID uint64) ([]*entities.AccessRequest, error) {
	r.logger.Debug("ListPendingByOwner called", zap.Uint64("owner_id", ownerID))

	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "datasets"},
			{Key: "localField", Value: "dataset_id"},
			{Key: "foreignField", Value: "id"},
			{Key: "as", Value: "dataset"},
		}}},
		{{Key: "$unwind", Value: "$dataset"}},
		{{Key: "$match", Value: bson.D{
			{Key: "dataset.owner_id", Value: ownerID},
			{Key: "status", Value: string(entities.AccessStatusPending)},
		}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error("aggregate ListPendingByOwner failed", zap.Error(err))
		return nil, fmt.Errorf("list pending: %w", err)
	}
	defer cursor.Close(ctx)

	var list []*entities.AccessRequest
	for cursor.Next(ctx) {
		var ar entities.AccessRequest
		if err := cursor.Decode(&ar); err != nil {
			r.logger.Error("failed to decode row", zap.Error(err))
			return nil, repositories.ErrRequestScan
		}
		list = append(list, &ar)
	}
	if err := cursor.Err(); err != nil {
		r.logger.Error("cursor error", zap.Error(err))
		return nil, repositories.ErrRequestScan
	}

	r.logger.Info("ListPendingByOwner completed",
		zap.Uint64("owner_id", ownerID),
		zap.Int("count", len(list)),
	)
	return list, nil
}

func (r *AccessRequestMongoRepo) UpdateStatus(ctx context.Context, id uint64, status string) error {
	r.logger.Debug("UpdateStatus called", zap.Uint64("id", id), zap.String("status", status))

	res, err := r.coll.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{"$set": bson.M{"status": status}},
	)
	if err != nil {
		r.logger.Error("failed to update status", zap.Error(err))
		return fmt.Errorf("update status: %w", err)
	}
	if res.MatchedCount == 0 {
		r.logger.Warn("no request found to update", zap.Uint64("id", id))
		return repositories.ErrRequestNotFound
	}
	r.logger.Info("UpdateStatus completed", zap.Uint64("id", id), zap.String("status", status))
	return nil
}

func (r *AccessRequestMongoRepo) FindByRequestID(ctx context.Context, requestID uint64) (*entities.AccessRequest, error) {
	r.logger.Debug("FindByRequestID called", zap.Uint64("request_id", requestID))

	var ar entities.AccessRequest
	err := r.coll.FindOne(ctx, bson.M{"id": requestID}).Decode(&ar)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.logger.Info("access request not found by ID", zap.Uint64("request_id", requestID))
			return nil, repositories.ErrRequestNotFound
		}
		r.logger.Error("failed to execute FindByRequestID", zap.Error(err))
		return nil, fmt.Errorf("find access_request by id: %w", err)
	}

	r.logger.Info("AccessRequest fetched by ID", zap.Uint64("request_id", ar.ID))
	return &ar, nil
}
