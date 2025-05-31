package postqbuild

import (
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type MetadataRepo struct {
	db *pgxpool.Pool
}

func NewMetadataRepo(pool *pgxpool.Pool) *MetadataRepo {
	return &MetadataRepo{db: pool}
}

func (r *MetadataRepo) Create(ctx context.Context, m *entities.Metadata) error {
	query := psql.
		Insert("metadata").
		Columns("format", "size", "tags", "dataset_version_id").
		Values(m.Format, m.Size, m.Tags, m.DatasetVersionID).
		Suffix("RETURNING id")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return repositories.ErrMetadataQueryBuild
	}

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&m.ID)
	if err != nil {
		return repositories.ErrMetadataCreate
	}
	return nil
}

func (r *MetadataRepo) Update(ctx context.Context, m *entities.Metadata) error {
	query := psql.
		Update("metadata").
		Set("format", m.Format).
		Set("size", m.Size).
		Set("tags", m.Tags).
		Where(sq.Eq{"id": m.ID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return repositories.ErrMetadataQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return repositories.ErrMetadataUpdate
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrMetadataNotFound
	}
	return nil
}

func (r *MetadataRepo) Delete(ctx context.Context, id uint64) error {
	query := psql.Delete("metadata").Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return repositories.ErrMetadataQueryBuild
	}

	cmd, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return repositories.ErrMetadataDelete
	}
	if cmd.RowsAffected() == 0 {
		return repositories.ErrMetadataNotFound
	}
	return nil
}

func (r *MetadataRepo) FindByID(ctx context.Context, id uint64) (*entities.Metadata, error) {
	query := psql.
		Select("id", "format", "size", "tags", "dataset_version_id").
		From("metadata").
		Where(sq.Eq{"id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, repositories.ErrMetadataQueryBuild
	}

	m := &entities.Metadata{}
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&m.ID,
		&m.Format,
		&m.Size,
		&m.Tags,
		&m.DatasetVersionID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrMetadataNotFound
		}
		return nil, repositories.ErrMetadataScan
	}
	return m, nil
}

func (r *MetadataRepo) FindByDatasetID(ctx context.Context, datasetVersionID uint64) ([]*entities.Metadata, error) {
	query := psql.
		Select("id", "format", "size", "tags", "dataset_version_id").
		From("metadata").
		Where(sq.Eq{"dataset_version_id": datasetVersionID}).
		OrderBy("id DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, repositories.ErrMetadataQueryBuild
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, repositories.ErrMetadataList
	}
	defer rows.Close()

	var list []*entities.Metadata
	for rows.Next() {
		m := &entities.Metadata{}
		if err := rows.Scan(
			&m.ID,
			&m.Format,
			&m.Size,
			&m.Tags,
			&m.DatasetVersionID,
		); err != nil {
			return nil, repositories.ErrMetadataScan
		}
		list = append(list, m)
	}
	if err := rows.Err(); err != nil {
		return nil, repositories.ErrMetadataList
	}
	return list, nil
}
