package repositories

import "ppo/internal/entities"

type CategoryRepository interface {
	Create(category *entities.Category) error
	Delete(id uint64) error
	Update(category *entities.Category) error
	FindByID(id uint64) (*entities.Category, error)
	FindAll() ([]*entities.Category, error)
}
