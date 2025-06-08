package mocks

import (
	"context"
	"io"

	"ppo/internal/entities"

	"github.com/stretchr/testify/mock"
)

type DatasetRepository struct {
	mock.Mock
}

func (m *DatasetRepository) FindByCategoryID(ctx context.Context, categoryID uint64) ([]*entities.Dataset, error) {
	//TODO implement me
	panic("implement me")
}

func (m *DatasetRepository) FindPublic(ctx context.Context) ([]*entities.Dataset, error) {
	panic("implement me")
}

func (m *DatasetRepository) FindAll(ctx context.Context) ([]*entities.Dataset, error) {
	//TODO implement me
	panic("implement me")
}

func (m *DatasetRepository) Create(ctx context.Context, ds *entities.Dataset) error {
	args := m.Called(ctx, ds)
	return args.Error(0)
}

func (m *DatasetRepository) FindByID(ctx context.Context, id uint64) (*entities.Dataset, error) {
	args := m.Called(ctx, id)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*entities.Dataset), args.Error(1)
}

func (m *DatasetRepository) Update(ctx context.Context, d *entities.Dataset) error {
	args := m.Called(ctx, d)
	return args.Error(0)
}

func (m *DatasetRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *DatasetRepository) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Dataset, error) {
	args := m.Called(ctx, userID)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]*entities.Dataset), args.Error(1)
}

// ---- Versions ----

type DatasetVersionRepository struct {
	mock.Mock
}

func (m *DatasetVersionRepository) Delete(ctx context.Context, id uint64) error {
	//TODO implement me
	panic("implement me")
}

func (m *DatasetVersionRepository) Update(ctx context.Context, v *entities.DatasetVersion) error {
	//TODO implement me
	panic("implement me")
}

func (m *DatasetVersionRepository) FindByID(ctx context.Context, id uint64) (*entities.DatasetVersion, error) {
	//TODO implement me
	panic("implement me")
}

func (m *DatasetVersionRepository) Create(ctx context.Context, ver *entities.DatasetVersion) error {
	args := m.Called(ctx, ver)
	return args.Error(0)
}

func (m *DatasetVersionRepository) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.DatasetVersion, error) {
	args := m.Called(ctx, datasetID)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]*entities.DatasetVersion), args.Error(1)
}

type MetadataRepository struct {
	mock.Mock
}

func (m2 *MetadataRepository) Update(ctx context.Context, m *entities.Metadata) error {
	//TODO implement me
	panic("implement me")
}

func (m2 *MetadataRepository) Delete(ctx context.Context, id uint64) error {
	//TODO implement me
	panic("implement me")
}

func (m2 *MetadataRepository) FindByID(ctx context.Context, id uint64) (*entities.Metadata, error) {
	panic("implement me")
}

func (m2 *MetadataRepository) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Metadata, error) {
	panic("implement me")
}

func (m *MetadataRepository) Create(ctx context.Context, md *entities.Metadata) error {
	args := m.Called(ctx, md)
	return args.Error(0)
}

// ---- Storage ----

type Storage struct {
	mock.Mock
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) GetURL(ctx context.Context, key string) (url string, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) Upload(ctx context.Context, key string, r io.Reader, size int64) (string, error) {
	args := s.Called(ctx, key, r, size)
	return args.String(0), args.Error(1)
}
