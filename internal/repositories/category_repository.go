package repositories

import (
	"context"
	"ppo/internal/entities"
)

type CategoryRepository interface {
	Create(ctx context.Context, c *entities.Category) error
	Delete(ctx context.Context, id uint64) error
	Update(ctx context.Context, c *entities.Category) error
	FindByID(ctx context.Context, id uint64) (*entities.Category, error)
	FindAll(ctx context.Context) ([]*entities.Category, error)
}
