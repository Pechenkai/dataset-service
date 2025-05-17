package services

import (
	"context"
	"fmt"
	"time"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type datasetService struct {
	dsRepo  repositories.DatasetRepository
	verRepo repositories.DatasetVersionRepository
	mdRepo  repositories.MetadataRepository
}

func NewDatasetService(
	dsRepo repositories.DatasetRepository,
	verRepo repositories.DatasetVersionRepository,
	mdRepo repositories.MetadataRepository,
) DatasetService {
	return &datasetService{dsRepo: dsRepo, verRepo: verRepo, mdRepo: mdRepo}
}

func (s *datasetService) CreateDataset(ctx context.Context, cmd CreateDatasetCmd) (uint64, error) {
	ds, err := entities.NewDataset(cmd.Name, cmd.Description, cmd.ActorID, cmd.CategoryID, cmd.IsPublic, time.Now())
	if err != nil {
		return 0, fmt.Errorf("invalid dataset: %w", err)
	}
	if err := s.dsRepo.Create(ctx, ds); err != nil {
		return 0, fmt.Errorf("create dataset: %w", err)
	}

	ver, err := entities.NewDatasetVersion("v0.1", "", "", ds.ID, time.Now())
	if err != nil {
		return 0, fmt.Errorf("init version: %w", err)
	}
	if err := s.verRepo.Create(ctx, ver); err != nil {
		return 0, fmt.Errorf("create version: %w", err)
	}

	if cmd.MetaFormat != "" || cmd.MetaSize != 0 {
		md, err := entities.NewMetadata(cmd.MetaFormat, cmd.MetaTags, cmd.MetaSize, ver.ID)
		if err != nil {
			return 0, fmt.Errorf("%w: %s", ErrInvalidMetadata, err)
		}
		if err := s.mdRepo.Create(ctx, md); err != nil {
			return 0, fmt.Errorf("create metadata: %w", err)
		}
	}

	return ds.ID, nil
}
func (s *datasetService) AddDatasetVersion(ctx context.Context, cmd AddVersionCmd) (uint64, error) {
	ds, err := s.dsRepo.FindByID(ctx, cmd.DatasetID)
	if err != nil {
		return 0, fmt.Errorf("fetch dataset: %w", err)
	}
	if ds == nil {
		return 0, ErrDatasetNotFound
	}

	vers, err := s.verRepo.FindByDatasetID(ctx, cmd.DatasetID)
	if err != nil {
		return 0, fmt.Errorf("fetch versions: %w", err)
	}
	next := nextVersionNumber(vers)

	ver, err := entities.NewDatasetVersion(next, cmd.FilePath, cmd.ChangeLog, cmd.DatasetID, time.Now())
	if err != nil {
		return 0, fmt.Errorf("invalid version data: %w", err)
	}

	if err := s.verRepo.Create(ctx, ver); err != nil {
		return 0, fmt.Errorf("create version: %w", err)
	}

	if cmd.MetaFormat != "" || cmd.MetaSize != 0 {
		md, err := entities.NewMetadata(cmd.MetaFormat, cmd.MetaTags, cmd.MetaSize, ver.ID)
		if err != nil {
			return 0, fmt.Errorf("%w: %s", ErrInvalidMetadata, err)
		}
		if err := s.mdRepo.Create(ctx, md); err != nil {
			return 0, fmt.Errorf("create metadata: %w", err)
		}
	}

	return ver.ID, nil
}

func (s *datasetService) GetDataset(ctx context.Context, id uint64) (*entities.Dataset, error) {
	ds, err := s.dsRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get dataset: %w", err)
	}
	if ds == nil {
		return nil, ErrDatasetNotFound
	}
	return ds, nil
}

func nextVersionNumber(existing []*entities.DatasetVersion) string {
	if len(existing) == 0 {
		return "v1.0"
	}
	last := existing[0].Number
	return last + ".1"
}
