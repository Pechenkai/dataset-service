package services_test

import (
	"bytes"
	"context"
	"errors"
	"go.uber.org/zap"
	"io"
	"ppo/internal/entities"
	"ppo/internal/services"
	"ppo/internal/tests/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func makeReader(content string) io.Reader {
	return bytes.NewReader([]byte(content))
}

func TestDatasetService_CreateDataset_Success_WithMetadata(t *testing.T) {
	dsRepo := new(mocks.DatasetRepository)
	verRepo := new(mocks.DatasetVersionRepository)
	mdRepo := new(mocks.MetadataRepository)
	storage := new(mocks.Storage)

	logger := zap.NewNop()

	svc := services.NewDatasetService(dsRepo, verRepo, mdRepo, storage, logger)

	dsRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Dataset")).Run(func(args mock.Arguments) {
		ds := args.Get(1).(*entities.Dataset)
		ds.ID = 10
	}).Return(nil)

	storage.On("Upload", mock.Anything, "datasets/10/myfile.txt", mock.Anything, int64(9)).
		Return("http://url/to/myfile.txt", nil)

	verRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DatasetVersion")).Run(func(args mock.Arguments) {
		ver := args.Get(1).(*entities.DatasetVersion)
		ver.ID = 20
	}).Return(nil)

	mdRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Metadata")).Return(nil)

	cmd := services.CreateDatasetCmd{
		ActorID:     1,
		Name:        "Test",
		Description: "Desc",
		CategoryID:  2,
		FileName:    "myfile.txt",
		IsPublic:    true,
		MetaFormat:  "csv",
		MetaTags:    "tag1,tag2",
		MetaSize:    123,
	}

	id, err := svc.CreateDataset(context.Background(), cmd, makeReader("content!!"), 9)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), id)

	dsRepo.AssertExpectations(t)
	storage.AssertExpectations(t)
	verRepo.AssertExpectations(t)
	mdRepo.AssertExpectations(t)
}

func TestDatasetService_CreateDataset_ValidationError(t *testing.T) {
	logger := zap.NewNop()
	svc := services.NewDatasetService(nil, nil, nil, nil, logger)
	_, err := svc.CreateDataset(context.Background(), services.CreateDatasetCmd{
		ActorID:    1,
		Name:       "   ",
		CategoryID: 1,
		FileName:   "f.txt",
	}, nil, 0)
	assert.ErrorIs(t, err, entities.ErrEmptyDatasetName)
}

func TestDatasetService_CreateDataset_DatasetRepoError(t *testing.T) {
	dsRepo := new(mocks.DatasetRepository)
	logger := zap.NewNop()
	svc := services.NewDatasetService(dsRepo, nil, nil, nil, logger)

	dsRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	_, err := svc.CreateDataset(context.Background(), services.CreateDatasetCmd{
		ActorID:    1,
		Name:       "Name",
		CategoryID: 1,
		FileName:   "f.txt",
	}, makeReader(""), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create dataset")
}

func TestDatasetService_CreateDataset_UploadError(t *testing.T) {
	dsRepo := new(mocks.DatasetRepository)
	storage := new(mocks.Storage)

	dsRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		ds := args.Get(1).(*entities.Dataset)
		ds.ID = 5
	}).Return(nil)

	storage.On("Upload", mock.Anything, "datasets/5/f.txt", mock.Anything, int64(0)).Return("", errors.New("net err"))
	dsRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
	logger := zap.NewNop()
	svc := services.NewDatasetService(dsRepo, nil, nil, storage, logger)

	_, err := svc.CreateDataset(context.Background(), services.CreateDatasetCmd{
		ActorID:    1,
		Name:       "Name",
		CategoryID: 1,
		FileName:   "f.txt",
	}, makeReader(""), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload file")
}

func TestDatasetService_AddDatasetVersion_Success_NoMetadata(t *testing.T) {
	dsRepo := new(mocks.DatasetRepository)
	verRepo := new(mocks.DatasetVersionRepository)
	mdRepo := new(mocks.MetadataRepository)
	storage := new(mocks.Storage)
	logger := zap.NewNop()
	svc := services.NewDatasetService(dsRepo, verRepo, mdRepo, storage, logger)

	dsRepo.On("FindByID", mock.Anything, uint64(99)).Return(&entities.Dataset{ID: 99, Name: "A"}, nil)
	verRepo.On("FindByDatasetID", mock.Anything, uint64(99)).Return([]*entities.DatasetVersion{}, nil)
	storage.On("Upload", mock.Anything, "datasets/99/vers.txt", mock.Anything, int64(0)).
		Return("url", nil)
	verRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DatasetVersion")).Run(func(args mock.Arguments) {
		v := args.Get(1).(*entities.DatasetVersion)
		v.ID = 77
	}).Return(nil)

	id, err := svc.AddDatasetVersion(context.Background(), services.AddVersionCmd{
		ActorID:   1,
		DatasetID: 99,
		FileName:  "vers.txt",
		ChangeLog: "chg",
	}, makeReader(""), 0)
	assert.NoError(t, err)
	assert.Equal(t, uint64(77), id)
}

func TestDatasetService_AddDatasetVersion_DatasetNotFound(t *testing.T) {
	dsRepo := new(mocks.DatasetRepository)
	logger := zap.NewNop()
	svc := services.NewDatasetService(dsRepo, nil, nil, nil, logger)

	dsRepo.On("FindByID", mock.Anything, uint64(123)).Return((*entities.Dataset)(nil), nil)

	_, err := svc.AddDatasetVersion(context.Background(), services.AddVersionCmd{DatasetID: 123}, nil, 0)
	assert.ErrorIs(t, err, services.ErrDatasetNotFound)
}

func TestDatasetService_GetDataset_Success(t *testing.T) {
	dsRepo := new(mocks.DatasetRepository)
	logger := zap.NewNop()
	svc := services.NewDatasetService(dsRepo, nil, nil, nil, logger)

	expected := &entities.Dataset{ID: 5, Name: "OK"}
	dsRepo.On("FindByID", mock.Anything, uint64(5)).Return(expected, nil)

	ds, err := svc.GetDataset(context.Background(), 5)
	assert.NoError(t, err)
	assert.Equal(t, expected, ds)
}

func TestDatasetService_GetDataset_NotFound(t *testing.T) {
	dsRepo := new(mocks.DatasetRepository)
	logger := zap.NewNop()
	svc := services.NewDatasetService(dsRepo, nil, nil, nil, logger)

	dsRepo.On("FindByID", mock.Anything, uint64(6)).Return((*entities.Dataset)(nil), nil)

	ds, err := svc.GetDataset(context.Background(), 6)
	assert.ErrorIs(t, err, services.ErrDatasetNotFound)
	assert.Nil(t, ds)
}
