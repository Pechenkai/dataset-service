package postqbuild

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type NotificationRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewNotificationRepo(pool *pgxpool.Pool, logger *zap.Logger) *NotificationRepo {
	logger.Debug("NewNotificationRepo initialized")
	return &NotificationRepo{db: pool, logger: logger}
}

func (r *NotificationRepo) Create(ctx context.Context, n *entities.Notification) error {
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	r.logger.Debug("Create Notification called",
		zap.Uint64("user_id", n.UserID),
		zap.Uint64("dataset_id", n.DatasetID),
		zap.String("message", n.Message),
		zap.Bool("is_read", n.IsRead),
	)

	query := psql.
		Insert("notifications").
		Columns("user_id", "dataset_id", "message", "is_read", "created_at").
		Values(n.UserID, n.DatasetID, n.Message, n.IsRead, n.CreatedAt).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build create notification query",
			zap.Error(err),
		)
		return repositories.ErrNotificationQueryBuild
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&n.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.logger.Error("failed to execute create notification query (PG error)",
				zap.Error(err),
			)
			return repositories.ErrNotificationCreate
		}
		r.logger.Error("failed to execute create notification query",
			zap.Error(err),
		)
		return repositories.ErrNotificationCreate
	}

	r.logger.Info("notification created successfully",
		zap.Uint64("id", n.ID),
		zap.Uint64("user_id", n.UserID),
		zap.Uint64("dataset_id", n.DatasetID),
	)
	return nil
}

func (r *NotificationRepo) FindByID(ctx context.Context, id uint64) (*entities.Notification, error) {
	r.logger.Debug("FindByID Notification called",
		zap.Uint64("id", id),
	)

	query := psql.
		Select("id", "user_id", "dataset_id", "message", "is_read", "created_at").
		From("notifications").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find notification by ID query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
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
			r.logger.Warn("notification not found by ID",
				zap.Uint64("id", id),
			)
			return nil, repositories.ErrNotificationNotFound
		}
		r.logger.Error("failed to execute find notification by ID query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, repositories.ErrNotificationCreate
	}

	r.logger.Info("notification fetched successfully",
		zap.Uint64("id", n.ID),
		zap.Uint64("user_id", n.UserID),
	)
	return n, nil
}

func (r *NotificationRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Notification, error) {
	r.logger.Debug("FindByUserID Notifications called",
		zap.Uint64("user_id", userID),
	)

	query := psql.
		Select("id", "user_id", "dataset_id", "message", "is_read", "created_at").
		From("notifications").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find notifications by user query",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		return nil, repositories.ErrNotificationQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute find notifications by user query",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
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
			r.logger.Error("failed to scan notification row",
				zap.Error(err),
			)
			return nil, repositories.ErrNotificationScanRow
		}
		list = append(list, n)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating over notification rows",
			zap.Error(err),
		)
		return nil, repositories.ErrNotificationIterateRows
	}

	r.logger.Info("notifications fetched by user successfully",
		zap.Uint64("user_id", userID),
		zap.Int("count", len(list)),
	)
	return list, nil
}

func (r *NotificationRepo) Update(ctx context.Context, n *entities.Notification) error {
	r.logger.Debug("Update Notification called",
		zap.Uint64("id", n.ID),
		zap.String("message", n.Message),
		zap.Bool("is_read", n.IsRead),
	)

	query := psql.
		Update("notifications").
		Set("message", n.Message).
		Set("is_read", n.IsRead).
		Where(sq.Eq{"id": n.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build update notification query",
			zap.Error(err),
			zap.Uint64("id", n.ID),
		)
		return repositories.ErrNotificationQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute update notification query",
			zap.Error(err),
			zap.Uint64("id", n.ID),
		)
		return repositories.ErrNotificationUpdateFail
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no notification found to update",
			zap.Uint64("id", n.ID),
		)
		return repositories.ErrNotificationNotFound
	}

	r.logger.Info("notification updated successfully",
		zap.Uint64("id", n.ID),
	)
	return nil
}

func (r *NotificationRepo) Delete(ctx context.Context, id uint64) error {
	r.logger.Debug("Delete Notification called",
		zap.Uint64("id", id),
	)

	query := psql.Delete("notifications").Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build delete notification query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return repositories.ErrNotificationQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute delete notification query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return repositories.ErrNotificationDeleteFail
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no notification found to delete",
			zap.Uint64("id", id),
		)
		return repositories.ErrNotificationNotFound
	}

	r.logger.Info("notification deleted successfully",
		zap.Uint64("id", id),
	)
	return nil
}
