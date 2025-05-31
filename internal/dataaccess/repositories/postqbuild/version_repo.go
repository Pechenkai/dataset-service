package postqbuild

import (
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"ppo/internal/entities"
	"ppo/internal/repositories"
	"time"
)

type VersionRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewVersionRepo(pool *pgxpool.Pool, logger *zap.Logger) *VersionRepo {
	logger.Debug("NewVersionRepo initialized")
	return &VersionRepo{db: pool, logger: logger}
}

func (r *VersionRepo) Create(ctx context.Context, v *entities.DatasetVersion) error {
	if v.UploadDate.IsZero() {
		v.UploadDate = time.Now().UTC()
	}
	r.logger.Debug("Create Version called",
		zap.String("number", v.Number),
		zap.Uint64("dataset_id", v.DatasetID),
		zap.String("filepath", v.Filepath),
		zap.String("change_log", v.ChangeLog),
	)

	query := psql.
		Insert("dataset_versions").
		Columns("number", "upload_date", "filepath", "dataset_id", "change_log").
		Values(v.Number, v.UploadDate, v.Filepath, v.DatasetID, v.ChangeLog).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build create version SQL",
			zap.Error(err),
		)
		return repositories.ErrVersionQueryBuild
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&v.ID)
	if err != nil {
		r.logger.Error("failed to execute create version query",
			zap.Error(err),
		)
		return repositories.ErrVersionCreate
	}

	r.logger.Info("version created successfully",
		zap.Uint64("id", v.ID),
		zap.String("number", v.Number),
		zap.Uint64("dataset_id", v.DatasetID),
	)
	return nil
}

func (r *VersionRepo) Update(ctx context.Context, v *entities.DatasetVersion) error {
	if v.UploadDate.IsZero() {
		v.UploadDate = time.Now().UTC()
	}
	r.logger.Debug("Update Version called",
		zap.Uint64("id", v.ID),
		zap.String("number", v.Number),
		zap.Uint64("dataset_id", v.DatasetID),
		zap.String("filepath", v.Filepath),
		zap.String("change_log", v.ChangeLog),
	)

	query := psql.
		Update("dataset_versions").
		Set("number", v.Number).
		Set("upload_date", v.UploadDate).
		Set("filepath", v.Filepath).
		Set("change_log", v.ChangeLog).
		Where(sq.Eq{"id": v.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build update version SQL",
			zap.Error(err),
			zap.Uint64("id", v.ID),
		)
		return repositories.ErrVersionQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute update version query",
			zap.Error(err),
			zap.Uint64("id", v.ID),
		)
		return repositories.ErrVersionUpdate
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no version found to update",
			zap.Uint64("id", v.ID),
		)
		return repositories.ErrVersionNotFound
	}

	r.logger.Info("version updated successfully",
		zap.Uint64("id", v.ID),
		zap.String("number", v.Number),
	)
	return nil
}

func (r *VersionRepo) Delete(ctx context.Context, id uint64) error {
	r.logger.Debug("Delete Version called",
		zap.Uint64("id", id),
	)

	query := psql.
		Delete("dataset_versions").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build delete version SQL",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return repositories.ErrVersionQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute delete version query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return repositories.ErrVersionDelete
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no version found to delete",
			zap.Uint64("id", id),
		)
		return repositories.ErrVersionNotFound
	}

	r.logger.Info("version deleted successfully",
		zap.Uint64("id", id),
	)
	return nil
}

func (r *VersionRepo) FindByID(ctx context.Context, id uint64) (*entities.DatasetVersion, error) {
	r.logger.Debug("FindByID Version called",
		zap.Uint64("id", id),
	)

	query := psql.
		Select("id", "number", "upload_date", "filepath", "dataset_id", "change_log").
		From("dataset_versions").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find version by ID SQL",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, repositories.ErrVersionQueryBuild
	}

	v := &entities.DatasetVersion{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&v.ID,
		&v.Number,
		&v.UploadDate,
		&v.Filepath,
		&v.DatasetID,
		&v.ChangeLog,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("version not found by ID",
				zap.Uint64("id", id),
			)
			return nil, repositories.ErrVersionNotFound
		}
		r.logger.Error("failed to execute find version by ID query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, repositories.ErrVersionScan
	}

	r.logger.Info("version fetched successfully",
		zap.Uint64("id", v.ID),
		zap.String("number", v.Number),
	)
	return v, nil
}

func (r *VersionRepo) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.DatasetVersion, error) {
	r.logger.Debug("FindByDatasetID Versions called",
		zap.Uint64("dataset_id", datasetID),
	)

	query := psql.
		Select("id", "number", "upload_date", "filepath", "dataset_id", "change_log").
		From("dataset_versions").
		Where(sq.Eq{"dataset_id": datasetID}).
		OrderBy("upload_date DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find versions by dataset ID SQL",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, repositories.ErrVersionQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute find versions by dataset ID query",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, repositories.ErrVersionList
	}
	defer rows.Close()

	var list []*entities.DatasetVersion
	for rows.Next() {
		v := &entities.DatasetVersion{}
		if err := rows.Scan(
			&v.ID,
			&v.Number,
			&v.UploadDate,
			&v.Filepath,
			&v.DatasetID,
			&v.ChangeLog,
		); err != nil {
			r.logger.Error("failed to scan version row",
				zap.Error(err),
			)
			return nil, repositories.ErrVersionScan
		}
		list = append(list, v)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating over version rows",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, repositories.ErrVersionList
	}

	r.logger.Info("versions fetched by dataset successfully",
		zap.Uint64("dataset_id", datasetID),
		zap.Int("count", len(list)),
	)
	return list, nil
}
