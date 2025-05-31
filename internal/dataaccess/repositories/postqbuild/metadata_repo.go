package postqbuild

import (
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type MetadataRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewMetadataRepo(pool *pgxpool.Pool, logger *zap.Logger) *MetadataRepo {
	logger.Debug("NewMetadataRepo initialized")
	return &MetadataRepo{db: pool, logger: logger}
}

func (r *MetadataRepo) Create(ctx context.Context, m *entities.Metadata) error {
	r.logger.Debug("Create Metadata called",
		zap.String("format", m.Format),
		zap.Uint64("size", m.Size),
		zap.String("tags", m.Tags),
		zap.Uint64("dataset_version_id", m.DatasetVersionID),
	)

	query := psql.
		Insert("metadata").
		Columns("format", "size", "tags", "dataset_version_id").
		Values(m.Format, m.Size, m.Tags, m.DatasetVersionID).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build create metadata query",
			zap.Error(err),
		)
		return repositories.ErrMetadataQueryBuild
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&m.ID)
	if err != nil {
		r.logger.Error("failed to execute create metadata query",
			zap.Error(err),
		)
		return repositories.ErrMetadataCreate
	}

	r.logger.Info("metadata created successfully",
		zap.Uint64("id", m.ID),
		zap.Uint64("dataset_version_id", m.DatasetVersionID),
	)
	return nil
}

func (r *MetadataRepo) Update(ctx context.Context, m *entities.Metadata) error {
	r.logger.Debug("Update Metadata called",
		zap.Uint64("id", m.ID),
		zap.String("format", m.Format),
		zap.Uint64("size", m.Size),
		zap.String("tags", m.Tags),
	)

	query := psql.
		Update("metadata").
		Set("format", m.Format).
		Set("size", m.Size).
		Set("tags", m.Tags).
		Where(sq.Eq{"id": m.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build update metadata query",
			zap.Error(err),
			zap.Uint64("id", m.ID),
		)
		return repositories.ErrMetadataQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute update metadata query",
			zap.Error(err),
			zap.Uint64("id", m.ID),
		)
		return repositories.ErrMetadataUpdate
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no metadata found to update",
			zap.Uint64("id", m.ID),
		)
		return repositories.ErrMetadataNotFound
	}

	r.logger.Info("metadata updated successfully",
		zap.Uint64("id", m.ID),
	)
	return nil
}

func (r *MetadataRepo) Delete(ctx context.Context, id uint64) error {
	r.logger.Debug("Delete Metadata called",
		zap.Uint64("id", id),
	)

	query := psql.Delete("metadata").Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build delete metadata query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return repositories.ErrMetadataQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute delete metadata query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return repositories.ErrMetadataDelete
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no metadata found to delete",
			zap.Uint64("id", id),
		)
		return repositories.ErrMetadataNotFound
	}

	r.logger.Info("metadata deleted successfully",
		zap.Uint64("id", id),
	)
	return nil
}

func (r *MetadataRepo) FindByID(ctx context.Context, id uint64) (*entities.Metadata, error) {
	r.logger.Debug("FindByID Metadata called",
		zap.Uint64("id", id),
	)

	query := psql.
		Select("id", "format", "size", "tags", "dataset_version_id").
		From("metadata").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find metadata by ID query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, repositories.ErrMetadataQueryBuild
	}

	m := &entities.Metadata{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&m.ID,
		&m.Format,
		&m.Size,
		&m.Tags,
		&m.DatasetVersionID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("metadata not found by ID",
				zap.Uint64("id", id),
			)
			return nil, repositories.ErrMetadataNotFound
		}
		r.logger.Error("failed to execute find metadata by ID query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, repositories.ErrMetadataScan
	}

	r.logger.Info("metadata fetched successfully",
		zap.Uint64("id", m.ID),
		zap.Uint64("dataset_version_id", m.DatasetVersionID),
	)
	return m, nil
}

func (r *MetadataRepo) FindByDatasetID(ctx context.Context, datasetVersionID uint64) ([]*entities.Metadata, error) {
	r.logger.Debug("FindByDatasetID Metadata called",
		zap.Uint64("dataset_version_id", datasetVersionID),
	)

	query := psql.
		Select("id", "format", "size", "tags", "dataset_version_id").
		From("metadata").
		Where(sq.Eq{"dataset_version_id": datasetVersionID}).
		OrderBy("id DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find metadata by dataset version query",
			zap.Error(err),
			zap.Uint64("dataset_version_id", datasetVersionID),
		)
		return nil, repositories.ErrMetadataQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute find metadata by dataset version query",
			zap.Error(err),
			zap.Uint64("dataset_version_id", datasetVersionID),
		)
		return nil, repositories.ErrMetadataList
	}
	defer rows.Close()

	var list []*entities.Metadata
	for rows.Next() {
		m := &entities.Metadata{}
		if err := rows.Scan(
			&m.ID,
			&m.Format,
			&m.Size,
			&m.Tags,
			&m.DatasetVersionID,
		); err != nil {
			r.logger.Error("failed to scan metadata row",
				zap.Error(err),
			)
			return nil, repositories.ErrMetadataScan
		}
		list = append(list, m)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating over metadata rows",
			zap.Error(err),
		)
		return nil, repositories.ErrMetadataList
	}

	r.logger.Info("metadata fetched by dataset version successfully",
		zap.Uint64("dataset_version_id", datasetVersionID),
		zap.Int("count", len(list)),
	)
	return list, nil
}
