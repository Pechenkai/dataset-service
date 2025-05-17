package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepo(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{db: pool}
}

func (r *SubscriptionRepo) Subscribe(ctx context.Context, userID, datasetID uint64) error {
	const sql = `
	INSERT INTO subscriptions
	  (user_id, dataset_id, created_at)
	VALUES ($1, $2, $3)
	`
	t := time.Now().UTC()
	_, err := r.db.Exec(ctx, sql, userID, datasetID, t)
	if err != nil {
		// уникальное ограничение (user_id, dataset_id)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadySubscribed
		}
		return fmt.Errorf("subscribe: %w", err)
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
