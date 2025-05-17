package repositories

import (
	"context"
	"ppo/internal/entities"
)

type MetadataRepository interface {
	Create(ctx context.Context, m *entities.Metadata) error
	Update(ctx context.Context, m *entities.Metadata) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*entities.Metadata, error)
	FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Metadata, error)
}
