package services

import (
	"context"
	"fmt"
	"io"
	"ppo/internal/storage"
	"time"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type datasetService struct {
	dsRepo  repositories.DatasetRepository
	verRepo repositories.DatasetVersionRepository
	mdRepo  repositories.MetadataRepository
	storage storage.Storage
}

func NewDatasetService(
	dsRepo repositories.DatasetRepository,
	verRepo repositories.DatasetVersionRepository,
	mdRepo repositories.MetadataRepository,
	storage storage.Storage,
) DatasetService {
	return &datasetService{dsRepo: dsRepo, verRepo: verRepo, mdRepo: mdRepo, storage: storage}
}

func (s *datasetService) CreateDataset(ctx context.Context, cmd CreateDatasetCmd, r io.Reader, size int64) (uint64, error) {
	ds, err := entities.NewDataset(cmd.Name, cmd.Description, cmd.ActorID, cmd.CategoryID, cmd.IsPublic, time.Now())
	if err != nil {
		return 0, fmt.Errorf("invalid dataset: %w", err)
	}
	if err := s.dsRepo.Create(ctx, ds); err != nil {
		return 0, fmt.Errorf("create dataset: %w", err)
	}

	key := fmt.Sprintf("datasets/%d/%s", ds.ID, cmd.FileName)
	url, err := s.storage.Upload(ctx, key, r, size)
	if err != nil {
		_ = s.dsRepo.Delete(ctx, ds.ID)
		return 0, fmt.Errorf("upload file: %w", err)
	}

	ver, err := entities.NewDatasetVersion("v0.1", url, "", ds.ID, time.Now())
	if err != nil {
		_ = s.storage.Delete(ctx, key)
		_ = s.dsRepo.Delete(ctx, ds.ID)
		return 0, fmt.Errorf("invalid version: %w", err)
	}

	if err := s.verRepo.Create(ctx, ver); err != nil {
		_ = s.storage.Delete(ctx, key)
		_ = s.dsRepo.Delete(ctx, ds.ID)
		return 0, fmt.Errorf("create version failed: %w", err)
	}

	if cmd.MetaFormat != "" || cmd.MetaSize != 0 {
		md, err := entities.NewMetadata(cmd.MetaFormat, cmd.MetaTags, cmd.MetaSize, ver.ID)
		if err != nil {
			_ = s.storage.Delete(ctx, key)
			_ = s.dsRepo.Delete(ctx, ds.ID)
			_ = s.verRepo.Delete(ctx, ver.ID)
			return 0, fmt.Errorf("%w: %s", ErrInvalidMetadata, err)
		}
		if err := s.mdRepo.Create(ctx, md); err != nil {
			_ = s.storage.Delete(ctx, key)
			_ = s.dsRepo.Delete(ctx, ds.ID)
			_ = s.verRepo.Delete(ctx, ver.ID)
			return 0, fmt.Errorf("create metadata: %w", err)
		}
	}

	return ds.ID, nil
}

func (s *datasetService) AddDatasetVersion(ctx context.Context, cmd AddVersionCmd, r io.Reader, size int64) (uint64, error) {
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

	key := fmt.Sprintf("datasets/%d/%s", ds.ID, cmd.FileName)
	url, err := s.storage.Upload(ctx, key, r, size)
	if err != nil {
		_ = s.dsRepo.Delete(ctx, ds.ID)
		return 0, fmt.Errorf("upload file: %w", err)
	}

	ver, err := entities.NewDatasetVersion(next, url, cmd.ChangeLog, cmd.DatasetID, time.Now())
	if err != nil {
		_ = s.storage.Delete(ctx, key)
		_ = s.dsRepo.Delete(ctx, ds.ID)
		return 0, fmt.Errorf("invalid version data: %w", err)
	}

	if err := s.verRepo.Create(ctx, ver); err != nil {
		_ = s.storage.Delete(ctx, key)
		_ = s.dsRepo.Delete(ctx, ds.ID)
		return 0, fmt.Errorf("create version: %w", err)
	}

	if cmd.MetaFormat != "" || cmd.MetaSize != 0 {
		md, err := entities.NewMetadata(cmd.MetaFormat, cmd.MetaTags, cmd.MetaSize, ver.ID)
		if err != nil {
			_ = s.storage.Delete(ctx, key)
			_ = s.dsRepo.Delete(ctx, ds.ID)
			_ = s.verRepo.Delete(ctx, ver.ID)
			return 0, fmt.Errorf("%w: %s", ErrInvalidMetadata, err)
		}
		if err := s.mdRepo.Create(ctx, md); err != nil {
			_ = s.storage.Delete(ctx, key)
			_ = s.dsRepo.Delete(ctx, ds.ID)
			_ = s.verRepo.Delete(ctx, ver.ID)
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

func (s *datasetService) ListDatasets(ctx context.Context, onlyPublic bool, ownerID *uint64) ([]*entities.Dataset, error) {
	if onlyPublic && ownerID == nil {
		return s.dsRepo.FindPublic(ctx)
	}
	if ownerID != nil {
		return s.dsRepo.FindByUserID(ctx, *ownerID)
	}
	return s.dsRepo.FindAll(ctx)
}

func (s *datasetService) GetVersion(ctx context.Context, versionID uint64) (*entities.DatasetVersion, error) {
	v, err := s.verRepo.FindByID(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("get version: %w", err)
	}
	if v == nil {
		return nil, ErrVersionNotFound
	}
	return v, nil
}

func (s *datasetService) ListVersions(ctx context.Context, datasetID uint64) ([]*entities.DatasetVersion, error) {
	vers, err := s.verRepo.FindByDatasetID(ctx, datasetID)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	return vers, nil
}
