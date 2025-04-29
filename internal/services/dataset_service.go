package services

import "ppo/internal/entities"

type DatasetService interface {
	CreateDataset(dataset *entities.Dataset, metadata *entities.Metadata) error
	UpdateDataset(dataset *entities.Dataset, changeLog string, filePath string, metadata *entities.Metadata) error
	GetDataset(id uint64) (*entities.Dataset, error)
}
