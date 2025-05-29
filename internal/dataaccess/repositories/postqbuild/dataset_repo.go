package postqbuild

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type DatasetRepo struct {
	db *pgxpool.Pool
}

func NewDatasetRepo(pool *pgxpool.Pool) *DatasetRepo {
	return &DatasetRepo{db: pool}
}

func (r *DatasetRepo) Create(ctx context.Context, d *entities.Dataset) error {
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}
	query := psql.
		Insert("datasets").
		Columns("name", "description", "owner_id", "category_id", "is_public", "created_at").
		Values(d.Name, d.Description, d.OwnerID, d.CategoryID, d.IsPublic, d.CreatedAt).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build insert dataset sql: %w", err)
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&d.ID)
	if err != nil {
		return fmt.Errorf("create dataset: %w", err)
	}
	return nil
}

func (r *DatasetRepo) Update(ctx context.Context, d *entities.Dataset) error {
	query := psql.
		Update("datasets").
		Set("name", d.Name).
		Set("description", d.Description).
		Set("category_id", d.CategoryID).
		Set("is_public", d.IsPublic).
		Where(sq.Eq{"id": d.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build update dataset sql: %w", err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("update dataset: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrDatasetNotFound
	}
	return nil
}

func (r *DatasetRepo) Delete(ctx context.Context, id uint64) error {
	query := psql.Delete("datasets").Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build delete dataset sql: %w", err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("delete dataset: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrDatasetNotFound
	}
	return nil
}

func (r *DatasetRepo) FindByID(ctx context.Context, id uint64) (*entities.Dataset, error) {
	query := psql.
		Select("id", "name", "description", "owner_id", "category_id", "is_public", "created_at").
		From("datasets").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find dataset by id sql: %w", err)
	}

	d := &entities.Dataset{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&d.ID, &d.Name, &d.Description,
		&d.OwnerID, &d.CategoryID, &d.IsPublic,
		&d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrDatasetNotFound
		}
		return nil, fmt.Errorf("find dataset by id: %w", err)
	}
	return d, nil
}

func (r *DatasetRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Dataset, error) {
	query := psql.
		Select("id", "name", "description", "owner_id", "category_id", "is_public", "created_at").
		From("datasets").
		Where(sq.Eq{"owner_id": userID}).
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find by user sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query datasets by user id: %w", err)
	}
	defer rows.Close()

	var list []*entities.Dataset
	for rows.Next() {
		d := &entities.Dataset{}
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Description,
			&d.OwnerID, &d.CategoryID, &d.IsPublic,
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
	query := psql.
		Select("id", "name", "description", "owner_id", "category_id", "is_public", "created_at").
		From("datasets").
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find all sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query all datasets: %w", err)
	}
	defer rows.Close()

	var list []*entities.Dataset
	for rows.Next() {
		d := &entities.Dataset{}
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Description,
			&d.OwnerID, &d.CategoryID, &d.IsPublic,
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

func (r *DatasetRepo) FindPublic(ctx context.Context) ([]*entities.Dataset, error) {
	query := psql.
		Select("id", "name", "description", "owner_id", "category_id", "is_public", "created_at").
		From("datasets").
		Where(sq.Eq{"is_public": true})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find public sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query public datasets: %w", err)
	}
	defer rows.Close()

	var list []*entities.Dataset
	for rows.Next() {
		d := &entities.Dataset{}
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Description,
			&d.OwnerID, &d.CategoryID, &d.IsPublic,
			&d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan public dataset row: %w", err)
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate public dataset rows: %w", err)
	}
	return list, nil
}
