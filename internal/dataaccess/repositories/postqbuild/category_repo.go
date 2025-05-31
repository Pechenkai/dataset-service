package postqbuild

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type CategoryRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewCategoryRepo создаёт экземпляр CategoryRepo с привязанным логгером.
func NewCategoryRepo(pool *pgxpool.Pool, logger *zap.Logger) *CategoryRepo {
	logger.Debug("NewCategoryRepo initialized")
	return &CategoryRepo{
		db:     pool,
		logger: logger,
	}
}

func (r *CategoryRepo) Create(ctx context.Context, c *entities.Category) error {
	r.logger.Debug("Create Category called",
		zap.String("name", c.Name),
		zap.String("description", c.Description),
	)

	query := psql.
		Insert("categories").
		Columns("name", "description").
		Values(c.Name, c.Description).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build create category query",
			zap.Error(err),
			zap.String("name", c.Name),
		)
		return fmt.Errorf("%w: %v", repositories.ErrCategoryQueryBuild, err)
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&c.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			r.logger.Warn("category already exists on create",
				zap.String("name", c.Name),
				zap.String("constraint", pgErr.ConstraintName),
			)
			return repositories.ErrCategoryAlreadyExists
		}
		r.logger.Error("failed to execute create category query",
			zap.Error(err),
			zap.String("name", c.Name),
		)
		return fmt.Errorf("%w: %v", repositories.ErrCategoryCreate, err)
	}

	r.logger.Info("category created successfully",
		zap.Uint64("id", c.ID),
		zap.String("name", c.Name),
	)
	return nil
}

func (r *CategoryRepo) Update(ctx context.Context, c *entities.Category) error {
	r.logger.Debug("Update Category called",
		zap.Uint64("id", c.ID),
		zap.String("name", c.Name),
		zap.String("description", c.Description),
	)

	query := psql.
		Update("categories").
		Set("name", c.Name).
		Set("description", c.Description).
		Where(sq.Eq{"id": c.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build update category query",
			zap.Error(err),
			zap.Uint64("id", c.ID),
		)
		return fmt.Errorf("%w: %v", repositories.ErrCategoryQueryBuild, err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			r.logger.Warn("category already exists on update",
				zap.String("name", c.Name),
				zap.String("constraint", pgErr.ConstraintName),
			)
			return repositories.ErrCategoryAlreadyExists
		}
		r.logger.Error("failed to execute update category query",
			zap.Error(err),
			zap.Uint64("id", c.ID),
		)
		return fmt.Errorf("%w: %v", repositories.ErrCategoryUpdate, err)
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no category found to update",
			zap.Uint64("id", c.ID),
		)
		return repositories.ErrCategoryNotFound
	}

	r.logger.Info("category updated successfully",
		zap.Uint64("id", c.ID),
		zap.String("name", c.Name),
	)
	return nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id uint64) error {
	r.logger.Debug("Delete Category called",
		zap.Uint64("id", id),
	)

	query := psql.Delete("categories").Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build delete category query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return fmt.Errorf("%w: %v", repositories.ErrCategoryQueryBuild, err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			r.logger.Warn("category not empty, cannot delete",
				zap.Uint64("id", id),
				zap.String("constraint", pgErr.ConstraintName),
			)
			return repositories.ErrCategoryNotEmpty
		}
		r.logger.Error("failed to execute delete category query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return fmt.Errorf("%w: %v", repositories.ErrCategoryDelete, err)
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no category found to delete",
			zap.Uint64("id", id),
		)
		return repositories.ErrCategoryNotFound
	}

	r.logger.Info("category deleted successfully",
		zap.Uint64("id", id),
	)
	return nil
}

func (r *CategoryRepo) FindByID(ctx context.Context, id uint64) (*entities.Category, error) {
	r.logger.Debug("FindByID Category called",
		zap.Uint64("id", id),
	)

	query := psql.
		Select("id", "name", "description").
		From("categories").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find category by ID query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryQueryBuild, err)
	}

	c := &entities.Category{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&c.ID, &c.Name, &c.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("category not found by ID",
				zap.Uint64("id", id),
			)
			return nil, repositories.ErrCategoryNotFound
		}
		r.logger.Error("failed to execute find category by ID query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryFind, err)
	}

	r.logger.Info("category fetched successfully",
		zap.Uint64("id", c.ID),
		zap.String("name", c.Name),
	)
	return c, nil
}

func (r *CategoryRepo) FindAll(ctx context.Context) ([]*entities.Category, error) {
	r.logger.Debug("FindAll Categories called")

	query := psql.
		Select("id", "name", "description").
		From("categories").
		OrderBy("name ASC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find all categories query",
			zap.Error(err),
		)
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryQueryBuild, err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute find all categories query",
			zap.Error(err),
		)
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryFind, err)
	}
	defer rows.Close()

	var list []*entities.Category
	for rows.Next() {
		c := &entities.Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Description); err != nil {
			r.logger.Error("failed to scan category row",
				zap.Error(err),
			)
			return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryScan, err)
		}
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating over category rows",
			zap.Error(err),
		)
		return nil, fmt.Errorf("%w: %v", repositories.ErrCategoryIterate, err)
	}

	r.logger.Info("categories fetched successfully",
		zap.Int("count", len(list)),
	)
	return list, nil
}
