package services

import "ppo/internal/entities"

type CategoryService interface {
	CreateCategory(category *entities.Category) error

	UpdateCategory(category *entities.Category) error

	DeleteCategory(id uint64) error

	GetCategoryByID(id uint64) (*entities.Category, error)

	GetAllCategories() ([]*entities.Category, error)
}
