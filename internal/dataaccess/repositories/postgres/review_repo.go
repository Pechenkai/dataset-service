package postqbuild

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/repositories"

	"ppo/internal/entities"
)

type ReviewRepo struct {
	db *pgxpool.Pool
}

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{db: pool}
}

func (r *ReviewRepo) Create(ctx context.Context, rv *entities.Review) error {
	const sql = `
	INSERT INTO reviews
	  (user_id, dataset_id, rating, created_at, text)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
	`
	if rv.CreatedAt.IsZero() {
		rv.CreatedAt = time.Now().UTC()
	}
	err := r.db.QueryRow(ctx, sql,
		rv.UserID,
		rv.DatasetID,
		rv.Rating,
		rv.CreatedAt,
		rv.Text,
	).Scan(&rv.ID)
	if err != nil {
		return fmt.Errorf("create review: %w", err)
	}
	return nil
}

func (r *ReviewRepo) Update(ctx context.Context, rv *entities.Review) error {
	const sql = `
	UPDATE reviews
	SET rating = $1,
	    text = $2
	WHERE id = $3
	`
	cmd, err := r.db.Exec(ctx, sql,
		rv.Rating,
		rv.Text,
		rv.ID,
	)
	if err != nil {
		return fmt.Errorf("update review: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrReviewNotFound
	}
	return nil
}

func (r *ReviewRepo) Delete(ctx context.Context, id uint64) error {
	const sql = `DELETE FROM reviews WHERE id = $1`
	exec, err := r.db.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("delete review: %w", err)
	}
	if exec.RowsAffected() == 0 {
		return repositories.ErrReviewNotFound
	}
	return nil
}

func (r *ReviewRepo) FindByID(ctx context.Context, id uint64) (*entities.Review, error) {
	const sql = `
	SELECT id, user_id, dataset_id, rating, created_at, text
	FROM reviews
	WHERE id = $1
	`
	rv := &entities.Review{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&rv.ID,
		&rv.UserID,
		&rv.DatasetID,
		&rv.Rating,
		&rv.CreatedAt,
		&rv.Text,
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
	const sql = `
	SELECT id, user_id, dataset_id, rating, created_at, text
	FROM reviews
	WHERE dataset_id = $1
	ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, sql, datasetID)
	if err != nil {
		return nil, fmt.Errorf("query reviews by dataset id: %w", err)
	}
	defer rows.Close()

	var list []*entities.Review
	for rows.Next() {
		rv := &entities.Review{}
		if err := rows.Scan(
			&rv.ID,
			&rv.UserID,
			&rv.DatasetID,
			&rv.Rating,
			&rv.CreatedAt,
			&rv.Text,
		); err != nil {
			return nil, fmt.Errorf("scan review row: %w", err)
		}
		list = append(list, rv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate review rows: %w", err)
	}
	return list, nil
}

func (r *ReviewRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Review, error) {
	const sql = `
	SELECT id, user_id, dataset_id, rating, created_at, text
	FROM reviews
	WHERE user_id = $1
	ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("query reviews by user id: %w", err)
	}
	defer rows.Close()

	var list []*entities.Review
	for rows.Next() {
		rv := &entities.Review{}
		if err := rows.Scan(
			&rv.ID,
			&rv.UserID,
			&rv.DatasetID,
			&rv.Rating,
			&rv.CreatedAt,
			&rv.Text,
		); err != nil {
			return nil, fmt.Errorf("scan review row: %w", err)
		}
		list = append(list, rv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate review rows: %w", err)
	}
	return list, nil
}
