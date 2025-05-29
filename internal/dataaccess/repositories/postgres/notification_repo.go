package postqbuild

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/repositories"

	"ppo/internal/entities"
)

type NotificationRepo struct {
	db *pgxpool.Pool
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{db: pool}
}

func (r *NotificationRepo) Create(ctx context.Context, n *entities.Notification) error {
	const sql = `
	INSERT INTO notifications
	  (user_id, dataset_id, message, is_read, created_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
	`
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	err := r.db.QueryRow(ctx, sql,
		n.UserID,
		n.DatasetID,
		n.Message,
		n.IsRead,
		n.CreatedAt,
	).Scan(&n.ID)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (r *NotificationRepo) FindByID(ctx context.Context, id uint64) (*entities.Notification, error) {
	const sql = `
	SELECT id, user_id, dataset_id, message, is_read, created_at
	FROM notifications
	WHERE id = $1
	`
	n := &entities.Notification{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&n.ID,
		&n.UserID,
		&n.DatasetID,
		&n.Message,
		&n.IsRead,
		&n.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("find notification by id: %w", err)
	}
	return n, nil
}

func (r *NotificationRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Notification, error) {
	const sql = `
	SELECT id, user_id, dataset_id, message, is_read, created_at
	FROM notifications
	WHERE user_id = $1
	ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("query notifications by user id: %w", err)
	}
	defer rows.Close()

	var list []*entities.Notification
	for rows.Next() {
		n := &entities.Notification{}
		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.DatasetID,
			&n.Message,
			&n.IsRead,
			&n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification row: %w", err)
		}
		list = append(list, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notification rows: %w", err)
	}
	return list, nil
}

func (r *NotificationRepo) Update(ctx context.Context, n *entities.Notification) error {
	const sql = `
	UPDATE notifications
	SET message = $1, is_read = $2
	WHERE id = $3
	`
	cmd, err := r.db.Exec(ctx, sql,
		n.Message,
		n.IsRead,
		n.ID,
	)
	if err != nil {
		return fmt.Errorf("update notification: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotificationNotFound
	}
	return nil
}

func (r *NotificationRepo) Delete(ctx context.Context, id uint64) error {
	const sql = `DELETE FROM notifications WHERE id = $1`
	exec, err := r.db.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	if exec.RowsAffected() == 0 {
		return repositories.ErrNotificationNotFound
	}
	return nil
}
