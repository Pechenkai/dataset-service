package postqbuild

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type ReviewRepo struct {
	db *pgxpool.Pool
}

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{db: pool}
}

func (r *ReviewRepo) Create(ctx context.Context, rv *entities.Review) error {
	if rv.CreatedAt.IsZero() {
		rv.CreatedAt = time.Now().UTC()
	}
	query := psql.
		Insert("reviews").
		Columns("user_id", "dataset_id", "rating", "created_at", "text").
		Values(rv.UserID, rv.DatasetID, rv.Rating, rv.CreatedAt, rv.Text).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build insert review sql: %w", err)
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&rv.ID)
	if err != nil {
		return fmt.Errorf("create review: %w", err)
	}
	return nil
}

func (r *ReviewRepo) Update(ctx context.Context, rv *entities.Review) error {
	query := psql.
		Update("reviews").
		Set("rating", rv.Rating).
		Set("text", rv.Text).
		Where(sq.Eq{"id": rv.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build update review sql: %w", err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("update review: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrReviewNotFound
	}
	return nil
}

func (r *ReviewRepo) Delete(ctx context.Context, id uint64) error {
	query := psql.Delete("reviews").Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build delete review sql: %w", err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("delete review: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrReviewNotFound
	}
	return nil
}

func (r *ReviewRepo) FindByID(ctx context.Context, id uint64) (*entities.Review, error) {
	query := psql.
		Select("id", "user_id", "dataset_id", "rating", "created_at", "text").
		From("reviews").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find review by id sql: %w", err)
	}

	rv := &entities.Review{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&rv.ID, &rv.UserID, &rv.DatasetID, &rv.Rating, &rv.CreatedAt, &rv.Text,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrReviewNotFound
		}
		return nil, fmt.Errorf("find review by id: %w", err)
	}
	return rv, nil
}

func (r *ReviewRepo) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Review, error) {
	query := psql.
		Select("id", "user_id", "dataset_id", "rating", "created_at", "text").
		From("reviews").
		Where(sq.Eq{"dataset_id": datasetID}).
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find reviews by dataset sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query reviews by dataset: %w", err)
	}
	defer rows.Close()

	var list []*entities.Review
	for rows.Next() {
		r := &entities.Review{}
		if err := rows.Scan(
			&r.ID, &r.UserID, &r.DatasetID, &r.Rating, &r.CreatedAt, &r.Text,
		); err != nil {
			return nil, fmt.Errorf("scan review row: %w", err)
		}
		list = append(list, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate review rows: %w", err)
	}
	return list, nil
}

func (r *ReviewRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Review, error) {
	query := psql.
		Select("id", "user_id", "dataset_id", "rating", "created_at", "text").
		From("reviews").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find reviews by user sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query reviews by user: %w", err)
	}
	defer rows.Close()

	var list []*entities.Review
	for rows.Next() {
		r := &entities.Review{}
		if err := rows.Scan(
			&r.ID, &r.UserID, &r.DatasetID, &r.Rating, &r.CreatedAt, &r.Text,
		); err != nil {
			return nil, fmt.Errorf("scan review row: %w", err)
		}
		list = append(list, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate review rows: %w", err)
	}
	return list, nil
}
