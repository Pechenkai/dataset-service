package postqbuild

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"ppo/internal/entities"
	"ppo/internal/repositories"
	"time"
)

type SubscriptionRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

type subscriptionRow struct {
	UserID    uint64
	DatasetID uint64
	CreatedAt time.Time
}

func subscriptionRowFromEntity(s *entities.Subscription) subscriptionRow {
	if s == nil {
		return subscriptionRow{}
	}
	return subscriptionRow{
		UserID:    s.UserID,
		DatasetID: s.DatasetID,
		CreatedAt: s.CreatedAt,
	}
}

func (row subscriptionRow) toEntity() *entities.Subscription {
	return &entities.Subscription{
		UserID:    row.UserID,
		DatasetID: row.DatasetID,
		CreatedAt: row.CreatedAt,
	}
}

func NewSubscriptionRepo(pool *pgxpool.Pool, logger *zap.Logger) *SubscriptionRepo {
	logger.Debug("NewSubscriptionRepo initialized")
	return &SubscriptionRepo{db: pool, logger: logger}
}

func (r *SubscriptionRepo) Create(ctx context.Context, s *entities.Subscription) error {
	row := subscriptionRowFromEntity(s)
	if row.CreatedAt.IsZero() {
		row.CreatedAt = time.Now().UTC()
		s.CreatedAt = row.CreatedAt
	}
	r.logger.Debug("Create Subscription called",
		zap.Uint64("user_id", row.UserID),
		zap.Uint64("dataset_id", row.DatasetID),
	)

	query := psql.
		Insert("subscriptions").
		Columns("user_id", "dataset_id", "created_at").
		Values(row.UserID, row.DatasetID, row.CreatedAt)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build insert subscription SQL",
			zap.Error(err),
		)
		return fmt.Errorf("build insert subscription sql: %w", err)
	}

	_, err = r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			r.logger.Warn("duplicate subscription on create",
				zap.Uint64("user_id", s.UserID),
				zap.Uint64("dataset_id", s.DatasetID),
				zap.String("constraint", pgErr.ConstraintName),
			)
			return repositories.ErrAlreadySubscribed
		}
		r.logger.Error("failed to execute insert subscription query",
			zap.Error(err),
			zap.Uint64("user_id", s.UserID),
			zap.Uint64("dataset_id", s.DatasetID),
		)
		return fmt.Errorf("subscribe: %w", err)
	}

	r.logger.Info("subscription created successfully",
		zap.Uint64("user_id", row.UserID),
		zap.Uint64("dataset_id", row.DatasetID),
	)
	return nil
}

func (r *SubscriptionRepo) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	r.logger.Debug("Unsubscribe called",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
	)

	query := psql.
		Delete("subscriptions").
		Where(sq.Eq{"user_id": userID, "dataset_id": datasetID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build delete subscription SQL",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return fmt.Errorf("build delete subscription sql: %w", err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute delete subscription query",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return fmt.Errorf("unsubscribe: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no subscription found to delete",
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return repositories.ErrSubscriptionNotFound
	}

	r.logger.Info("subscription deleted successfully",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
	)
	return nil
}

func (r *SubscriptionRepo) IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error) {
	r.logger.Debug("IsSubscribed called",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
	)

	query := psql.
		Select("1").
		From("subscriptions").
		Where(sq.Eq{"user_id": userID, "dataset_id": datasetID}).
		Limit(1)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build is_subscribed SQL",
			zap.Error(err),
		)
		return false, fmt.Errorf("build is_subscribed sql: %w", err)
	}

	var dummy int
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Info("user is not subscribed",
				zap.Uint64("user_id", userID),
				zap.Uint64("dataset_id", datasetID),
			)
			return false, nil
		}
		r.logger.Error("error executing is_subscribed query",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return false, fmt.Errorf("is subscribed: %w", err)
	}

	r.logger.Info("user is subscribed",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
	)
	return true, nil
}

func (r *SubscriptionRepo) GetSubscribers(ctx context.Context, datasetID uint64) ([]*entities.Subscription, error) {
	r.logger.Debug("GetSubscribers called",
		zap.Uint64("dataset_id", datasetID),
	)

	query := psql.
		Select("user_id", "dataset_id", "created_at").
		From("subscriptions").
		Where(sq.Eq{"dataset_id": datasetID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build get subscribers SQL",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, fmt.Errorf("build get subscribers sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute get subscribers query",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, fmt.Errorf("get subscribers: %w", err)
	}
	defer rows.Close()

	var list []*entities.Subscription
	for rows.Next() {
		row := subscriptionRow{}
		if err := rows.Scan(&row.UserID, &row.DatasetID, &row.CreatedAt); err != nil {
			r.logger.Error("failed to scan subscriber row",
				zap.Error(err),
			)
			return nil, fmt.Errorf("scan subscriber row: %w", err)
		}
		list = append(list, row.toEntity())
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating subscriber rows",
			zap.Error(err),
		)
		return nil, fmt.Errorf("iterate subscriber rows: %w", err)
	}

	r.logger.Info("subscribers fetched successfully",
		zap.Uint64("dataset_id", datasetID),
		zap.Int("count", len(list)),
	)
	return list, nil
}

func (r *SubscriptionRepo) GetByUser(ctx context.Context, userID uint64) ([]*entities.Subscription, error) {
	r.logger.Debug("GetByUser called",
		zap.Uint64("user_id", userID),
	)

	query := psql.
		Select("user_id", "dataset_id", "created_at").
		From("subscriptions").
		Where(sq.Eq{"user_id": userID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build get subscriptions by user SQL",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("build get subscriptions by user sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute get subscriptions by user query",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("get_subscriptions_by_user: %w", err)
	}
	defer rows.Close()

	var datasets []*entities.Subscription
	for rows.Next() {
		row := subscriptionRow{}
		if err := rows.Scan(&row.UserID, &row.DatasetID, &row.CreatedAt); err != nil {
			r.logger.Error("failed to scan subscription row",
				zap.Error(err),
			)
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		datasets = append(datasets, row.toEntity())
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating subscription rows",
			zap.Error(err),
		)
		return nil, fmt.Errorf("iterate subscription rows: %w", err)
	}

	r.logger.Info("subscriptions fetched by user successfully",
		zap.Uint64("user_id", userID),
		zap.Int("count", len(datasets)),
	)
	return datasets, nil
}

func (r *SubscriptionRepo) Delete(ctx context.Context, userID, datasetID uint64) error {
	r.logger.Debug("Delete Subscription called",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
	)

	query := psql.Delete("subscriptions").Where(sq.Eq{
		"user_id":    userID,
		"dataset_id": datasetID,
	})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build delete subscription SQL",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return fmt.Errorf("build delete subscription sql: %w", err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute delete subscription query",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return fmt.Errorf("delete subscription: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no subscription found to delete",
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return repositories.ErrSubscriptionNotFound
	}

	r.logger.Info("subscription deleted successfully",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
	)
	return nil
}

func (r *SubscriptionRepo) GetAll(ctx context.Context) ([]*entities.Subscription, error) {
	r.logger.Debug("GetAll subscriptions called")

	query := psql.
		Select("user_id", "dataset_id", "created_at").
		From("subscriptions").
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build get all subscriptions SQL",
			zap.Error(err),
		)
		return nil, fmt.Errorf("build get all subscriptions sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute get all subscriptions query",
			zap.Error(err),
		)
		return nil, fmt.Errorf("get all subscriptions: %w", err)
	}
	defer rows.Close()

	var result []*entities.Subscription
	for rows.Next() {
		row := subscriptionRow{}
		if err := rows.Scan(&row.UserID, &row.DatasetID, &row.CreatedAt); err != nil {
			r.logger.Error("failed to scan subscription row", zap.Error(err))
			return nil, fmt.Errorf("scan subscription row: %w", err)
		}
		result = append(result, row.toEntity())
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating subscription rows", zap.Error(err))
		return nil, fmt.Errorf("iterate subscription rows: %w", err)
	}

	r.logger.Info("all subscriptions fetched successfully", zap.Int("count", len(result)))
	return result, nil
}
