package repositories

import (
	"context"
	"ppo/internal/entities"
)

type DatasetVersionRepository interface {
	Create(ctx context.Context, v *entities.DatasetVersion) error
	Delete(ctx context.Context, id uint64) error
	Update(ctx context.Context, v *entities.DatasetVersion) error
	FindByID(ctx context.Context, id uint64) (*entities.DatasetVersion, error)
	FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.DatasetVersion, error)
}
