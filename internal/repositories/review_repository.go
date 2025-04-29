package repositories

import "ppo/internal/entities"

type ReviewRepository interface {
	Create(dataset *entities.Review) error
	Update(dataset *entities.Review) error
	Delete(datasetID uint64) error
	FindByID(datasetID uint64) (*entities.Review, error)
	FindByDatasetID(datasetID uint64) ([]*entities.Review, error)
	FindByUserID(userID uint64) ([]*entities.Review, error)
}
