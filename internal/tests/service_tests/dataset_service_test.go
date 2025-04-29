package services

import (
	"errors"
	"ppo/internal/entities"
	"ppo/internal/services"
	"testing"
)

type mockDatasetRepo struct {
	createErr   error
	updateErr   error
	byIDRes     *entities.Dataset
	deleteErr   error
	findAllRes  []*entities.Dataset
	byUserIDRes []*entities.Dataset
}

func (m *mockDatasetRepo) Delete(id int64) error {
	return m.deleteErr
}

func (m *mockDatasetRepo) FindByUserID(userID int64) ([]*entities.Dataset, error) {
	return m.byUserIDRes, nil
}

func (m *mockDatasetRepo) FindAll() ([]*entities.Dataset, error) {
	return m.findAllRes, nil
}

func (m *mockDatasetRepo) Create(d *entities.Dataset) error { return m.createErr }

func (m *mockDatasetRepo) Update(d *entities.Dataset) error { return m.updateErr }

func (m *mockDatasetRepo) FindByID(id int64) (*entities.Dataset, error) {
	return &entities.Dataset{ID: id}, nil
}

type mockVersionRepo struct {
	createErr error
	versions  []*entities.DatasetVersion
	byIDRes   *entities.DatasetVersion
	deleteErr error
	updateErr error
}

func (m *mockVersionRepo) Delete(id int64) error {
	return m.deleteErr
}

func (m *mockVersionRepo) Update(datasetVersion *entities.DatasetVersion) error {
	return m.updateErr
}

func (m *mockVersionRepo) FindByID(id int64) (*entities.DatasetVersion, error) {
	return m.byIDRes, nil
}

func (m *mockVersionRepo) Create(v *entities.DatasetVersion) error { return m.createErr }

func (m *mockVersionRepo) FindByDatasetID(id int64) ([]*entities.DatasetVersion, error) {
	return m.versions, nil
}

type mockMetadataRepo struct {
	createErr  error
	updateErr  error
	byIDRes    *entities.Metadata
	deleteErr  error
	findAllRes []*entities.Metadata
}

func (m *mockMetadataRepo) Update(dataset *entities.Metadata) error {
	return m.updateErr
}

func (m *mockMetadataRepo) Delete(id int64) error {
	return m.deleteErr
}

func (m *mockMetadataRepo) FindByID(id int64) (*entities.Metadata, error) {
	return m.byIDRes, nil
}

func (m *mockMetadataRepo) FindByDatasetID(datasetID int64) ([]*entities.Metadata, error) {
	return m.findAllRes, nil
}

func (m *mockMetadataRepo) Create(met *entities.Metadata) error { return m.createErr }

func TestDatasetService_CreateAndUpdate(t *testing.T) {
	// CreateDataset: nil Dataset
	svc := services.NewDatasetService(&mockDatasetRepo{}, &mockVersionRepo{}, &mockMetadataRepo{})
	err := svc.CreateDataset(nil, &entities.Metadata{})
	if !errors.Is(err, services.ErrNilDataset) {
		t.Error("ожидается ошибка при передаче nil Dataset в CreateDataset")
	}

	// CreateDataset: ошибка создания версии
	dRepo := &mockDatasetRepo{}
	vRepoErr := &mockVersionRepo{createErr: errors.New("сбой создания версии")}
	mRepo := &mockMetadataRepo{}
	svc = services.NewDatasetService(dRepo, vRepoErr, mRepo)
	err = svc.CreateDataset(&entities.Dataset{}, &entities.Metadata{})
	if err == nil {
		t.Error("ожидается ошибка при создании версии")
	}

	// CreateDataset: ошибка создания метаданных
	vRepoOK := &mockVersionRepo{}
	mRepoErr := &mockMetadataRepo{createErr: errors.New("сбой создания метаданных")}
	svc = services.NewDatasetService(dRepo, vRepoOK, mRepoErr)
	err = svc.CreateDataset(&entities.Dataset{}, &entities.Metadata{})
	if err == nil {
		t.Error("ожидается ошибка при создании метаданных")
	}

	// UpdateDataset: nil Dataset
	svc = services.NewDatasetService(&mockDatasetRepo{}, &mockVersionRepo{}, &mockMetadataRepo{})
	err = svc.UpdateDataset(nil, "1.1", "файл", &entities.Metadata{})
	if !errors.Is(err, services.ErrNilDataset) {
		t.Error("ожидается ошибка при передаче nil Dataset в UpdateDataset")
	}
}
