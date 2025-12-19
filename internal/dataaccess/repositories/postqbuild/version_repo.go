package postqbuild

import (
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"
	pgx "github.com/jackc/pgx/v5"
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

type versionRow struct {
	ID         uint64
	Number     string
	UploadDate time.Time
	Filepath   string
	DatasetID  uint64
	ChangeLog  string
}

func versionRowFromEntity(v *entities.DatasetVersion) versionRow {
	if v == nil {
		return versionRow{}
	}
	return versionRow{
		ID:         v.ID,
		Number:     v.Number,
		UploadDate: v.UploadDate,
		Filepath:   v.Filepath,
		DatasetID:  v.DatasetID,
		ChangeLog:  v.ChangeLog,
	}
}

func (row versionRow) toEntity() *entities.DatasetVersion {
	return &entities.DatasetVersion{
		ID:         row.ID,
		Number:     row.Number,
		UploadDate: row.UploadDate,
		Filepath:   row.Filepath,
		DatasetID:  row.DatasetID,
		ChangeLog:  row.ChangeLog,
	}
}

func NewVersionRepo(pool *pgxpool.Pool, logger *zap.Logger) *VersionRepo {
	logger.Debug("NewVersionRepo initialized")
	return &VersionRepo{db: pool, logger: logger}
}

func (r *VersionRepo) Create(ctx context.Context, v *entities.DatasetVersion) error {
	row := versionRowFromEntity(v)
	if row.UploadDate.IsZero() {
		row.UploadDate = time.Now().UTC()
		v.UploadDate = row.UploadDate
	}
	r.logger.Debug("Create Version called",
		zap.String("number", row.Number),
		zap.Uint64("dataset_id", row.DatasetID),
		zap.String("filepath", row.Filepath),
		zap.String("change_log", row.ChangeLog),
	)

	query := psql.
		Insert("dataset_versions").
		Columns("number", "upload_date", "filepath", "dataset_id", "change_log").
		Values(row.Number, row.UploadDate, row.Filepath, row.DatasetID, row.ChangeLog).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build create version SQL",
			zap.Error(err),
		)
		return repositories.ErrVersionQueryBuild
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&row.ID)
	if err != nil {
		r.logger.Error("failed to execute create version query",
			zap.Error(err),
		)
		return repositories.ErrVersionCreate
	}

	v.ID = row.ID
	r.logger.Info("version created successfully",
		zap.Uint64("id", row.ID),
		zap.String("number", row.Number),
		zap.Uint64("dataset_id", row.DatasetID),
	)
	return nil
}

func (r *VersionRepo) Update(ctx context.Context, v *entities.DatasetVersion) error {
	row := versionRowFromEntity(v)
	if row.UploadDate.IsZero() {
		row.UploadDate = time.Now().UTC()
		v.UploadDate = row.UploadDate
	}
	r.logger.Debug("Update Version called",
		zap.Uint64("id", row.ID),
		zap.String("number", row.Number),
		zap.Uint64("dataset_id", row.DatasetID),
		zap.String("filepath", row.Filepath),
		zap.String("change_log", row.ChangeLog),
	)

	query := psql.
		Update("dataset_versions").
		Set("number", row.Number).
		Set("upload_date", row.UploadDate).
		Set("filepath", row.Filepath).
		Set("change_log", row.ChangeLog).
		Where(sq.Eq{"id": row.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build update version SQL",
			zap.Error(err),
			zap.Uint64("id", row.ID),
		)
		return repositories.ErrVersionQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute update version query",
			zap.Error(err),
			zap.Uint64("id", row.ID),
		)
		return repositories.ErrVersionUpdate
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no version found to update",
			zap.Uint64("id", row.ID),
		)
		return repositories.ErrVersionNotFound
	}

	r.logger.Info("version updated successfully",
		zap.Uint64("id", row.ID),
		zap.String("number", row.Number),
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

	row := &versionRow{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&row.ID,
		&row.Number,
		&row.UploadDate,
		&row.Filepath,
		&row.DatasetID,
		&row.ChangeLog,
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

	entity := row.toEntity()
	r.logger.Info("version fetched successfully",
		zap.Uint64("id", entity.ID),
		zap.String("number", entity.Number),
	)
	return entity, nil
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
		row := versionRow{}
		if err := rows.Scan(
			&row.ID,
			&row.Number,
			&row.UploadDate,
			&row.Filepath,
			&row.DatasetID,
			&row.ChangeLog,
		); err != nil {
			r.logger.Error("failed to scan version row",
				zap.Error(err),
			)
			return nil, repositories.ErrVersionScan
		}
		list = append(list, row.toEntity())
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
