package postqbuild

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/repositories"

	"ppo/internal/entities"
)

type MetadataRepo struct {
	db *pgxpool.Pool
}

func NewMetadataRepo(pool *pgxpool.Pool) *MetadataRepo {
	return &MetadataRepo{db: pool}
}

func (r *MetadataRepo) Create(ctx context.Context, m *entities.Metadata) error {
	const sql = `
	INSERT INTO metadata
	  (format, size, tags, dataset_version_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`
	err := r.db.QueryRow(ctx, sql,
		m.Format,
		m.Size,
		m.Tags,
		m.DatasetVersionID,
	).Scan(&m.ID)
	if err != nil {
		return fmt.Errorf("create metadata: %w", err)
	}
	return nil
}

func (r *MetadataRepo) Update(ctx context.Context, m *entities.Metadata) error {
	const sql = `
	UPDATE metadata
	SET format = $1,
	    size = $2,
	    tags = $3
	WHERE id = $4
	`
	cmd, err := r.db.Exec(ctx, sql,
		m.Format,
		m.Size,
		m.Tags,
		m.ID,
	)
	if err != nil {
		return fmt.Errorf("update metadata: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrMetadataNotFound
	}
	return nil
}

func (r *MetadataRepo) Delete(ctx context.Context, id uint64) error {
	const sql = `DELETE FROM metadata WHERE id = $1`
	exec, err := r.db.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("delete metadata: %w", err)
	}
	if exec.RowsAffected() == 0 {
		return repositories.ErrMetadataNotFound
	}
	return nil
}

func (r *MetadataRepo) FindByID(ctx context.Context, id uint64) (*entities.Metadata, error) {
	const sql = `
	SELECT id, format, size, tags, dataset_version_id
	FROM metadata
	WHERE id = $1
	`
	m := &entities.Metadata{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&m.ID,
		&m.Format,
		&m.Size,
		&m.Tags,
		&m.DatasetVersionID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrMetadataNotFound
		}
		return nil, fmt.Errorf("find metadata by id: %w", err)
	}
	return m, nil
}

func (r *MetadataRepo) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Metadata, error) {
	const sql = `
	SELECT id, format, size, tags, dataset_version_id
	FROM metadata
	WHERE dataset_version_id = $1
	ORDER BY id DESC
	`
	rows, err := r.db.Query(ctx, sql, datasetID)
	if err != nil {
		return nil, fmt.Errorf("query metadata by dataset version id: %w", err)
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
			return nil, fmt.Errorf("scan metadata row: %w", err)
		}
		list = append(list, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate metadata rows: %w", err)
	}
	return list, nil
}
