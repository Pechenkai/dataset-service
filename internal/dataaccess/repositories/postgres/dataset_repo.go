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

type DatasetRepo struct {
	db *pgxpool.Pool
}

func NewDatasetRepo(pool *pgxpool.Pool) *DatasetRepo {
	return &DatasetRepo{db: pool}
}

func (r *DatasetRepo) Create(ctx context.Context, d *entities.Dataset) error {
	const sql = `
	INSERT INTO datasets
	  (name, description, owner_id, category_id, is_public, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id
	`
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}
	err := r.db.QueryRow(ctx, sql,
		d.Name,
		d.Description,
		d.OwnerID,
		d.CategoryID,
		d.IsPublic,
		d.CreatedAt,
	).Scan(&d.ID)
	if err != nil {
		return fmt.Errorf("create dataset: %w", err)
	}
	return nil
}

func (r *DatasetRepo) Update(ctx context.Context, d *entities.Dataset) error {
	const sql = `
	UPDATE datasets
	SET name = $1,
	    description = $2,
	    category_id = $3,
	    is_public = $4
	WHERE id = $5
	`
	cmd, err := r.db.Exec(ctx, sql,
		d.Name,
		d.Description,
		d.CategoryID,
		d.IsPublic,
		d.ID,
	)
	if err != nil {
		return fmt.Errorf("update dataset: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrDatasetNotFound
	}
	return nil
}

func (r *DatasetRepo) Delete(ctx context.Context, id uint64) error {
	const sql = `DELETE FROM datasets WHERE id = $1`
	exec, err := r.db.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("delete dataset: %w", err)
	}
	if exec.RowsAffected() == 0 {
		return ErrDatasetNotFound
	}
	return nil
}

// FindByID возвращает один датасет или ErrDatasetNotFound, если не найден.
func (r *DatasetRepo) FindByID(ctx context.Context, id uint64) (*entities.Dataset, error) {
	const sql = `
	SELECT id, name, description, owner_id, category_id, is_public, created_at
	FROM datasets
	WHERE id = $1
	`
	d := &entities.Dataset{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&d.ID,
		&d.Name,
		&d.Description,
		&d.OwnerID,
		&d.CategoryID,
		&d.IsPublic,
		&d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDatasetNotFound
		}
		return nil, fmt.Errorf("find dataset by id: %w", err)
	}
	return d, nil
}

func (r *DatasetRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Dataset, error) {
	const sql = `
	SELECT id, name, description, owner_id, category_id, is_public, created_at
	FROM datasets
	WHERE owner_id = $1
	ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("query datasets by user id: %w", err)
	}
	defer rows.Close()

	var list []*entities.Dataset
	for rows.Next() {
		d := &entities.Dataset{}
		if err := rows.Scan(
			&d.ID,
			&d.Name,
			&d.Description,
			&d.OwnerID,
			&d.CategoryID,
			&d.IsPublic,
			&d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dataset row: %w", err)
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dataset rows: %w", err)
	}
	return list, nil
}

func (r *DatasetRepo) FindAll(ctx context.Context) ([]*entities.Dataset, error) {
	const sql = `
	SELECT id, name, description, owner_id, category_id, is_public, created_at
	FROM datasets
	ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query all datasets: %w", err)
	}
	defer rows.Close()

	var list []*entities.Dataset
	for rows.Next() {
		d := &entities.Dataset{}
		if err := rows.Scan(
			&d.ID,
			&d.Name,
			&d.Description,
			&d.OwnerID,
			&d.CategoryID,
			&d.IsPublic,
			&d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dataset row: %w", err)
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dataset rows: %w", err)
	}
	return list, nil
}
