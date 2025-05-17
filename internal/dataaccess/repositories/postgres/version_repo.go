package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/entities"
	"time"
)

type VersionRepo struct {
	db *pgxpool.Pool
}

func NewVersionRepo(pool *pgxpool.Pool) *VersionRepo {
	return &VersionRepo{db: pool}
}

func (r *VersionRepo) Create(ctx context.Context, v *entities.DatasetVersion) error {
	const sql = `
	INSERT INTO dataset_versions
	  (number, upload_date, filepath, dataset_id, change_log)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
	`
	if v.UploadDate.IsZero() {
		v.UploadDate = time.Now().UTC()
	}
	err := r.db.QueryRow(ctx, sql,
		v.Number,
		v.UploadDate,
		v.Filepath,
		v.DatasetID,
		v.ChangeLog,
	).Scan(&v.ID)
	if err != nil {
		return fmt.Errorf("create version: %w", err)
	}
	return nil
}

func (r *VersionRepo) Update(ctx context.Context, v *entities.DatasetVersion) error {
	const sql = `
	UPDATE dataset_versions
	SET number = $1,
	    upload_date = $2,
	    filepath = $3,
	    change_log = $4
	WHERE id = $5
	`
	cmd, err := r.db.Exec(ctx, sql,
		v.Number,
		v.UploadDate,
		v.Filepath,
		v.ChangeLog,
		v.ID,
	)
	if err != nil {
		return fmt.Errorf("update version: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrVersionNotFound
	}
	return nil
}

func (r *VersionRepo) Delete(ctx context.Context, id uint64) error {
	const sql = `DELETE FROM dataset_versions WHERE id = $1`
	exec, err := r.db.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("delete version: %w", err)
	}
	if exec.RowsAffected() == 0 {
		return ErrVersionNotFound
	}
	return nil
}

func (r *VersionRepo) FindByID(ctx context.Context, id uint64) (*entities.DatasetVersion, error) {
	const sql = `
	SELECT id, number, upload_date, filepath, dataset_id, change_log
	FROM dataset_versions
	WHERE id = $1
	`
	v := &entities.DatasetVersion{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&v.ID,
		&v.Number,
		&v.UploadDate,
		&v.Filepath,
		&v.DatasetID,
		&v.ChangeLog,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVersionNotFound
		}
		return nil, fmt.Errorf("find version by id: %w", err)
	}
	return v, nil
}

func (r *VersionRepo) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.DatasetVersion, error) {
	const sql = `
	SELECT id, number, upload_date, filepath, dataset_id, change_log
	FROM dataset_versions
	WHERE dataset_id = $1
	ORDER BY upload_date DESC
	`
	rows, err := r.db.Query(ctx, sql, datasetID)
	if err != nil {
		return nil, fmt.Errorf("query versions by dataset id: %w", err)
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
			return nil, fmt.Errorf("scan version row: %w", err)
		}
		list = append(list, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate version rows: %w", err)
	}
	return list, nil
}
