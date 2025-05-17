package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ppo/internal/entities"
)

type CategoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{db: pool}
}

func (r *CategoryRepo) Create(ctx context.Context, c *entities.Category) error {
	const sql = `
	INSERT INTO categories (name, description)
	VALUES ($1, $2)
	RETURNING id
	`
	err := r.db.QueryRow(ctx, sql, c.Name, c.Description).Scan(&c.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrCategoryAlreadyExists
		}
		return fmt.Errorf("create category: %w", err)
	}
	return nil
}

func (r *CategoryRepo) Update(ctx context.Context, c *entities.Category) error {
	const sql = `
	UPDATE categories
	SET name = $1,
	    description = $2
	WHERE id = $3
	`
	cmd, err := r.db.Exec(ctx, sql, c.Name, c.Description, c.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrCategoryAlreadyExists
		}
		return fmt.Errorf("update category: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id uint64) error {
	const sql = `DELETE FROM categories WHERE id = $1`
	exec, err := r.db.Exec(ctx, sql, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrCategoryNotEmpty
		}
		return fmt.Errorf("delete category: %w", err)
	}
	if exec.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepo) FindByID(ctx context.Context, id uint64) (*entities.Category, error) {
	const sql = `
	SELECT id, name, description
	FROM categories
	WHERE id = $1
	`
	c := &entities.Category{}
	err := r.db.QueryRow(ctx, sql, id).Scan(&c.ID, &c.Name, &c.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("find category by id: %w", err)
	}
	return c, nil
}

func (r *CategoryRepo) FindAll(ctx context.Context) ([]*entities.Category, error) {
	const sql = `
	SELECT id, name, description
	FROM categories
	ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query all categories: %w", err)
	}
	defer rows.Close()

	var list []*entities.Category
	for rows.Next() {
		c := &entities.Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Description); err != nil {
			return nil, fmt.Errorf("scan category row: %w", err)
		}
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate category rows: %w", err)
	}
	return list, nil
}
