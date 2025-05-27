package services

import (
	"context"
	"ppo/internal/entities"
)

type CreateCategoryCmd struct {
	Name        string
	Description string
}

type UpdateCategoryCmd struct {
	Name        string
	Description string
	ID          uint64
}

type CategoryService interface {
	CreateCategory(ctx context.Context, cmd CreateCategoryCmd) (uint64, error)
	UpdateCategory(ctx context.Context, cmd UpdateCategoryCmd) error
	DeleteCategory(ctx context.Context, id uint64) error
	GetCategoryByID(ctx context.Context, id uint64) (*entities.Category, error)
	ListCategories(ctx context.Context) ([]*entities.Category, error)
}
