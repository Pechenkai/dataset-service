package services_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ppo/internal/entities"
	"ppo/internal/repositories"
	"ppo/internal/services"
	"ppo/internal/tests/mocks"
	"ppo/internal/tests/testdata"
)

type DatasetServiceSuite struct {
	suite.Suite
}

type serviceMocks struct {
	datasetRepo  *mocks.DatasetRepository
	versionRepo  *mocks.DatasetVersionRepository
	metadataRepo *mocks.MetadataRepository
	storage      *mocks.Storage
	svc          services.DatasetService
}

func newServiceMocks(t provider.T) serviceMocks {
	t.Helper()
	datasetRepo := &mocks.DatasetRepository{}
	versionRepo := &mocks.DatasetVersionRepository{}
	metadataRepo := &mocks.MetadataRepository{}
	storage := &mocks.Storage{}
	svc := services.NewDatasetService(datasetRepo, versionRepo, metadataRepo, storage, zap.NewNop())

	return serviceMocks{
		datasetRepo:  datasetRepo,
		versionRepo:  versionRepo,
		metadataRepo: metadataRepo,
		storage:      storage,
		svc:          svc,
	}
}

func (m serviceMocks) AssertExpectations(t provider.T) {
	t.Helper()
	m.datasetRepo.AssertExpectations(t)
	m.versionRepo.AssertExpectations(t)
	m.metadataRepo.AssertExpectations(t)
	m.storage.AssertExpectations(t)
}

func (s *DatasetServiceSuite) TestCreateDataset_Success(t provider.T) {
	fabric := testdata.NewFabric()
	actor := fabric.RegularUser()
	category := fabric.Category()
	cmd := fabric.CreateDatasetCommand(category, actor)
	reader, size := fabric.DatasetFile("initial payload")
	m := newServiceMocks(t)

	expectedDatasetID := uint64(9001)
	expectedVersionID := uint64(9002)
	expectedMetadataID := uint64(9003)
	expectedObjectKey := fmt.Sprintf("datasets/%d/v0.1/%s", expectedDatasetID, cmd.FileName)
	expectedURL := fmt.Sprintf("https://storage/%d", expectedDatasetID)

	var (
		id  uint64
		err error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Dataset")).Run(func(args mock.Arguments) {
			ds := args.Get(1).(*entities.Dataset)
			require.NotZero(t, ds.OwnerID)
			ds.ID = expectedDatasetID
		}).Return(nil)
		m.versionRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DatasetVersion")).Run(func(args mock.Arguments) {
			ver := args.Get(1).(*entities.DatasetVersion)
			ver.ID = expectedVersionID
		}).Return(nil)
		m.metadataRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Metadata")).Run(func(args mock.Arguments) {
			md := args.Get(1).(*entities.Metadata)
			md.ID = expectedMetadataID
		}).Return(nil)
		m.storage.On("Upload", mock.Anything, expectedObjectKey, mock.Anything, size).Return(expectedURL, nil)
	})

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		id, err = m.svc.CreateDataset(context.Background(), cmd, reader, size)
	})

	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, expectedDatasetID, id)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestCreateDataset_InvalidPayload(t provider.T) {
	fabric := testdata.NewFabric()
	actor := fabric.RegularUser()
	category := fabric.Category()
	cmd := fabric.InvalidCreateDatasetCommand(category, actor)
	m := newServiceMocks(t)

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.CreateDataset(context.Background(), cmd, nil, 0)
	})

	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid dataset")
	})
}

func (s *DatasetServiceSuite) TestCreateDataset_ClassicStyle(t provider.T) {
	fabric := testdata.NewFabricAt(time.Now().Add(-time.Hour))
	actor := fabric.RegularUser()
	category := fabric.Category()
	cmd := fabric.CreateDatasetCommand(category, actor)
	reader, size := fabric.DatasetFile("payload")

	repo := &inMemoryDatasetRepo{}
	verRepo := &inMemoryVersionRepo{}
	mdRepo := &inMemoryMetadataRepo{}
	storage := &inMemoryStorage{}
	svc := services.NewDatasetService(repo, verRepo, mdRepo, storage, zap.NewNop())

	var (
		id  uint64
		err error
	)

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		id, err = svc.CreateDataset(context.Background(), cmd, reader, size)
	})

	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.NotZero(t, id)
		assert.Len(t, repo.datasets, 1)
		assert.Len(t, verRepo.versions, 1)
	})
}

func (s *DatasetServiceSuite) TestAddDatasetVersion_Success(t provider.T) {
	fabric := testdata.NewFabric()
	fixture := fabric.DatasetFixtureWithMetadata()
	cmd := fabric.AddVersionCommand(fixture.Dataset, fabric.RegularUser())
	reader, size := fabric.DatasetVersionFile("new version data")
	m := newServiceMocks(t)

	existingVersion := *fixture.Version
	existingVersion.Number = "v0.1"
	existingVersion.Filepath = fmt.Sprintf("datasets/%d/v0.1/%s", fixture.Dataset.ID, cmd.FileName)

	expectedVersionID := uint64(9100)
	expectedKey := fmt.Sprintf("datasets/%d/v0.2/%s", fixture.Dataset.ID, cmd.FileName)
	expectedURL := "https://storage/new-version"

	var (
		id  uint64
		err error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, fixture.Dataset.ID).Return(fixture.Dataset, nil)
		m.versionRepo.On("FindByDatasetID", mock.Anything, fixture.Dataset.ID).Return([]*entities.DatasetVersion{&existingVersion}, nil)
		m.storage.On("Upload", mock.Anything, expectedKey, mock.Anything, size).Return(expectedURL, nil)
		m.versionRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DatasetVersion")).Run(func(args mock.Arguments) {
			ver := args.Get(1).(*entities.DatasetVersion)
			ver.ID = expectedVersionID
		}).Return(nil)
		m.metadataRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Metadata")).Return(nil)
	})

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		id, err = m.svc.AddDatasetVersion(context.Background(), cmd, reader, size)
	})

	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, expectedVersionID, id)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestAddDatasetVersion_DatasetMissing(t provider.T) {
	fabric := testdata.NewFabric()
	dataset := fabric.DatasetFixture().Dataset
	cmd := fabric.AddVersionCommand(dataset, fabric.RegularUser())
	m := newServiceMocks(t)

	var err error
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return((*entities.Dataset)(nil), repositories.ErrDatasetNotFound)
	})

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.AddDatasetVersion(context.Background(), cmd, nil, 0)
	})

	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrDatasetNotFound)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestAddDatasetVersion_UploadFails(t provider.T) {
	fabric := testdata.NewFabric()
	fixture := fabric.DatasetFixture()
	cmd := fabric.AddVersionCommand(fixture.Dataset, fabric.RegularUser())
	reader, size := fabric.DatasetVersionFile("payload")
	m := newServiceMocks(t)

	uploadErr := errors.New("storage offline")
	expectedKey := fmt.Sprintf("datasets/%d/v0.1/%s", fixture.Dataset.ID, cmd.FileName)

	var err error
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, fixture.Dataset.ID).Return(fixture.Dataset, nil)
		m.versionRepo.On("FindByDatasetID", mock.Anything, fixture.Dataset.ID).Return(nil, nil)
		m.storage.On("Upload", mock.Anything, expectedKey, mock.Anything, size).Return("", uploadErr)
	})

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.AddDatasetVersion(context.Background(), cmd, reader, size)
	})

	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorContains(t, err, "upload file")
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestGetDataset_Success(t provider.T) {
	fabric := testdata.NewFabric()
	fixture := fabric.DatasetFixture()
	m := newServiceMocks(t)

	var (
		result *entities.Dataset
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, fixture.Dataset.ID).Return(fixture.Dataset, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetDataset(context.Background(), fixture.Dataset.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, fixture.Dataset.ID, result.ID)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestGetDataset_NotFound(t provider.T) {
	m := newServiceMocks(t)
	datasetID := uint64(42)

	var (
		result *entities.Dataset
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, datasetID).Return((*entities.Dataset)(nil), repositories.ErrDatasetNotFound)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetDataset(context.Background(), datasetID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, services.ErrDatasetNotFound)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestListDatasets_ByOwnerClassic(t provider.T) {
	fabric := testdata.NewFabric()
	owner := fabric.RegularUser()
	category := fabric.Category()
	repo := &inMemoryDatasetRepo{}
	dataset := fabric.Dataset(owner, category)
	repo.datasets = append(repo.datasets, dataset)
	svc := services.NewDatasetService(repo, &inMemoryVersionRepo{}, &inMemoryMetadataRepo{}, &inMemoryStorage{}, zap.NewNop())

	var (
		result []*entities.Dataset
		err    error
	)

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		onlyPublic := false
		ownerID := owner.ID
		result, err = svc.ListDatasets(context.Background(), onlyPublic, &ownerID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, dataset.ID, result[0].ID)
	})
}

func (s *DatasetServiceSuite) TestListDatasets_RepoError(t provider.T) {
	m := newServiceMocks(t)

	var (
		result []*entities.Dataset
		err    error
	)
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindAll", mock.Anything).Return(nil, repositories.ErrDatasetList)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListDatasets(context.Background(), false, nil)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, repositories.ErrDatasetList)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestGetVersion_Success(t provider.T) {
	fabric := testdata.NewFabric()
	fixture := fabric.DatasetFixture()
	m := newServiceMocks(t)

	var (
		result *entities.DatasetVersion
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.versionRepo.On("FindByID", mock.Anything, fixture.Version.ID).Return(fixture.Version, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetVersion(context.Background(), fixture.Version.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, fixture.Version.ID, result.ID)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestGetVersion_NotFound(t provider.T) {
	m := newServiceMocks(t)
	versionID := uint64(77)

	var (
		result *entities.DatasetVersion
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.versionRepo.On("FindByID", mock.Anything, versionID).Return((*entities.DatasetVersion)(nil), repositories.ErrVersionNotFound)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetVersion(context.Background(), versionID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, services.ErrVersionNotFound)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestListVersions_Success(t provider.T) {
	fabric := testdata.NewFabric()
	fixture := fabric.DatasetFixture()
	versions := []*entities.DatasetVersion{fixture.Version}
	m := newServiceMocks(t)

	var (
		result []*entities.DatasetVersion
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.versionRepo.On("FindByDatasetID", mock.Anything, fixture.Dataset.ID).Return(versions, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListVersions(context.Background(), fixture.Dataset.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, versions, result)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestListVersions_RepoError(t provider.T) {
	m := newServiceMocks(t)
	datasetID := uint64(55)
	expectedErr := errors.New("db down")

	var (
		result []*entities.DatasetVersion
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.versionRepo.On("FindByDatasetID", mock.Anything, datasetID).Return(nil, expectedErr)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListVersions(context.Background(), datasetID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "list versions")
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestListByCategory_Success(t provider.T) {
	fabric := testdata.NewFabric()
	fixture := fabric.DatasetFixture()
	datasets := []*entities.Dataset{fixture.Dataset}
	m := newServiceMocks(t)

	var (
		result []*entities.Dataset
		err    error
	)
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByCategoryID", mock.Anything, fixture.Category.ID).Return(datasets, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListByCategory(context.Background(), fixture.Category.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, datasets, result)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestListByCategory_RepoError(t provider.T) {
	m := newServiceMocks(t)
	categoryID := uint64(33)
	expectedErr := errors.New("query failed")

	var (
		result []*entities.Dataset
		err    error
	)
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByCategoryID", mock.Anything, categoryID).Return(nil, expectedErr)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListByCategory(context.Background(), categoryID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "list datasets by category")
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestDeleteDataset_Success(t provider.T) {
	m := newServiceMocks(t)
	datasetID := uint64(88)

	var err error
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("Delete", mock.Anything, datasetID).Return(nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.DeleteDataset(context.Background(), datasetID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestDeleteDataset_NotFound(t provider.T) {
	m := newServiceMocks(t)
	datasetID := uint64(99)

	var err error
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("Delete", mock.Anything, datasetID).Return(repositories.ErrDatasetNotFound)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.DeleteDataset(context.Background(), datasetID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrDatasetNotFound)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestGetDownloadURL_Success(t provider.T) {
	fabric := testdata.NewFabric()
	fixture := fabric.DatasetFixtureWithMetadata()
	expectedURL := "https://presigned"
	m := newServiceMocks(t)

	var (
		url string
		err error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.versionRepo.On("FindByDatasetID", mock.Anything, fixture.Dataset.ID).Return([]*entities.DatasetVersion{fixture.Version}, nil)
		m.storage.On("GetURL", mock.Anything, fixture.Version.Filepath).Return(expectedURL, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		url, err = m.svc.GetDownloadURL(context.Background(), fixture.Dataset.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, expectedURL, url)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestGetDownloadURL_NoVersions(t provider.T) {
	m := newServiceMocks(t)
	datasetID := uint64(123)

	var (
		url string
		err error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.versionRepo.On("FindByDatasetID", mock.Anything, datasetID).Return([]*entities.DatasetVersion{}, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		url, err = m.svc.GetDownloadURL(context.Background(), datasetID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Empty(t, url)
		assert.ErrorIs(t, err, services.ErrVersionNotFound)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestUpdateDataset_Success(t provider.T) {
	fabric := testdata.NewFabric()
	fixture := fabric.DatasetFixture()
	category := fabric.Category()
	cmd := fabric.UpdateDatasetCommand(fixture.Dataset, category)
	m := newServiceMocks(t)

	var err error

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, cmd.ID).Return(fixture.Dataset, nil)
		m.datasetRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Dataset")).Return(nil)
	})

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateDataset(context.Background(), cmd)
	})

	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, cmd.Name, fixture.Dataset.Name)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestUpdateDataset_NotFound(t provider.T) {
	fabric := testdata.NewFabric()
	fixture := fabric.DatasetFixture()
	cmd := fabric.UpdateDatasetCommand(fixture.Dataset, fabric.Category())
	m := newServiceMocks(t)

	var err error
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, cmd.ID).Return((*entities.Dataset)(nil), repositories.ErrDatasetNotFound)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateDataset(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrDatasetNotFound)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestGetVersionDownloadURL_Success(t provider.T) {
	fabric := testdata.NewFabric()
	fixture := fabric.DatasetFixture()
	expectedURL := "https://version-url"
	m := newServiceMocks(t)

	var (
		url string
		err error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.versionRepo.On("FindByID", mock.Anything, fixture.Version.ID).Return(fixture.Version, nil)
		m.storage.On("GetURL", mock.Anything, fixture.Version.Filepath).Return(expectedURL, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		url, err = m.svc.GetVersionDownloadURL(context.Background(), fixture.Version.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, expectedURL, url)
		m.AssertExpectations(t)
	})
}

func (s *DatasetServiceSuite) TestGetVersionDownloadURL_StorageError(t provider.T) {
	fabric := testdata.NewFabric()
	fixture := fabric.DatasetFixture()
	expectedErr := errors.New("s3 down")
	m := newServiceMocks(t)

	var (
		url string
		err error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.versionRepo.On("FindByID", mock.Anything, fixture.Version.ID).Return(fixture.Version, nil)
		m.storage.On("GetURL", mock.Anything, fixture.Version.Filepath).Return("", expectedErr)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		url, err = m.svc.GetVersionDownloadURL(context.Background(), fixture.Version.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Empty(t, url)
		assert.ErrorContains(t, err, "presign version")
		m.AssertExpectations(t)
	})
}

func TestDatasetServiceSuite(t *testing.T) {
	suite.RunSuite(t, new(DatasetServiceSuite))
}

// --- Classic-style in-memory fakes ------------------------------------------------

type inMemoryDatasetRepo struct {
	datasets []*entities.Dataset
}

func (r *inMemoryDatasetRepo) Create(ctx context.Context, d *entities.Dataset) error {
	r.datasets = append(r.datasets, d)
	if d.ID == 0 {
		d.ID = uint64(len(r.datasets))
	}
	return nil
}

func (r *inMemoryDatasetRepo) Delete(ctx context.Context, id uint64) error {
	return nil
}

func (r *inMemoryDatasetRepo) Update(ctx context.Context, d *entities.Dataset) error {
	return nil
}

func (r *inMemoryDatasetRepo) FindByID(ctx context.Context, id uint64) (*entities.Dataset, error) {
	for _, ds := range r.datasets {
		if ds.ID == id {
			return ds, nil
		}
	}
	return nil, repositories.ErrDatasetNotFound
}

func (r *inMemoryDatasetRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Dataset, error) {
	var result []*entities.Dataset
	for _, ds := range r.datasets {
		if ds.OwnerID == userID {
			result = append(result, ds)
		}
	}
	return result, nil
}

func (r *inMemoryDatasetRepo) FindAll(ctx context.Context) ([]*entities.Dataset, error) {
	return r.datasets, nil
}

func (r *inMemoryDatasetRepo) FindPublic(ctx context.Context) ([]*entities.Dataset, error) {
	var result []*entities.Dataset
	for _, ds := range r.datasets {
		if ds.IsPublic {
			result = append(result, ds)
		}
	}
	return result, nil
}

func (r *inMemoryDatasetRepo) FindByCategoryID(ctx context.Context, categoryID uint64) ([]*entities.Dataset, error) {
	var result []*entities.Dataset
	for _, ds := range r.datasets {
		if ds.CategoryID == categoryID {
			result = append(result, ds)
		}
	}
	return result, nil
}

// ---

type inMemoryVersionRepo struct {
	versions []*entities.DatasetVersion
}

func (r *inMemoryVersionRepo) Create(ctx context.Context, v *entities.DatasetVersion) error {
	r.versions = append(r.versions, v)
	if v.ID == 0 {
		v.ID = uint64(len(r.versions))
	}
	return nil
}

func (r *inMemoryVersionRepo) Delete(ctx context.Context, id uint64) error { return nil }
func (r *inMemoryVersionRepo) Update(ctx context.Context, v *entities.DatasetVersion) error {
	return nil
}

func (r *inMemoryVersionRepo) FindByID(ctx context.Context, id uint64) (*entities.DatasetVersion, error) {
	for _, v := range r.versions {
		if v.ID == id {
			return v, nil
		}
	}
	return nil, repositories.ErrVersionNotFound
}

func (r *inMemoryVersionRepo) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.DatasetVersion, error) {
	var result []*entities.DatasetVersion
	for _, v := range r.versions {
		if v.DatasetID == datasetID {
			result = append(result, v)
		}
	}
	return result, nil
}

// ---

type inMemoryMetadataRepo struct{}

func (r *inMemoryMetadataRepo) Create(ctx context.Context, m *entities.Metadata) error { return nil }
func (r *inMemoryMetadataRepo) Update(ctx context.Context, m *entities.Metadata) error { return nil }
func (r *inMemoryMetadataRepo) Delete(ctx context.Context, id uint64) error            { return nil }
func (r *inMemoryMetadataRepo) FindByID(ctx context.Context, id uint64) (*entities.Metadata, error) {
	return nil, repositories.ErrMetadataNotFound
}
func (r *inMemoryMetadataRepo) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Metadata, error) {
	return nil, nil
}

// ---

type inMemoryStorage struct{}

func (s *inMemoryStorage) Upload(ctx context.Context, key string, r io.Reader, size int64) (string, error) {
	return fmt.Sprintf("memory://%s", key), nil
}
func (s *inMemoryStorage) Delete(ctx context.Context, key string) error { return nil }
func (s *inMemoryStorage) GetURL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("memory://%s", key), nil
}
