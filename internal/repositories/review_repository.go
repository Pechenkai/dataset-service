package repositories

import (
	"context"
	"ppo/internal/entities"
)

type ReviewRepository interface {
	Create(ctx context.Context, rv *entities.Review) error
	Update(ctx context.Context, rv *entities.Review) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*entities.Review, error)
	FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Review, error)
	FindByUserID(ctx context.Context, userID uint64) ([]*entities.Review, error)
}
