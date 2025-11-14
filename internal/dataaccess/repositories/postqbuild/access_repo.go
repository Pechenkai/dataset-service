package postqbuild

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"

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

type accessRequestRow struct {
	ID        uint64
	DatasetID uint64
	UserID    uint64
	Status    string
	CreatedAt time.Time
}

func accessRequestRowFromEntity(ar *entities.AccessRequest) accessRequestRow {
	if ar == nil {
		return accessRequestRow{}
	}
	return accessRequestRow{
		ID:        ar.ID,
		DatasetID: ar.DatasetID,
		UserID:    ar.UserID,
		Status:    string(ar.Status),
		CreatedAt: ar.CreatedAt,
	}
}

func (row accessRequestRow) toEntity() *entities.AccessRequest {
	status := entities.AccessStatus(row.Status)
	if _, ok := entities.ValidAccessStatuses[status]; !ok {
		status = entities.AccessStatusPending
	}
	return &entities.AccessRequest{
		ID:        row.ID,
		DatasetID: row.DatasetID,
		UserID:    row.UserID,
		Status:    status,
		CreatedAt: row.CreatedAt,
	}
}

func NewAccessRequestRepo(pool *pgxpool.Pool, logger *zap.Logger) *AccessRequestRepo {
	logger.Debug("NewAccessRequestRepo initialized")
	return &AccessRequestRepo{db: pool, logger: logger}
}

func (r *AccessRequestRepo) Create(ctx context.Context, ar *entities.AccessRequest) error {
	row := accessRequestRowFromEntity(ar)
	r.logger.Debug("Create AccessRequest called",
		zap.Uint64("dataset_id", row.DatasetID),
		zap.Uint64("user_id", row.UserID),
		zap.String("status", row.Status),
	)

	query := psql.
		Insert("access_requests").
		Columns("dataset_id", "user_id", "status", "created_at").
		Values(row.DatasetID, row.UserID, row.Status, row.CreatedAt).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build Create query", zap.Error(err))
		return repositories.ErrRequestQueryBuild
	}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&row.ID)
	if err != nil {
		r.logger.Error("failed to execute Create query", zap.Error(err))
		return fmt.Errorf("create access_request: %w", err)
	}
	ar.ID = row.ID
	r.logger.Info("AccessRequest created", zap.Uint64("id", row.ID))
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

	var row accessRequestRow
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&row.ID, &row.DatasetID, &row.UserID, &row.Status, &row.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("AccessRequest not found", zap.Uint64("dataset_id", datasetID), zap.Uint64("user_id", userID))
			return nil, nil
		}
		r.logger.Error("failed to execute Find query", zap.Error(err))
		return nil, fmt.Errorf("find access_request: %w", err)
	}

	entity := row.toEntity()
	r.logger.Info("AccessRequest fetched", zap.Uint64("id", entity.ID))
	return entity, nil
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
		row := accessRequestRow{}
		if err := rows.Scan(&row.ID, &row.DatasetID, &row.UserID, &row.Status, &row.CreatedAt); err != nil {
			r.logger.Error("failed to scan row", zap.Error(err))
			return nil, repositories.ErrRequestScan
		}
		list = append(list, row.toEntity())
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

	row := &accessRequestRow{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&row.ID, &row.DatasetID, &row.UserID, &row.Status, &row.CreatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "P0002" {
			r.logger.Info("access request not found by ID", zap.Uint64("request_id", requestID))
			return nil, repositories.ErrRequestNotFound
		}
		r.logger.Error("failed to execute FindByRequestID query", zap.Error(err))
		return nil, fmt.Errorf("find access_request by id: %w", err)
	}

	entity := row.toEntity()
	r.logger.Info("AccessRequest fetched by ID", zap.Uint64("request_id", entity.ID))
	return entity, nil
}
