package repositories

import (
	"context"
	"ppo/internal/entities"
)

type DatasetRepository interface {
	Create(ctx context.Context, d *entities.Dataset) error
	Delete(ctx context.Context, id uint64) error
	Update(ctx context.Context, d *entities.Dataset) error
	FindByID(ctx context.Context, id uint64) (*entities.Dataset, error)
	FindByUserID(ctx context.Context, userID uint64) ([]*entities.Dataset, error)
	FindAll(ctx context.Context) ([]*entities.Dataset, error)
	FindPublic(ctx context.Context) ([]*entities.Dataset, error)
}
