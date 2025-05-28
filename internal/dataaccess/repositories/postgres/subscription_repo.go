package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/entities"
)

type SubscriptionRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepo(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{db: pool}
}

func (r *SubscriptionRepo) Create(ctx context.Context, s *entities.Subscription) error {
	const sql = `
	INSERT INTO subscriptions
	  (user_id, dataset_id, created_at)
	VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, sql, s.UserID, s.DatasetID, s.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadySubscribed
		}
		return fmt.Errorf("subscribe: %w", err)
	}
	return nil
}

func (r *SubscriptionRepo) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	const sql = `DELETE FROM subscriptions WHERE user_id = $1 AND dataset_id = $2`

	cmdTag, err := r.db.Exec(ctx, sql, userID, datasetID)
	if err != nil {
		return fmt.Errorf("unsubscribe: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}
	return nil
}

func (r *SubscriptionRepo) IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error) {
	const sql = `
	SELECT 1 FROM subscriptions
	WHERE user_id = $1 AND dataset_id = $2
	LIMIT 1
	`
	row := r.db.QueryRow(ctx, sql, userID, datasetID)
	var dummy int
	err := row.Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("is subscribed: %w", err)
	}
	return true, nil
}

func (r *SubscriptionRepo) GetSubscribers(ctx context.Context, datasetID uint64) ([]uint64, error) {
	const sql = `
	SELECT user_id FROM subscriptions
	WHERE dataset_id = $1
	`
	rows, err := r.db.Query(ctx, sql, datasetID)
	if err != nil {
		return nil, fmt.Errorf("get subscribers: %w", err)
	}
	defer rows.Close()

	var list []uint64
	for rows.Next() {
		var uid uint64
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan subscriber row: %w", err)
		}
		list = append(list, uid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriber rows: %w", err)
	}
	return list, nil
}

func (r *SubscriptionRepo) GetByUser(ctx context.Context, userID uint64) ([]uint64, error) {
	rows, err := r.db.Query(ctx,
		`SELECT dataset_id FROM subscriptions WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("get_subscriptions_by_user: %w", err)
	}
	defer rows.Close()

	var datasets []uint64
	for rows.Next() {
		var dsid uint64
		if err := rows.Scan(&dsid); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		datasets = append(datasets, dsid)
	}
	return datasets, rows.Err()
}
