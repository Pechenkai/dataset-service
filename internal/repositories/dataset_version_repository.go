package repositories

import "ppo/internal/entities"

type DatasetVersionRepository interface {
	Create(datasetVersion *entities.DatasetVersion) error
	Delete(id uint64) error
	Update(datasetVersion *entities.DatasetVersion) error
	FindByID(id uint64) (*entities.DatasetVersion, error)
	FindByDatasetID(datasetID uint64) ([]*entities.DatasetVersion, error)
}
