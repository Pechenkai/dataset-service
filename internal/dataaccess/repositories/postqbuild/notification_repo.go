package postqbuild

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type NotificationRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

type notificationRow struct {
	ID        uint64
	UserID    uint64
	DatasetID uint64
	Message   string
	IsRead    bool
	CreatedAt time.Time
}

func notificationRowFromEntity(n *entities.Notification) notificationRow {
	if n == nil {
		return notificationRow{}
	}
	return notificationRow{
		ID:        n.ID,
		UserID:    n.UserID,
		DatasetID: n.DatasetID,
		Message:   n.Message,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt,
	}
}

func (row notificationRow) toEntity() *entities.Notification {
	return &entities.Notification{
		ID:        row.ID,
		UserID:    row.UserID,
		DatasetID: row.DatasetID,
		Message:   row.Message,
		IsRead:    row.IsRead,
		CreatedAt: row.CreatedAt,
	}
}

func NewNotificationRepo(pool *pgxpool.Pool, logger *zap.Logger) *NotificationRepo {
	logger.Debug("NewNotificationRepo initialized")
	return &NotificationRepo{db: pool, logger: logger}
}

func (r *NotificationRepo) Create(ctx context.Context, n *entities.Notification) error {
	row := notificationRowFromEntity(n)
	if row.CreatedAt.IsZero() {
		row.CreatedAt = time.Now().UTC()
		n.CreatedAt = row.CreatedAt
	}
	r.logger.Debug("Create Notification called",
		zap.Uint64("user_id", row.UserID),
		zap.Uint64("dataset_id", row.DatasetID),
		zap.String("message", row.Message),
		zap.Bool("is_read", row.IsRead),
	)

	query := psql.
		Insert("notifications").
		Columns("user_id", "dataset_id", "message", "is_read", "created_at").
		Values(row.UserID, row.DatasetID, row.Message, row.IsRead, row.CreatedAt).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build create notification query",
			zap.Error(err),
		)
		return repositories.ErrNotificationQueryBuild
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&row.ID)
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

	n.ID = row.ID
	r.logger.Info("notification created successfully",
		zap.Uint64("id", row.ID),
		zap.Uint64("user_id", row.UserID),
		zap.Uint64("dataset_id", row.DatasetID),
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

	row := &notificationRow{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&row.ID,
		&row.UserID,
		&row.DatasetID,
		&row.Message,
		&row.IsRead,
		&row.CreatedAt,
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

	entity := row.toEntity()
	r.logger.Info("notification fetched successfully",
		zap.Uint64("id", entity.ID),
		zap.Uint64("user_id", entity.UserID),
	)
	return entity, nil
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
		row := notificationRow{}
		if err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.DatasetID,
			&row.Message,
			&row.IsRead,
			&row.CreatedAt,
		); err != nil {
			r.logger.Error("failed to scan notification row",
				zap.Error(err),
			)
			return nil, repositories.ErrNotificationScanRow
		}
		list = append(list, row.toEntity())
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
	row := notificationRowFromEntity(n)
	r.logger.Debug("Update Notification called",
		zap.Uint64("id", row.ID),
		zap.String("message", row.Message),
		zap.Bool("is_read", row.IsRead),
	)

	query := psql.
		Update("notifications").
		Set("message", row.Message).
		Set("is_read", row.IsRead).
		Where(sq.Eq{"id": row.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build update notification query",
			zap.Error(err),
			zap.Uint64("id", row.ID),
		)
		return repositories.ErrNotificationQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute update notification query",
			zap.Error(err),
			zap.Uint64("id", row.ID),
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
		zap.Uint64("id", row.ID),
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
