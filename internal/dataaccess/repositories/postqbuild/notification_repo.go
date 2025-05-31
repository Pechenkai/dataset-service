package postqbuild

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type NotificationRepo struct {
	db *pgxpool.Pool
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{db: pool}
}

func (r *NotificationRepo) Create(ctx context.Context, n *entities.Notification) error {
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}

	query := psql.
		Insert("notifications").
		Columns("user_id", "dataset_id", "message", "is_read", "created_at").
		Values(n.UserID, n.DatasetID, n.Message, n.IsRead, n.CreatedAt).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return repositories.ErrNotificationQueryBuild
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&n.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return repositories.ErrNotificationCreate
		}
		return repositories.ErrNotificationCreate
	}
	return nil
}

func (r *NotificationRepo) FindByID(ctx context.Context, id uint64) (*entities.Notification, error) {
	query := psql.
		Select("id", "user_id", "dataset_id", "message", "is_read", "created_at").
		From("notifications").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, repositories.ErrNotificationQueryBuild
	}

	n := &entities.Notification{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
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
		return nil, repositories.ErrNotificationCreate
	}
	return n, nil
}

func (r *NotificationRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Notification, error) {
	query := psql.
		Select("id", "user_id", "dataset_id", "message", "is_read", "created_at").
		From("notifications").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, repositories.ErrNotificationQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", repositories.ErrNotificationUpdateFail, err.Error())
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
			return nil, repositories.ErrNotificationScanRow
		}
		list = append(list, n)
	}
	if err := rows.Err(); err != nil {
		return nil, repositories.ErrNotificationIterateRows
	}
	return list, nil
}

func (r *NotificationRepo) Update(ctx context.Context, n *entities.Notification) error {
	query := psql.
		Update("notifications").
		Set("message", n.Message).
		Set("is_read", n.IsRead).
		Where(sq.Eq{"id": n.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return repositories.ErrNotificationQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return repositories.ErrNotificationUpdateFail
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotificationNotFound
	}
	return nil
}

func (r *NotificationRepo) Delete(ctx context.Context, id uint64) error {
	query := psql.Delete("notifications").Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return repositories.ErrNotificationQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return repositories.ErrNotificationDeleteFail
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrNotificationNotFound
	}
	return nil
}
