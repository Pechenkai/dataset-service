package postqbuild

import (
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/entities"
	"ppo/internal/repositories"
	"time"
)

type VersionRepo struct {
	db *pgxpool.Pool
}

func NewVersionRepo(pool *pgxpool.Pool) *VersionRepo {
	return &VersionRepo{db: pool}
}

func (r *VersionRepo) Create(ctx context.Context, v *entities.DatasetVersion) error {
	if v.UploadDate.IsZero() {
		v.UploadDate = time.Now().UTC()
	}

	query := psql.
		Insert("dataset_versions").
		Columns("number", "upload_date", "filepath", "dataset_id", "change_log").
		Values(v.Number, v.UploadDate, v.Filepath, v.DatasetID, v.ChangeLog).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return repositories.ErrVersionQueryBuild
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&v.ID)
	if err != nil {
		return repositories.ErrVersionCreate
	}
	return nil
}

func (r *VersionRepo) Update(ctx context.Context, v *entities.DatasetVersion) error {
	if v.UploadDate.IsZero() {
		v.UploadDate = time.Now().UTC()
	}

	query := psql.
		Update("dataset_versions").
		Set("number", v.Number).
		Set("upload_date", v.UploadDate).
		Set("filepath", v.Filepath).
		Set("change_log", v.ChangeLog).
		Where(sq.Eq{"id": v.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return repositories.ErrVersionQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return repositories.ErrVersionUpdate
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrVersionNotFound
	}
	return nil
}

func (r *VersionRepo) Delete(ctx context.Context, id uint64) error {
	query := psql.
		Delete("dataset_versions").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return repositories.ErrVersionQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return repositories.ErrVersionDelete
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrVersionNotFound
	}
	return nil
}

func (r *VersionRepo) FindByID(ctx context.Context, id uint64) (*entities.DatasetVersion, error) {
	query := psql.
		Select("id", "number", "upload_date", "filepath", "dataset_id", "change_log").
		From("dataset_versions").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
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
			return nil, repositories.ErrVersionNotFound
		}
		return nil, repositories.ErrVersionScan
	}
	return v, nil
}

func (r *VersionRepo) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.DatasetVersion, error) {
	query := psql.
		Select("id", "number", "upload_date", "filepath", "dataset_id", "change_log").
		From("dataset_versions").
		Where(sq.Eq{"dataset_id": datasetID}).
		OrderBy("upload_date DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, repositories.ErrVersionQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
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
			return nil, repositories.ErrVersionScan
		}
		list = append(list, v)
	}
	if err := rows.Err(); err != nil {
		return nil, repositories.ErrVersionList
	}
	return list, nil
}
