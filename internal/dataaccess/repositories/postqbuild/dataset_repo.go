package postqbuild

import (
	"context"
	"errors"
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
		return repositories.ErrDatasetQueryBuild
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&d.ID)
	if err != nil {
		return repositories.ErrDatasetCreate
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
		return repositories.ErrDatasetQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return repositories.ErrDatasetUpdate
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
		return repositories.ErrDatasetQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return repositories.ErrDatasetDelete
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
		return nil, repositories.ErrDatasetQueryBuild
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
		return nil, repositories.ErrDatasetScan
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
		return nil, repositories.ErrDatasetQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, repositories.ErrDatasetScan
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
			return nil, repositories.ErrDatasetScan
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil {
		return nil, repositories.ErrDatasetScan
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
		return nil, repositories.ErrDatasetQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, repositories.ErrDatasetList
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
			return nil, repositories.ErrDatasetList
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil {
		return nil, repositories.ErrDatasetList
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
		return nil, repositories.ErrDatasetQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, repositories.ErrDatasetScan
	}

	var list []*entities.Dataset
	for rows.Next() {
		d := &entities.Dataset{}
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Description,
			&d.OwnerID, &d.CategoryID, &d.IsPublic,
			&d.CreatedAt,
		); err != nil {
			return nil, repositories.ErrDatasetScan
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil {
		return nil, repositories.ErrDatasetScan
	}
	return list, nil
}
