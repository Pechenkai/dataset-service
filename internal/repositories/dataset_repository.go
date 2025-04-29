package repositories

import "ppo/internal/entities"

type DatasetRepository interface {
	Create(dataset *entities.Dataset) error
	Delete(id uint64) error
	Update(dataset *entities.Dataset) error
	FindByID(datasetID uint64) (*entities.Dataset, error)
	FindByUserID(userID uint64) ([]*entities.Dataset, error)
	FindAll() ([]*entities.Dataset, error)
}
