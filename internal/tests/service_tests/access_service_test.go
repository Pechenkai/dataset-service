package services_test

import (
	"context"
	"errors"
	"testing"

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

type AccessServiceSuite struct {
	suite.Suite
}

type accessMocks struct {
	datasetRepo *mocks.DatasetRepository
	arRepo      *mocks.AccessRequestRepository
	svc         services.AccessService
}

func newAccessMocks(t provider.T) accessMocks {
	t.Helper()
	datasetRepo := &mocks.DatasetRepository{}
	arRepo := &mocks.AccessRequestRepository{}
	svc := services.NewAccessService(datasetRepo, arRepo, zap.NewNop())
	return accessMocks{datasetRepo: datasetRepo, arRepo: arRepo, svc: svc}
}

func (m accessMocks) AssertExpectations(t provider.T) {
	t.Helper()
	m.datasetRepo.AssertExpectations(t)
	m.arRepo.AssertExpectations(t)
}

func (s *AccessServiceSuite) TestRequest_Success(t provider.T) {
	m := newAccessMocks(t)
	fabric := testdata.NewFabric()
	dataset := fabric.Dataset(fabric.RegularUser(), fabric.Category())
	dataset.IsPublic = false
	cmd := fabric.RequestAccessCommand(dataset.ID, 999)
	ar := testdata.NewAccessRequestBuilder().WithDatasetID(cmd.DatasetID).WithUserID(cmd.UserID).Build()

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
		m.arRepo.On("Find", mock.Anything, cmd.DatasetID, cmd.UserID).Return((*entities.AccessRequest)(nil), repositories.ErrRequestNotFound)
		m.arRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.AccessRequest")).Run(func(args mock.Arguments) {
			created := args.Get(1).(*entities.AccessRequest)
			created.ID = ar.ID
		}).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Request(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestRequest_DatasetNotFound(t provider.T) {
	m := newAccessMocks(t)
	cmd := services.RequestAccessCmd{DatasetID: 42, UserID: 10}

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, cmd.DatasetID).Return((*entities.Dataset)(nil), repositories.ErrDatasetNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Request(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrRequestNotFound)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestRequest_PublicDataset(t provider.T) {
	m := newAccessMocks(t)
	dataset := &entities.Dataset{ID: 1, OwnerID: 2, IsPublic: true}
	cmd := services.RequestAccessCmd{DatasetID: dataset.ID, UserID: 3}

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Request(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrBadRequest)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestRequest_OwnerCannotRequest(t provider.T) {
	m := newAccessMocks(t)
	dataset := &entities.Dataset{ID: 2, OwnerID: 5, IsPublic: false}
	cmd := services.RequestAccessCmd{DatasetID: dataset.ID, UserID: dataset.OwnerID}

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Request(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrBadRequest)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestRequest_DeniedResetToPending(t provider.T) {
	m := newAccessMocks(t)
	dataset := &entities.Dataset{ID: 1, OwnerID: 99, IsPublic: false}
	cmd := services.RequestAccessCmd{DatasetID: dataset.ID, UserID: 3}
	existing := testdata.NewAccessRequestBuilder().WithID(55).WithDatasetID(dataset.ID).WithUserID(cmd.UserID).WithStatus(entities.AccessStatusDenied).Build()

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
		m.arRepo.On("Find", mock.Anything, cmd.DatasetID, cmd.UserID).Return(existing, nil)
		m.arRepo.On("UpdateStatus", mock.Anything, existing.ID, string(entities.AccessStatusPending)).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Request(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestRequest_AlreadyExists(t provider.T) {
	m := newAccessMocks(t)
	dataset := &entities.Dataset{ID: 1, OwnerID: 99, IsPublic: false}
	cmd := services.RequestAccessCmd{DatasetID: dataset.ID, UserID: 3}
	existing := testdata.NewAccessRequestBuilder().WithDatasetID(dataset.ID).WithUserID(cmd.UserID).Build()

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
		m.arRepo.On("Find", mock.Anything, cmd.DatasetID, cmd.UserID).Return(existing, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Request(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrRequestAlreadyExists)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestRequest_CreateError(t provider.T) {
	m := newAccessMocks(t)
	dataset := &entities.Dataset{ID: 1, OwnerID: 99, IsPublic: false}
	cmd := services.RequestAccessCmd{DatasetID: dataset.ID, UserID: 3}
	expected := errors.New("insert fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
		m.arRepo.On("Find", mock.Anything, cmd.DatasetID, cmd.UserID).Return((*entities.AccessRequest)(nil), repositories.ErrRequestNotFound)
		m.arRepo.On("Create", mock.Anything, mock.Anything).Return(expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Request(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestRequest_InvalidNewRequest(t provider.T) {
	m := newAccessMocks(t)
	cmd := services.RequestAccessCmd{DatasetID: 0, UserID: 0}
	dataset := &entities.Dataset{ID: 0, OwnerID: 1, IsPublic: false}

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.datasetRepo.On("FindByID", mock.Anything, cmd.DatasetID).Return(dataset, nil)
		m.arRepo.On("Find", mock.Anything, cmd.DatasetID, cmd.UserID).Return((*entities.AccessRequest)(nil), repositories.ErrRequestNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Request(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestApprove_Success(t provider.T) {
	m := newAccessMocks(t)
	dataset := &entities.Dataset{ID: 10, OwnerID: 5}
	req := testdata.NewAccessRequestBuilder().WithID(40).WithDatasetID(dataset.ID).WithUserID(9).Build()

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("FindByRequestID", mock.Anything, req.ID).Return(req, nil)
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
		m.arRepo.On("UpdateStatus", mock.Anything, req.ID, string(entities.AccessStatusApproved)).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Approve(context.Background(), req.ID, dataset.OwnerID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestApprove_RequestNotFound(t provider.T) {
	m := newAccessMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("FindByRequestID", mock.Anything, uint64(1)).Return((*entities.AccessRequest)(nil), repositories.ErrRequestNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Approve(context.Background(), 1, 5)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrRequestNotFound)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestApprove_DatasetNotFound(t provider.T) {
	m := newAccessMocks(t)
	req := testdata.NewAccessRequestBuilder().WithID(5).WithDatasetID(9).Build()

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("FindByRequestID", mock.Anything, req.ID).Return(req, nil)
		m.datasetRepo.On("FindByID", mock.Anything, req.DatasetID).Return((*entities.Dataset)(nil), repositories.ErrDatasetNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Approve(context.Background(), req.ID, 1)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrRequestNotFound)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestApprove_Forbidden(t provider.T) {
	m := newAccessMocks(t)
	dataset := &entities.Dataset{ID: 10, OwnerID: 5}
	req := testdata.NewAccessRequestBuilder().WithID(40).WithDatasetID(dataset.ID).Build()

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("FindByRequestID", mock.Anything, req.ID).Return(req, nil)
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Approve(context.Background(), req.ID, 999)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrRequestForbidden)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestApprove_UpdateError(t provider.T) {
	m := newAccessMocks(t)
	dataset := &entities.Dataset{ID: 10, OwnerID: 5}
	req := testdata.NewAccessRequestBuilder().WithID(40).WithDatasetID(dataset.ID).Build()
	expected := errors.New("update fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("FindByRequestID", mock.Anything, req.ID).Return(req, nil)
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
		m.arRepo.On("UpdateStatus", mock.Anything, req.ID, string(entities.AccessStatusApproved)).Return(expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Approve(context.Background(), req.ID, dataset.OwnerID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestDeny_Success(t provider.T) {
	m := newAccessMocks(t)
	dataset := &entities.Dataset{ID: 7, OwnerID: 4}
	req := testdata.NewAccessRequestBuilder().WithID(12).WithDatasetID(dataset.ID).Build()

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("FindByRequestID", mock.Anything, req.ID).Return(req, nil)
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
		m.arRepo.On("UpdateStatus", mock.Anything, req.ID, string(entities.AccessStatusDenied)).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Deny(context.Background(), req.ID, dataset.OwnerID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestDeny_Forbidden(t provider.T) {
	m := newAccessMocks(t)
	dataset := &entities.Dataset{ID: 7, OwnerID: 4}
	req := testdata.NewAccessRequestBuilder().WithID(12).WithDatasetID(dataset.ID).Build()

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("FindByRequestID", mock.Anything, req.ID).Return(req, nil)
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Deny(context.Background(), req.ID, 999)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrRequestForbidden)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestDeny_UpdateError(t provider.T) {
	m := newAccessMocks(t)
	dataset := &entities.Dataset{ID: 7, OwnerID: 4}
	req := testdata.NewAccessRequestBuilder().WithID(12).WithDatasetID(dataset.ID).Build()
	expected := errors.New("update fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("FindByRequestID", mock.Anything, req.ID).Return(req, nil)
		m.datasetRepo.On("FindByID", mock.Anything, dataset.ID).Return(dataset, nil)
		m.arRepo.On("UpdateStatus", mock.Anything, req.ID, string(entities.AccessStatusDenied)).Return(expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Deny(context.Background(), req.ID, dataset.OwnerID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestListPending_Success(t provider.T) {
	m := newAccessMocks(t)
	reqs := []*entities.AccessRequest{testdata.NewAccessRequestBuilder().WithID(1).Build()}

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("ListPendingByOwner", mock.Anything, uint64(5)).Return(reqs, nil)
	})

	var (
		result []*entities.AccessRequest
		err    error
	)

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListPending(context.Background(), 5)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, reqs, result)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestListPending_Error(t provider.T) {
	m := newAccessMocks(t)
	expected := errors.New("list fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("ListPendingByOwner", mock.Anything, uint64(5)).Return(nil, expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.ListPending(context.Background(), 5)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestFindByRequestID_Success(t provider.T) {
	m := newAccessMocks(t)
	req := testdata.NewAccessRequestBuilder().WithID(99).Build()

	var (
		result *entities.AccessRequest
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("FindByRequestID", mock.Anything, req.ID).Return(req, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.FindByRequestID(context.Background(), req.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, req.ID, result.ID)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestFindByRequestID_Error(t provider.T) {
	m := newAccessMocks(t)
	expected := errors.New("fetch fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("FindByRequestID", mock.Anything, uint64(55)).Return((*entities.AccessRequest)(nil), expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.FindByRequestID(context.Background(), 55)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestFind_Success(t provider.T) {
	m := newAccessMocks(t)
	req := testdata.NewAccessRequestBuilder().WithID(7).Build()

	var (
		result *entities.AccessRequest
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("Find", mock.Anything, req.DatasetID, req.UserID).Return(req, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.Find(context.Background(), req.DatasetID, req.UserID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, req.ID, result.ID)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestFind_NotFound(t provider.T) {
	m := newAccessMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("Find", mock.Anything, uint64(1), uint64(2)).Return((*entities.AccessRequest)(nil), repositories.ErrRequestNotFound)
	})

	var (
		result *entities.AccessRequest
		err    error
	)
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.Find(context.Background(), 1, 2)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, repositories.ErrRequestNotFound)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestFind_Error(t provider.T) {
	m := newAccessMocks(t)
	expected := errors.New("find fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.arRepo.On("Find", mock.Anything, uint64(1), uint64(2)).Return((*entities.AccessRequest)(nil), expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.Find(context.Background(), 1, 2)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *AccessServiceSuite) TestRequest_ClassicStyle(t provider.T) {
	datasetRepo := newFakeDatasetRepo()
	arRepo := newFakeAccessRequestRepo()
	dataset := testdata.NewDatasetBuilder().WithID(1).WithOwner(5).Public(false).Build()
	datasetRepo.datasets[dataset.ID] = dataset
	svc := services.NewAccessService(datasetRepo, arRepo, zap.NewNop())
	cmd := services.RequestAccessCmd{DatasetID: dataset.ID, UserID: 7}

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = svc.Request(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		stored, _ := arRepo.Find(context.Background(), cmd.DatasetID, cmd.UserID)
		require.NotNil(t, stored)
		assert.Equal(t, entities.AccessStatusPending, stored.Status)
	})
}

func TestAccessServiceSuite(t *testing.T) {
	suite.RunSuite(t, new(AccessServiceSuite))
}

// --- Classic-style fakes -----------------------------------------------------------

type fakeDatasetRepo struct {
	datasets map[uint64]*entities.Dataset
}

func newFakeDatasetRepo() *fakeDatasetRepo {
	return &fakeDatasetRepo{datasets: make(map[uint64]*entities.Dataset)}
}

func (r *fakeDatasetRepo) Create(ctx context.Context, d *entities.Dataset) error { return nil }

func (r *fakeDatasetRepo) Delete(ctx context.Context, id uint64) error { return nil }

func (r *fakeDatasetRepo) Update(ctx context.Context, d *entities.Dataset) error { return nil }

func (r *fakeDatasetRepo) FindByID(ctx context.Context, id uint64) (*entities.Dataset, error) {
	d, ok := r.datasets[id]
	if !ok {
		return nil, repositories.ErrDatasetNotFound
	}
	return d, nil
}

func (r *fakeDatasetRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Dataset, error) {
	return nil, nil
}
func (r *fakeDatasetRepo) FindAll(ctx context.Context) ([]*entities.Dataset, error) { return nil, nil }
func (r *fakeDatasetRepo) FindPublic(ctx context.Context) ([]*entities.Dataset, error) {
	return nil, nil
}
func (r *fakeDatasetRepo) FindByCategoryID(ctx context.Context, categoryID uint64) ([]*entities.Dataset, error) {
	return nil, nil
}

var _ repositories.DatasetRepository = (*fakeDatasetRepo)(nil)

type fakeAccessRequestRepo struct {
	nextID   uint64
	requests map[uint64]*entities.AccessRequest
}

func newFakeAccessRequestRepo() *fakeAccessRequestRepo {
	return &fakeAccessRequestRepo{nextID: 1, requests: make(map[uint64]*entities.AccessRequest)}
}

func (r *fakeAccessRequestRepo) Create(ctx context.Context, ar *entities.AccessRequest) error {
	ar.ID = r.nextID
	r.nextID++
	r.requests[ar.ID] = ar
	return nil
}

func (r *fakeAccessRequestRepo) Find(ctx context.Context, datasetID, userID uint64) (*entities.AccessRequest, error) {
	for _, ar := range r.requests {
		if ar.DatasetID == datasetID && ar.UserID == userID {
			return ar, nil
		}
	}
	return nil, repositories.ErrRequestNotFound
}

func (r *fakeAccessRequestRepo) ListPendingByOwner(ctx context.Context, ownerID uint64) ([]*entities.AccessRequest, error) {
	var result []*entities.AccessRequest
	for _, ar := range r.requests {
		result = append(result, ar)
	}
	return result, nil
}

func (r *fakeAccessRequestRepo) UpdateStatus(ctx context.Context, id uint64, status string) error {
	ar, ok := r.requests[id]
	if !ok {
		return repositories.ErrRequestNotFound
	}
	ar.Status = entities.AccessStatus(status)
	return nil
}

func (r *fakeAccessRequestRepo) FindByRequestID(ctx context.Context, requestID uint64) (*entities.AccessRequest, error) {
	if ar, ok := r.requests[requestID]; ok {
		return ar, nil
	}
	return nil, repositories.ErrRequestNotFound
}

var _ repositories.AccessRequestRepository = (*fakeAccessRequestRepo)(nil)
