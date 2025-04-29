package services

import (
	"time"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type datasetService struct {
	datasetRepo  repositories.DatasetRepository
	versionRepo  repositories.DatasetVersionRepository
	metadataRepo repositories.MetadataRepository
}

func NewDatasetService(
	datasetRepo repositories.DatasetRepository,
	versionRepo repositories.DatasetVersionRepository,
	metadataRepo repositories.MetadataRepository,
) DatasetService {
	return &datasetService{
		datasetRepo:  datasetRepo,
		versionRepo:  versionRepo,
		metadataRepo: metadataRepo,
	}
}

func (s *datasetService) CreateDataset(dataset *entities.Dataset, metadata *entities.Metadata) error {
	if dataset == nil {
		return ErrNilDataset
	}

	if dataset.CreatedAt.IsZero() {
		dataset.CreatedAt = time.Now()
	}

	if err := s.datasetRepo.Create(dataset); err != nil {
		return err
	}

	initVersion, err := entities.NewDatasetVersion("v0.1", "", "", dataset.ID, time.Now())

	if err != nil {
		return err
	}

	if err := s.versionRepo.Create(initVersion); err != nil {
		return err
	}

	if metadata != nil {
		metadata.DatasetVersionID = initVersion.ID

		if err := s.metadataRepo.Create(metadata); err != nil {
			return err
		}
	}

	return nil
}

func (s *datasetService) UpdateDataset(dataset *entities.Dataset, changeLog string, filePath string, metadata *entities.Metadata) error {
	if dataset == nil {
		return ErrNilDataset
	}

	if err := s.datasetRepo.Update(dataset); err != nil {
		return err
	}

	newVersion, err := entities.NewDatasetVersion(s.generateNextVersionNumber(dataset.ID), filePath, changeLog, dataset.ID, time.Now())

	if err != nil {
		return err
	}

	if err := s.versionRepo.Create(newVersion); err != nil {
		return err
	}

	if metadata != nil {
		metadata.DatasetVersionID = newVersion.ID
		if err := s.metadataRepo.Create(metadata); err != nil {
			return err
		}
	}

	return nil
}

func (s *datasetService) GetDataset(id uint64) (*entities.Dataset, error) {
	return s.datasetRepo.FindByID(id)
}

func (s *datasetService) generateNextVersionNumber(datasetID uint64) string {
	versions, err := s.versionRepo.FindByDatasetID(datasetID)
	if err != nil || len(versions) == 0 {
		return "v1.0"
	}
	lastVersion := versions[len(versions)-1].Number
	return lastVersion + "+"
}
