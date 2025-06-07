package postqbuild

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type DatasetRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewDatasetRepo(pool *pgxpool.Pool, logger *zap.Logger) *DatasetRepo {
	logger.Debug("NewDatasetRepo initialized")
	return &DatasetRepo{db: pool, logger: logger}
}

func (r *DatasetRepo) Create(ctx context.Context, d *entities.Dataset) error {
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now().UTC()
	}
	r.logger.Debug("Create Dataset called",
		zap.String("name", d.Name),
		zap.String("description", d.Description),
		zap.Uint64("owner_id", d.OwnerID),
		zap.Uint64("category_id", d.CategoryID),
		zap.Bool("is_public", d.IsPublic),
	)

	query := psql.
		Insert("datasets").
		Columns("name", "description", "owner_id", "category_id", "is_public", "created_at").
		Values(d.Name, d.Description, d.OwnerID, d.CategoryID, d.IsPublic, d.CreatedAt).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build create dataset query",
			zap.Error(err),
		)
		return repositories.ErrDatasetQueryBuild
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&d.ID)
	if err != nil {
		r.logger.Error("failed to execute create dataset query",
			zap.Error(err),
		)
		return repositories.ErrDatasetCreate
	}

	r.logger.Info("dataset created successfully",
		zap.Uint64("id", d.ID),
		zap.String("name", d.Name),
	)
	return nil
}

func (r *DatasetRepo) Update(ctx context.Context, d *entities.Dataset) error {
	r.logger.Debug("Update Dataset called",
		zap.Uint64("id", d.ID),
		zap.String("name", d.Name),
		zap.String("description", d.Description),
		zap.Uint64("category_id", d.CategoryID),
		zap.Bool("is_public", d.IsPublic),
	)

	query := psql.
		Update("datasets").
		Set("name", d.Name).
		Set("description", d.Description).
		Set("category_id", d.CategoryID).
		Set("is_public", d.IsPublic).
		Where(sq.Eq{"id": d.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build update dataset query",
			zap.Error(err),
			zap.Uint64("id", d.ID),
		)
		return repositories.ErrDatasetQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute update dataset query",
			zap.Error(err),
			zap.Uint64("id", d.ID),
		)
		return repositories.ErrDatasetUpdate
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no dataset found to update",
			zap.Uint64("id", d.ID),
		)
		return repositories.ErrDatasetNotFound
	}

	r.logger.Info("dataset updated successfully",
		zap.Uint64("id", d.ID),
		zap.String("name", d.Name),
	)
	return nil
}

func (r *DatasetRepo) Delete(ctx context.Context, id uint64) error {
	r.logger.Debug("Delete Dataset called",
		zap.Uint64("id", id),
	)

	query := psql.Delete("datasets").Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build delete dataset query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return repositories.ErrDatasetQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute delete dataset query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return repositories.ErrDatasetDelete
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no dataset found to delete",
			zap.Uint64("id", id),
		)
		return repositories.ErrDatasetNotFound
	}

	r.logger.Info("dataset deleted successfully",
		zap.Uint64("id", id),
	)
	return nil
}

func (r *DatasetRepo) FindByID(ctx context.Context, id uint64) (*entities.Dataset, error) {
	r.logger.Debug("FindByID Dataset called",
		zap.Uint64("id", id),
	)

	query := psql.
		Select("id", "name", "description", "owner_id", "category_id", "is_public", "created_at").
		From("datasets").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find dataset by ID query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
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
			r.logger.Warn("dataset not found by ID",
				zap.Uint64("id", id),
			)
			return nil, repositories.ErrDatasetNotFound
		}
		r.logger.Error("failed to execute find dataset by ID query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, repositories.ErrDatasetScan
	}

	r.logger.Info("dataset fetched successfully",
		zap.Uint64("id", d.ID),
		zap.String("name", d.Name),
	)
	return d, nil
}

func (r *DatasetRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Dataset, error) {
	r.logger.Debug("FindByUserID Datasets called",
		zap.Uint64("user_id", userID),
	)

	query := psql.
		Select("id", "name", "description", "owner_id", "category_id", "is_public", "created_at").
		From("datasets").
		Where(sq.Eq{"owner_id": userID}).
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find datasets by user query",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		return nil, repositories.ErrDatasetQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute find datasets by user query",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
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
			r.logger.Error("failed to scan dataset row",
				zap.Error(err),
			)
			return nil, repositories.ErrDatasetScan
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating over dataset rows",
			zap.Error(err),
		)
		return nil, repositories.ErrDatasetScan
	}

	r.logger.Info("datasets fetched by user successfully",
		zap.Uint64("user_id", userID),
		zap.Int("count", len(list)),
	)
	return list, nil
}

func (r *DatasetRepo) FindAll(ctx context.Context) ([]*entities.Dataset, error) {
	r.logger.Debug("FindAll Datasets called")

	query := psql.
		Select("id", "name", "description", "owner_id", "category_id", "is_public", "created_at").
		From("datasets").
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find all datasets query",
			zap.Error(err),
		)
		return nil, repositories.ErrDatasetQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute find all datasets query",
			zap.Error(err),
		)
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
			r.logger.Error("failed to scan dataset row",
				zap.Error(err),
			)
			return nil, repositories.ErrDatasetList
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating over dataset rows",
			zap.Error(err),
		)
		return nil, repositories.ErrDatasetList
	}

	r.logger.Info("all datasets fetched successfully",
		zap.Int("count", len(list)),
	)
	return list, nil
}

func (r *DatasetRepo) FindPublic(ctx context.Context) ([]*entities.Dataset, error) {
	r.logger.Debug("FindPublic Datasets called")

	query := psql.
		Select("id", "name", "description", "owner_id", "category_id", "is_public", "created_at").
		From("datasets").
		Where(sq.Eq{"is_public": true})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find public datasets query",
			zap.Error(err),
		)
		return nil, repositories.ErrDatasetQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute find public datasets query",
			zap.Error(err),
		)
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
			r.logger.Error("failed to scan public dataset row",
				zap.Error(err),
			)
			return nil, repositories.ErrDatasetScan
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating over public dataset rows",
			zap.Error(err),
		)
		return nil, repositories.ErrDatasetScan
	}

	r.logger.Info("public datasets fetched successfully",
		zap.Int("count", len(list)),
	)
	return list, nil
}

func (r *DatasetRepo) FindByCategoryID(ctx context.Context, categoryID uint64) ([]*entities.Dataset, error) {
	r.logger.Debug("FindByCategoryID Datasets called", zap.Uint64("category_id", categoryID))

	query := psql.
		Select("id", "name", "description", "owner_id", "category_id", "is_public", "created_at").
		From("datasets").
		Where(sq.Eq{"category_id": categoryID}).
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find datasets by category query", zap.Error(err), zap.Uint64("category_id", categoryID))
		return nil, repositories.ErrDatasetQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute find datasets by category query", zap.Error(err), zap.Uint64("category_id", categoryID))
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
			r.logger.Error("failed to scan dataset row", zap.Error(err))
			return nil, repositories.ErrDatasetScan
		}
		list = append(list, d)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating over dataset rows", zap.Error(err))
		return nil, repositories.ErrDatasetScan
	}

	r.logger.Info("datasets fetched by category successfully", zap.Uint64("category_id", categoryID), zap.Int("count", len(list)))
	return list, nil
}
