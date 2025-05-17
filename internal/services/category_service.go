package services

import (
	"context"
	"ppo/internal/entities"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, name, description string) (uint64, error)
	UpdateCategory(ctx context.Context, id uint64, name, description string) error
	DeleteCategory(ctx context.Context, id uint64) error
	GetCategoryByID(ctx context.Context, id uint64) (*entities.Category, error)
	ListCategories(ctx context.Context) ([]*entities.Category, error)
}
