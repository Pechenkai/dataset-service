package postqbuild

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type CategoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{db: pool}
}

func (r *CategoryRepo) Create(ctx context.Context, c *entities.Category) error {
	query := psql.
		Insert("categories").
		Columns("name", "description").
		Values(c.Name, c.Description).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%w: %v", repositories.ErrCategoryQueryBuild, err)
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&c.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repositories.ErrCategoryAlreadyExists
		}
		return fmt.Errorf("%w: %v", repositories.ErrCategoryCreate, err)
	}
	return nil
}

func (r *CategoryRepo) Update(ctx context.Context, c *entities.Category) error {
	query := psql.
		Update("categories").
		Set("name", c.Name).
		Set("description", c.Description).
		Where(sq.Eq{"id": c.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%w: %v", repositories.ErrCategoryQueryBuild, err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repositories.ErrCategoryAlreadyExists
		}
		return fmt.Errorf("%w: %v", repositories.ErrCategoryUpdate, err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id uint64) error {
	query := psql.Delete("categories").Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%w: %v", repositories.ErrCategoryQueryBuild, err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return repositories.ErrCategoryNotEmpty
		}
		return fmt.Errorf("%w: %v", repositories.ErrCategoryDelete, err)
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepo) FindByID(ctx context.Context, id uint64) (*entities.Category, error) {
	query := psql.
		Select("id", "name", "description").
		From("categories").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryQueryBuild, err)
	}

	c := &entities.Category{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&c.ID, &c.Name, &c.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryFind, err)
	}
	return c, nil
}

func (r *CategoryRepo) FindAll(ctx context.Context) ([]*entities.Category, error) {
	query := psql.
		Select("id", "name", "description").
		From("categories").
		OrderBy("name ASC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryQueryBuild, err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryFind, err)
	}
	defer rows.Close()

	var list []*entities.Category
	for rows.Next() {
		c := &entities.Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Description); err != nil {
			return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryScan, err)
		}
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryIterate, err)
	}
	return list, nil
}
