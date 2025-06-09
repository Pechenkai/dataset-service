package postqbuild

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type UserMongoRepo struct {
	coll        *mongo.Collection
	counterColl *mongo.Collection
	logger      *zap.Logger
}

func NewUserMongoRepo(db *mongo.Database, logger *zap.Logger) *UserMongoRepo {
	_, _ = db.Collection("users").Indexes().CreateMany(
		context.Background(),
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{{Key: "registration_date", Value: -1}},
			},
		},
	)
	return &UserMongoRepo{
		coll:        db.Collection("users"),
		counterColl: db.Collection("counters"),
		logger:      logger,
	}
}

func (r *UserMongoRepo) getNextSeq(ctx context.Context, name string) (uint64, error) {
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

func (r *UserMongoRepo) Create(ctx context.Context, u *entities.User) error {
	if u.RegistrationDate.IsZero() {
		u.RegistrationDate = time.Now().UTC()
	}
	r.logger.Debug("Create User called",
		zap.String("username", u.Username),
		zap.String("email", u.Email),
	)

	id, err := r.getNextSeq(ctx, "users")
	if err != nil {
		return fmt.Errorf("create user seq: %w", err)
	}
	u.ID = id

	_, err = r.coll.InsertOne(ctx, u)
	if err != nil {
		var we mongo.WriteException
		if errors.As(err, &we) {
			for _, e := range we.WriteErrors {
				if e.Code == 11000 {
					r.logger.Warn("duplicate email on create user", zap.String("email", u.Email))
					return repositories.ErrEmailAlreadyExists
				}
			}
		}
		r.logger.Error("failed to insert user", zap.Error(err))
		return repositories.ErrUserCreate
	}
	r.logger.Info("user created successfully", zap.Uint64("id", u.ID), zap.String("email", u.Email))
	return nil
}

func (r *UserMongoRepo) Update(ctx context.Context, u *entities.User) error {
	r.logger.Debug("Update User called",
		zap.Uint64("id", u.ID),
		zap.String("email", u.Email),
	)

	update := bson.M{
		"$set": bson.M{
			"username":   u.Username,
			"email":      u.Email,
			"password":   u.Password,
			"country":    u.Country,
			"is_blocked": u.IsBlocked,
			"role":       u.Role,
		},
	}
	res, err := r.coll.UpdateOne(ctx, bson.M{"id": u.ID}, update)
	if err != nil {
		var we mongo.WriteException
		if errors.As(err, &we) {
			for _, e := range we.WriteErrors {
				if e.Code == 11000 {
					r.logger.Warn("duplicate email on update user", zap.String("email", u.Email))
					return repositories.ErrEmailAlreadyExists
				}
			}
		}
		r.logger.Error("failed to update user", zap.Error(err))
		return repositories.ErrUserUpdate
	}
	if res.MatchedCount == 0 {
		r.logger.Warn("no user found to update", zap.Uint64("id", u.ID))
		return repositories.ErrUserNotFound
	}
	r.logger.Info("user updated successfully", zap.Uint64("id", u.ID))
	return nil
}

func (r *UserMongoRepo) Delete(ctx context.Context, id uint64) error {
	r.logger.Debug("Delete User called", zap.Uint64("id", id))

	res, err := r.coll.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		r.logger.Error("failed to delete user", zap.Error(err))
		return repositories.ErrUserDelete
	}
	if res.DeletedCount == 0 {
		r.logger.Warn("no user found to delete", zap.Uint64("id", id))
		return repositories.ErrUserNotFound
	}
	r.logger.Info("user deleted successfully", zap.Uint64("id", id))
	return nil
}

func (r *UserMongoRepo) FindByID(ctx context.Context, id uint64) (*entities.User, error) {
	r.logger.Debug("FindByID User called", zap.Uint64("id", id))

	var u entities.User
	err := r.coll.FindOne(ctx, bson.M{"id": id}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		r.logger.Warn("user not found by ID", zap.Uint64("id", id))
		return nil, repositories.ErrUserNotFound
	}
	if err != nil {
		r.logger.Error("failed to find user by ID", zap.Error(err))
		return nil, repositories.ErrUserGet
	}
	r.logger.Info("user fetched successfully", zap.Uint64("id", u.ID), zap.String("email", u.Email))
	return &u, nil
}

func (r *UserMongoRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	r.logger.Debug("FindByEmail User called", zap.String("email", email))

	var u entities.User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		r.logger.Warn("user not found by email", zap.String("email", email))
		return nil, repositories.ErrUserNotFound
	}
	if err != nil {
		r.logger.Error("failed to find user by email", zap.Error(err))
		return nil, repositories.ErrUserGet
	}
	r.logger.Info("user fetched successfully by email", zap.Uint64("id", u.ID), zap.String("email", u.Email))
	return &u, nil
}

func (r *UserMongoRepo) FindAll(ctx context.Context) ([]*entities.User, error) {
	r.logger.Debug("FindAll Users called")

	opts := options.Find().SetSort(bson.D{{Key: "registration_date", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		r.logger.Error("failed to list all users", zap.Error(err))
		return nil, repositories.ErrUserGetAll
	}
	defer cursor.Close(ctx)

	var list []*entities.User
	for cursor.Next(ctx) {
		var u entities.User
		if err := cursor.Decode(&u); err != nil {
			r.logger.Error("failed to decode user", zap.Error(err))
			return nil, repositories.ErrUserScan
		}
		list = append(list, &u)
	}
	if cursor.Err() != nil {
		r.logger.Error("error iterating user rows", zap.Error(cursor.Err()))
		return nil, repositories.ErrUserGetAll
	}
	r.logger.Info("all users fetched successfully", zap.Int("count", len(list)))
	return list, nil
}
