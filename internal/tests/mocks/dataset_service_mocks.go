package mocks

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"

	"ppo/internal/entities"
	"ppo/internal/repositories"
	"ppo/internal/storage"
)

type DatasetRepository struct {
	mock.Mock
}

func (m *DatasetRepository) Create(ctx context.Context, d *entities.Dataset) error {
	return m.Called(ctx, d).Error(0)
}

func (m *DatasetRepository) Delete(ctx context.Context, id uint64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *DatasetRepository) Update(ctx context.Context, d *entities.Dataset) error {
	return m.Called(ctx, d).Error(0)
}

func (m *DatasetRepository) FindByID(ctx context.Context, id uint64) (*entities.Dataset, error) {
	args := m.Called(ctx, id)
	var ds *entities.Dataset
	if value := args.Get(0); value != nil {
		ds, _ = value.(*entities.Dataset)
	}
	return ds, args.Error(1)
}

func (m *DatasetRepository) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Dataset, error) {
	args := m.Called(ctx, userID)
	var list []*entities.Dataset
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Dataset)
	}
	return list, args.Error(1)
}

func (m *DatasetRepository) FindAll(ctx context.Context) ([]*entities.Dataset, error) {
	args := m.Called(ctx)
	var list []*entities.Dataset
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Dataset)
	}
	return list, args.Error(1)
}

func (m *DatasetRepository) FindPublic(ctx context.Context) ([]*entities.Dataset, error) {
	args := m.Called(ctx)
	var list []*entities.Dataset
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Dataset)
	}
	return list, args.Error(1)
}

func (m *DatasetRepository) FindByCategoryID(ctx context.Context, categoryID uint64) ([]*entities.Dataset, error) {
	args := m.Called(ctx, categoryID)
	var list []*entities.Dataset
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Dataset)
	}
	return list, args.Error(1)
}

type DatasetVersionRepository struct {
	mock.Mock
}

func (m *DatasetVersionRepository) Create(ctx context.Context, v *entities.DatasetVersion) error {
	return m.Called(ctx, v).Error(0)
}

func (m *DatasetVersionRepository) Delete(ctx context.Context, id uint64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *DatasetVersionRepository) Update(ctx context.Context, v *entities.DatasetVersion) error {
	return m.Called(ctx, v).Error(0)
}

func (m *DatasetVersionRepository) FindByID(ctx context.Context, id uint64) (*entities.DatasetVersion, error) {
	args := m.Called(ctx, id)
	var version *entities.DatasetVersion
	if value := args.Get(0); value != nil {
		version, _ = value.(*entities.DatasetVersion)
	}
	return version, args.Error(1)
}

func (m *DatasetVersionRepository) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.DatasetVersion, error) {
	args := m.Called(ctx, datasetID)
	var list []*entities.DatasetVersion
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.DatasetVersion)
	}
	return list, args.Error(1)
}

type MetadataRepository struct {
	mock.Mock
}

func (m *MetadataRepository) Create(ctx context.Context, md *entities.Metadata) error {
	return m.Called(ctx, md).Error(0)
}

func (m *MetadataRepository) Update(ctx context.Context, md *entities.Metadata) error {
	return m.Called(ctx, md).Error(0)
}

func (m *MetadataRepository) Delete(ctx context.Context, id uint64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MetadataRepository) FindByID(ctx context.Context, id uint64) (*entities.Metadata, error) {
	args := m.Called(ctx, id)
	var metadata *entities.Metadata
	if value := args.Get(0); value != nil {
		metadata, _ = value.(*entities.Metadata)
	}
	return metadata, args.Error(1)
}

func (m *MetadataRepository) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Metadata, error) {
	args := m.Called(ctx, datasetID)
	var list []*entities.Metadata
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Metadata)
	}
	return list, args.Error(1)
}

type Storage struct {
	mock.Mock
}

func (m *Storage) Upload(ctx context.Context, key string, r io.Reader, size int64) (string, error) {
	args := m.Called(ctx, key, r, size)
	return args.String(0), args.Error(1)
}

func (m *Storage) Delete(ctx context.Context, key string) error {
	return m.Called(ctx, key).Error(0)
}

func (m *Storage) GetURL(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

var _ repositories.DatasetRepository = (*DatasetRepository)(nil)
var _ repositories.DatasetVersionRepository = (*DatasetVersionRepository)(nil)
var _ repositories.MetadataRepository = (*MetadataRepository)(nil)
var _ storage.Storage = (*Storage)(nil)
