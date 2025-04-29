package repositories

import "ppo/internal/entities"

type MetadataRepository interface {
	Create(dataset *entities.Metadata) error
	Update(dataset *entities.Metadata) error
	Delete(id uint64) error
	FindByID(id uint64) (*entities.Metadata, error)
	FindByDatasetID(datasetID uint64) ([]*entities.Metadata, error)
}
