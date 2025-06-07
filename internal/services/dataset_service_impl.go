package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
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
	logger  *zap.Logger
}

func NewDatasetService(
	dsRepo repositories.DatasetRepository,
	verRepo repositories.DatasetVersionRepository,
	mdRepo repositories.MetadataRepository,
	storage storage.Storage,
	logger *zap.Logger,
) DatasetService {
	return &datasetService{dsRepo: dsRepo, verRepo: verRepo, mdRepo: mdRepo, storage: storage, logger: logger}
}

func (s *datasetService) CreateDataset(
	ctx context.Context,
	cmd CreateDatasetCmd,
	r io.Reader,
	size int64,
) (uint64, error) {
	ds, err := entities.NewDataset(
		cmd.Name,
		cmd.Description,
		cmd.ActorID,
		cmd.CategoryID,
		cmd.IsPublic,
		time.Now(),
	)
	if err != nil {
		s.logger.Error("failed to validate dataset payload",
			zap.Error(err),
			zap.String("name", cmd.Name),
			zap.Uint64("category_id", cmd.CategoryID),
			zap.Uint64("actor_id", cmd.ActorID),
		)
		return 0, fmt.Errorf("invalid dataset: %w", err)
	}
	s.logger.Debug("dataset entity initialized",
		zap.String("name", ds.Name),
		zap.Uint64("category_id", ds.CategoryID),
		zap.Uint64("actor_id", ds.OwnerID),
	)

	if err := s.dsRepo.Create(ctx, ds); err != nil {
		s.logger.Error("failed to insert dataset into repository",
			zap.Error(err),
			zap.String("name", ds.Name),
			zap.Uint64("category_id", ds.CategoryID),
			zap.Uint64("actor_id", ds.OwnerID),
		)
		return 0, fmt.Errorf("create dataset: %w", err)
	}
	s.logger.Info("dataset created",
		zap.Uint64("dataset_id", ds.ID),
		zap.String("name", ds.Name),
	)

	key := fmt.Sprintf("datasets/%d/%s", ds.ID, cmd.FileName)
	url, err := s.storage.Upload(ctx, key, r, size)
	if err != nil {
		s.logger.Error("failed to upload dataset file to storage",
			zap.Error(err),
			zap.Uint64("dataset_id", ds.ID),
			zap.String("storage_key", key),
		)
		if delErr := s.dsRepo.Delete(ctx, ds.ID); delErr != nil {
			s.logger.Warn("failed to delete dataset after upload error",
				zap.Error(delErr),
				zap.Uint64("dataset_id", ds.ID),
			)
		}
		return 0, fmt.Errorf("upload file: %w", err)
	}
	s.logger.Info("dataset file uploaded",
		zap.Uint64("dataset_id", ds.ID),
		zap.String("storage_key", key),
		zap.String("url", url),
	)

	ver, err := entities.NewDatasetVersion("v0.1", url, "", ds.ID, time.Now())
	if err != nil {
		s.logger.Error("failed to construct initial dataset version entity",
			zap.Error(err),
			zap.Uint64("dataset_id", ds.ID),
		)
		_ = s.storage.Delete(ctx, key)
		_ = s.dsRepo.Delete(ctx, ds.ID)
		return 0, fmt.Errorf("invalid version data: %w", err)
	}
	s.logger.Debug("initial dataset version entity constructed",
		zap.Uint64("dataset_id", ds.ID),
		zap.String("version_number", ver.Number),
	)

	if err := s.verRepo.Create(ctx, ver); err != nil {
		s.logger.Error("failed to insert dataset version into repository",
			zap.Error(err),
			zap.Uint64("dataset_id", ds.ID),
			zap.String("version_number", ver.Number),
		)
		_ = s.storage.Delete(ctx, key)
		_ = s.dsRepo.Delete(ctx, ds.ID)
		return 0, fmt.Errorf("create version: %w", err)
	}
	s.logger.Info("dataset version created",
		zap.Uint64("dataset_id", ds.ID),
		zap.Uint64("version_id", ver.ID),
		zap.String("version_number", ver.Number),
	)

	if cmd.MetaFormat != "" || cmd.MetaSize != 0 {
		md, err := entities.NewMetadata(cmd.MetaFormat, cmd.MetaTags, cmd.MetaSize, ver.ID)
		if err != nil {
			s.logger.Error("failed to construct metadata entity",
				zap.Error(err),
				zap.Uint64("version_id", ver.ID),
			)
			_ = s.storage.Delete(ctx, key)
			_ = s.verRepo.Delete(ctx, ver.ID)
			_ = s.dsRepo.Delete(ctx, ds.ID)
			return 0, fmt.Errorf("%w: %s", ErrInvalidMetadata, err)
		}
		s.logger.Debug("metadata entity constructed", zap.Uint64("version_id", ver.ID))

		if err := s.mdRepo.Create(ctx, md); err != nil {
			s.logger.Error("failed to insert metadata into repository",
				zap.Error(err),
				zap.Uint64("version_id", ver.ID),
			)
			_ = s.storage.Delete(ctx, key)
			_ = s.verRepo.Delete(ctx, ver.ID)
			_ = s.dsRepo.Delete(ctx, ds.ID)
			return 0, fmt.Errorf("create metadata: %w", err)
		}
		s.logger.Info("metadata created", zap.Uint64("metadata_id", md.ID), zap.Uint64("version_id", ver.ID))
	}

	s.logger.Info("CreateDataset completed successfully", zap.Uint64("dataset_id", ds.ID))
	return ds.ID, nil
}

func (s *datasetService) AddDatasetVersion(
	ctx context.Context,
	cmd AddVersionCmd,
	r io.Reader,
	size int64,
) (uint64, error) {
	ds, err := s.dsRepo.FindByID(ctx, cmd.DatasetID)
	if err != nil {
		s.logger.Error("error fetching dataset before adding version",
			zap.Error(err),
			zap.Uint64("dataset_id", cmd.DatasetID),
		)
		if err == repositories.ErrDatasetNotFound {
			return 0, ErrDatasetNotFound
		}
		return 0, fmt.Errorf("fetch dataset: %w", err)
	}
	if ds == nil {
		s.logger.Warn("dataset not found when trying to add version", zap.Uint64("dataset_id", cmd.DatasetID))
		return 0, ErrDatasetNotFound
	}
	s.logger.Debug("dataset found for new version", zap.Uint64("dataset_id", ds.ID))

	vers, err := s.verRepo.FindByDatasetID(ctx, cmd.DatasetID)
	if err != nil {
		s.logger.Error("error fetching existing versions",
			zap.Error(err),
			zap.Uint64("dataset_id", cmd.DatasetID),
		)
		return 0, fmt.Errorf("fetch versions: %w", err)
	}
	next := nextVersionNumber(vers)
	s.logger.Debug("calculated next version number", zap.String("next_version", next), zap.Uint64("dataset_id", ds.ID))

	key := fmt.Sprintf("datasets/%d/%s", ds.ID, cmd.FileName)
	url, err := s.storage.Upload(ctx, key, r, size)
	if err != nil {
		s.logger.Error("failed to upload new version file",
			zap.Error(err),
			zap.Uint64("dataset_id", ds.ID),
			zap.String("storage_key", key),
		)
		return 0, fmt.Errorf("upload file: %w", err)
	}
	s.logger.Info("new version file uploaded", zap.Uint64("dataset_id", ds.ID), zap.String("storage_key", key), zap.String("url", url))

	ver, err := entities.NewDatasetVersion(next, url, cmd.ChangeLog, cmd.DatasetID, time.Now())
	if err != nil {
		s.logger.Error("failed to construct version entity",
			zap.Error(err),
			zap.Uint64("dataset_id", ds.ID),
			zap.String("version_number", next),
		)
		_ = s.storage.Delete(ctx, key)
		return 0, fmt.Errorf("invalid version data: %w", err)
	}
	s.logger.Debug("dataset version entity constructed",
		zap.Uint64("dataset_id", ds.ID),
		zap.String("version_number", ver.Number),
	)

	if err := s.verRepo.Create(ctx, ver); err != nil {
		s.logger.Error("failed to insert new version into repository",
			zap.Error(err),
			zap.Uint64("dataset_id", ds.ID),
			zap.String("version_number", ver.Number),
		)
		_ = s.storage.Delete(ctx, key)
		if err == repositories.ErrVersionNotFound {
			return 0, ErrVersionNotFound
		}
		return 0, fmt.Errorf("create version: %w", err)
	}
	s.logger.Info("new dataset version created",
		zap.Uint64("dataset_id", ds.ID),
		zap.Uint64("version_id", ver.ID),
		zap.String("version_number", ver.Number),
	)

	if cmd.MetaFormat != "" || cmd.MetaSize != 0 {
		md, err := entities.NewMetadata(cmd.MetaFormat, cmd.MetaTags, cmd.MetaSize, ver.ID)
		if err != nil {
			s.logger.Error("failed to construct metadata entity for new version",
				zap.Error(err),
				zap.Uint64("version_id", ver.ID),
			)
			_ = s.storage.Delete(ctx, key)
			_ = s.verRepo.Delete(ctx, ver.ID)
			return 0, fmt.Errorf("%w: %s", ErrInvalidMetadata, err)
		}
		s.logger.Debug("metadata entity for new version constructed", zap.Uint64("version_id", ver.ID))

		if err := s.mdRepo.Create(ctx, md); err != nil {
			s.logger.Error("failed to insert metadata for new version",
				zap.Error(err),
				zap.Uint64("version_id", ver.ID),
			)
			_ = s.storage.Delete(ctx, key)
			_ = s.verRepo.Delete(ctx, ver.ID)
			return 0, fmt.Errorf("create metadata: %w", err)
		}
		s.logger.Info("metadata for new version created", zap.Uint64("metadata_id", md.ID), zap.Uint64("version_id", ver.ID))
	}

	s.logger.Info("AddDatasetVersion completed successfully",
		zap.Uint64("dataset_id", ds.ID),
		zap.Uint64("version_id", ver.ID),
	)
	return ver.ID, nil
}

func (s *datasetService) GetDataset(ctx context.Context, id uint64) (*entities.Dataset, error) {
	s.logger.Debug("GetDataset called", zap.Uint64("dataset_id", id))

	ds, err := s.dsRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("error fetching dataset",
			zap.Error(err),
			zap.Uint64("dataset_id", id),
		)
		if err == repositories.ErrDatasetNotFound {
			return nil, ErrDatasetNotFound
		}
		return nil, fmt.Errorf("get dataset: %w", err)
	}
	if ds == nil {
		s.logger.Warn("dataset not found", zap.Uint64("dataset_id", id))
		return nil, ErrDatasetNotFound
	}
	s.logger.Info("dataset fetched successfully",
		zap.Uint64("dataset_id", ds.ID),
		zap.String("name", ds.Name),
	)
	return ds, nil
}

func (s *datasetService) ListDatasets(
	ctx context.Context,
	onlyPublic bool,
	ownerID *uint64,
) ([]*entities.Dataset, error) {
	s.logger.Debug("ListDatasets called", zap.Bool("only_public", onlyPublic))
	if ownerID != nil {
		s.logger.Debug("filter by owner", zap.Uint64("owner_id", *ownerID))
		return s.dsRepo.FindByUserID(ctx, *ownerID)
	}
	if onlyPublic {
		return s.dsRepo.FindPublic(ctx)
	}
	return s.dsRepo.FindAll(ctx)
}

func (s *datasetService) GetVersion(ctx context.Context, versionID uint64) (*entities.DatasetVersion, error) {
	s.logger.Debug("GetVersion called", zap.Uint64("version_id", versionID))

	v, err := s.verRepo.FindByID(ctx, versionID)
	if err != nil {
		s.logger.Error("error fetching version",
			zap.Error(err),
			zap.Uint64("version_id", versionID),
		)
		if err == repositories.ErrVersionNotFound {
			return nil, ErrVersionNotFound
		}
		return nil, fmt.Errorf("get version: %w", err)
	}
	if v == nil {
		s.logger.Warn("version not found", zap.Uint64("version_id", versionID))
		return nil, ErrVersionNotFound
	}
	s.logger.Info("version fetched successfully",
		zap.Uint64("version_id", v.ID),
		zap.String("version_number", v.Number),
	)
	return v, nil
}

func (s *datasetService) ListVersions(ctx context.Context, datasetID uint64) ([]*entities.DatasetVersion, error) {
	s.logger.Debug("ListVersions called", zap.Uint64("dataset_id", datasetID))
	vers, err := s.verRepo.FindByDatasetID(ctx, datasetID)
	if err != nil {
		s.logger.Error("error fetching versions list",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, fmt.Errorf("list versions: %w", err)
	}
	s.logger.Info("versions list fetched", zap.Int("count", len(vers)), zap.Uint64("dataset_id", datasetID))
	return vers, nil
}

func nextVersionNumber(existing []*entities.DatasetVersion) string {
	if len(existing) == 0 {
		return "v1.0"
	}
	last := existing[0].Number
	return last + ".1"
}

func (s *datasetService) ListByCategory(ctx context.Context, categoryID uint64) ([]*entities.Dataset, error) {
	s.logger.Debug("ListByCategory called", zap.Uint64("category_id", categoryID))

	dsets, err := s.dsRepo.FindByCategoryID(ctx, categoryID)
	if err != nil {
		s.logger.Error("ListByCategory failed", zap.Error(err), zap.Uint64("category_id", categoryID))
		return nil, fmt.Errorf("list datasets by category: %w", err)
	}
	s.logger.Info("ListByCategory completed", zap.Uint64("category_id", categoryID), zap.Int("count", len(dsets)))
	return dsets, nil
}
