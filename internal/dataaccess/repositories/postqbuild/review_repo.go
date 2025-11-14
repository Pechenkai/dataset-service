package postqbuild

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"go.uber.org/zap"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type ReviewRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

type reviewRow struct {
	ID        uint64
	UserID    uint64
	DatasetID uint64
	Rating    int
	CreatedAt time.Time
	Text      string
}

func reviewRowFromEntity(rv *entities.Review) reviewRow {
	if rv == nil {
		return reviewRow{}
	}
	return reviewRow{
		ID:        rv.ID,
		UserID:    rv.UserID,
		DatasetID: rv.DatasetID,
		Rating:    int(rv.Rating),
		CreatedAt: rv.CreatedAt,
		Text:      rv.Text,
	}
}

func (row reviewRow) toEntity() *entities.Review {
	return &entities.Review{
		ID:        row.ID,
		UserID:    row.UserID,
		DatasetID: row.DatasetID,
		Rating:    entities.Rating(row.Rating),
		CreatedAt: row.CreatedAt,
		Text:      row.Text,
	}
}

func NewReviewRepo(pool *pgxpool.Pool, logger *zap.Logger) *ReviewRepo {
	logger.Debug("NewReviewRepo initialized")
	return &ReviewRepo{db: pool, logger: logger}
}

func (r *ReviewRepo) Create(ctx context.Context, rv *entities.Review) error {
	row := reviewRowFromEntity(rv)
	if row.CreatedAt.IsZero() {
		row.CreatedAt = time.Now().UTC()
		rv.CreatedAt = row.CreatedAt
	}
	r.logger.Debug("Create Review called",
		zap.Uint64("user_id", row.UserID),
		zap.Uint64("dataset_id", row.DatasetID),
		zap.Int("rating", row.Rating),
		zap.String("text", row.Text),
	)

	query := psql.
		Insert("reviews").
		Columns(
			"user_id",
			"dataset_id",
			"rating",
			"created_at",
			"text",
		).
		Values(
			row.UserID,
			row.DatasetID,
			row.Rating,
			row.CreatedAt,
			row.Text,
		).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build insert review SQL",
			zap.Error(err),
		)
		return fmt.Errorf("build insert review sql: %w", err)
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&row.ID)
	if err != nil {
		r.logger.Error("failed to execute insert review query",
			zap.Error(err),
		)
		return fmt.Errorf("create review: %w", err)
	}

	rv.ID = row.ID
	r.logger.Info("review created successfully",
		zap.Uint64("id", row.ID),
		zap.Uint64("user_id", row.UserID),
		zap.Uint64("dataset_id", row.DatasetID),
		zap.Int("rating", row.Rating),
	)
	return nil
}

func (r *ReviewRepo) Update(ctx context.Context, rv *entities.Review) error {
	row := reviewRowFromEntity(rv)
	r.logger.Debug("Update Review called",
		zap.Uint64("id", row.ID),
		zap.Int("rating", row.Rating),
		zap.String("text", row.Text),
	)

	query := psql.
		Update("reviews").
		Set("rating", row.Rating).
		Set("text", row.Text).
		Where(sq.Eq{"id": row.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build update review SQL",
			zap.Error(err),
			zap.Uint64("id", row.ID),
		)
		return fmt.Errorf("build update review sql: %w", err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute update review query",
			zap.Error(err),
			zap.Uint64("id", row.ID),
		)
		return fmt.Errorf("update review: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no review found to update",
			zap.Uint64("id", row.ID),
		)
		return repositories.ErrReviewNotFound
	}

	r.logger.Info("review updated successfully",
		zap.Uint64("id", row.ID),
	)
	return nil
}

func (r *ReviewRepo) Delete(ctx context.Context, id uint64) error {
	r.logger.Debug("Delete Review called",
		zap.Uint64("id", id),
	)

	query := psql.Delete("reviews").Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build delete review SQL",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return fmt.Errorf("build delete review sql: %w", err)
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute delete review query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return fmt.Errorf("delete review: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no review found to delete",
			zap.Uint64("id", id),
		)
		return repositories.ErrReviewNotFound
	}

	r.logger.Info("review deleted successfully",
		zap.Uint64("id", id),
	)
	return nil
}

func (r *ReviewRepo) FindByID(ctx context.Context, id uint64) (*entities.Review, error) {
	r.logger.Debug("FindByID Review called",
		zap.Uint64("id", id),
	)

	query := psql.
		Select("id", "user_id", "dataset_id", "rating", "created_at", "text").
		From("reviews").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find review by ID SQL",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, fmt.Errorf("build find review by id sql: %w", err)
	}

	row := &reviewRow{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&row.ID,
		&row.UserID,
		&row.DatasetID,
		&row.Rating,
		&row.CreatedAt,
		&row.Text,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("review not found by ID",
				zap.Uint64("id", id),
			)
			return nil, repositories.ErrReviewNotFound
		}
		r.logger.Error("failed to execute find review by ID query",
			zap.Error(err),
			zap.Uint64("id", id),
		)
		return nil, fmt.Errorf("find review by id: %w", err)
	}

	entity := row.toEntity()
	r.logger.Info("review fetched successfully",
		zap.Uint64("id", entity.ID),
		zap.Uint64("user_id", entity.UserID),
		zap.Uint64("dataset_id", entity.DatasetID),
		zap.Int("rating", int(entity.Rating)),
	)
	return entity, nil
}

func (r *ReviewRepo) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Review, error) {
	r.logger.Debug("FindByDatasetID Reviews called",
		zap.Uint64("dataset_id", datasetID),
	)

	query := psql.
		Select("id", "user_id", "dataset_id", "rating", "created_at", "text").
		From("reviews").
		Where(sq.Eq{"dataset_id": datasetID}).
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find reviews by dataset SQL",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, fmt.Errorf("build find reviews by dataset sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute find reviews by dataset query",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, fmt.Errorf("query reviews by dataset: %w", err)
	}
	defer rows.Close()

	var list []*entities.Review
	for rows.Next() {
		row := reviewRow{}
		if err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.DatasetID,
			&row.Rating,
			&row.CreatedAt,
			&row.Text,
		); err != nil {
			r.logger.Error("failed to scan review row",
				zap.Error(err),
			)
			return nil, fmt.Errorf("scan review row: %w", err)
		}
		list = append(list, row.toEntity())
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating over review rows",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, fmt.Errorf("iterate review rows: %w", err)
	}

	r.logger.Info("reviews fetched by dataset successfully",
		zap.Uint64("dataset_id", datasetID),
		zap.Int("count", len(list)),
	)
	return list, nil
}

func (r *ReviewRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Review, error) {
	r.logger.Debug("FindByUserID Reviews called",
		zap.Uint64("user_id", userID),
	)

	query := psql.
		Select("id", "user_id", "dataset_id", "rating", "created_at", "text").
		From("reviews").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("failed to build find reviews by user SQL",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("build find reviews by user sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		r.logger.Error("failed to execute find reviews by user query",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("query reviews by user: %w", err)
	}
	defer rows.Close()

	var list []*entities.Review
	for rows.Next() {
		row := reviewRow{}
		if err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.DatasetID,
			&row.Rating,
			&row.CreatedAt,
			&row.Text,
		); err != nil {
			r.logger.Error("failed to scan review row",
				zap.Error(err),
			)
			return nil, fmt.Errorf("scan review row: %w", err)
		}
		list = append(list, row.toEntity())
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating over review rows",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("iterate review rows: %w", err)
	}

	r.logger.Info("reviews fetched by user successfully",
		zap.Uint64("user_id", userID),
		zap.Int("count", len(list)),
	)
	return list, nil
}
