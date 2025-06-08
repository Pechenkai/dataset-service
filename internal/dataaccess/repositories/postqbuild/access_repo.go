package postqbuild

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type AccessRequestRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewAccessRequestRepo(pool *pgxpool.Pool, logger *zap.Logger) *AccessRequestRepo {
	logger.Debug("NewAccessRequestRepo initialized")
	return &AccessRequestRepo{db: pool, logger: logger}
}

func (r *AccessRequestRepo) Create(ctx context.Context, ar *entities.AccessRequest) error {
	r.logger.Debug("Create AccessRequest called",
		zap.Uint64("dataset_id", ar.DatasetID),
		zap.Uint64("user_id", ar.UserID),
		zap.String("status", string(ar.Status)),
	)

	query := psql.
		Insert("access_requests").
		Columns("dataset_id", "user_id", "status", "created_at").
		Values(ar.DatasetID, ar.UserID, string(ar.Status), ar.CreatedAt).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build Create query", zap.Error(err))
		return repositories.ErrRequestQueryBuild
	}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&ar.ID)
	if err != nil {
		r.logger.Error("failed to execute Create query", zap.Error(err))
		return fmt.Errorf("create access_request: %w", err)
	}
	r.logger.Info("AccessRequest created", zap.Uint64("id", ar.ID))
	return nil
}

func (r *AccessRequestRepo) Find(ctx context.Context, datasetID, userID uint64) (*entities.AccessRequest, error) {
	r.logger.Debug("Find AccessRequest called",
		zap.Uint64("dataset_id", datasetID),
		zap.Uint64("user_id", userID),
	)

	query := psql.
		Select("id", "dataset_id", "user_id", "status", "created_at").
		From("access_requests").
		Where(sq.Eq{"dataset_id": datasetID, "user_id": userID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build Find query", zap.Error(err))
		return nil, repositories.ErrRequestQueryBuild
	}

	var ar entities.AccessRequest
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&ar.ID, &ar.DatasetID, &ar.UserID, &ar.Status, &ar.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("AccessRequest not found", zap.Uint64("dataset_id", datasetID), zap.Uint64("user_id", userID))
			return nil, nil
		}
		r.logger.Error("failed to execute Find query", zap.Error(err))
		return nil, fmt.Errorf("find access_request: %w", err)
	}

	r.logger.Info("AccessRequest fetched", zap.Uint64("id", ar.ID))
	return &ar, nil
}

func (r *AccessRequestRepo) ListPendingByOwner(ctx context.Context, ownerID uint64) ([]*entities.AccessRequest, error) {
	r.logger.Debug("ListPendingByOwner called", zap.Uint64("owner_id", ownerID))

	query := psql.
		Select("ar.id", "ar.dataset_id", "ar.user_id", "ar.status", "ar.created_at").
		From("access_requests ar").
		Join("datasets d ON ar.dataset_id = d.id").
		Where(sq.And{
			sq.Eq{"d.owner_id": ownerID},
			sq.Eq{"ar.status": string(entities.AccessStatusPending)},
		})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build ListPendingByOwner query", zap.Error(err))
		return nil, repositories.ErrRequestQueryBuild
	}
	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute ListPendingByOwner query", zap.Error(err))
		return nil, fmt.Errorf("list pending: %w", err)
	}
	defer rows.Close()

	var list []*entities.AccessRequest
	for rows.Next() {
		ar := &entities.AccessRequest{}
		if err := rows.Scan(&ar.ID, &ar.DatasetID, &ar.UserID, &ar.Status, &ar.CreatedAt); err != nil {
			r.logger.Error("failed to scan row", zap.Error(err))
			return nil, repositories.ErrRequestScan
		}
		list = append(list, ar)
	}
	if rows.Err() != nil {
		r.logger.Error("error iterating rows", zap.Error(rows.Err()))
		return nil, repositories.ErrRequestScan
	}
	r.logger.Info("ListPendingByOwner completed", zap.Uint64("owner_id", ownerID), zap.Int("count", len(list)))
	return list, nil
}

func (r *AccessRequestRepo) UpdateStatus(ctx context.Context, id uint64, status string) error {
	r.logger.Debug("UpdateStatus called", zap.Uint64("id", id), zap.String("status", string(status)))
	query := psql.
		Update("access_requests").
		Set("status", string(status)).
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build UpdateStatus query", zap.Error(err))
		return repositories.ErrRequestQueryBuild
	}
	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute UpdateStatus query", zap.Error(err))
		return fmt.Errorf("update status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no request found to update", zap.Uint64("id", id))
		return repositories.ErrRequestNotFound
	}
	r.logger.Info("UpdateStatus completed", zap.Uint64("id", id), zap.String("status", string(status)))
	return nil
}

func (r *AccessRequestRepo) FindByRequestID(ctx context.Context, requestID uint64) (*entities.AccessRequest, error) {
	r.logger.Debug("FindByRequestID called", zap.Uint64("request_id", requestID))

	query := psql.
		Select("id", "dataset_id", "user_id", "status", "created_at").
		From("access_requests").
		Where(sq.Eq{"id": requestID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build FindByRequestID query", zap.Error(err))
		return nil, repositories.ErrRequestQueryBuild
	}

	ar := &entities.AccessRequest{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&ar.ID, &ar.DatasetID, &ar.UserID, &ar.Status, &ar.CreatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "P0002" {
			r.logger.Info("access request not found by ID", zap.Uint64("request_id", requestID))
			return nil, repositories.ErrRequestNotFound
		}
		r.logger.Error("failed to execute FindByRequestID query", zap.Error(err))
		return nil, fmt.Errorf("find access_request by id: %w", err)
	}

	r.logger.Info("AccessRequest fetched by ID", zap.Uint64("request_id", ar.ID))
	return ar, nil
}
