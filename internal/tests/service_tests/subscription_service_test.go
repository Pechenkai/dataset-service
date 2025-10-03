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

	"ppo/internal/repositories"
	"ppo/internal/services"
	"ppo/internal/tests/mocks"
	"ppo/internal/tests/testdata"
)

type SubscriptionServiceSuite struct {
	suite.Suite
}

type subscriptionMocks struct {
	repo *mocks.SubscriptionRepository
	svc  services.SubscriptionService
}

func newSubscriptionMocks(t provider.T) subscriptionMocks {
	t.Helper()
	repo := &mocks.SubscriptionRepository{}
	svc := services.NewSubscriptionService(repo, zap.NewNop())
	return subscriptionMocks{repo: repo, svc: svc}
}

func (m subscriptionMocks) AssertExpectations(t provider.T) {
	t.Helper()
	m.repo.AssertExpectations(t)
}

func (s *SubscriptionServiceSuite) TestSubscribe_Success(t provider.T) {
	fabric := testdata.NewFabric()
	user := fabric.RegularUser()
	dataset := fabric.Dataset(user, fabric.Category())
	m := newSubscriptionMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, user.ID, dataset.ID).Return(false, nil)
		m.repo.On("Create", mock.Anything, mock.Anything).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Subscribe(context.Background(), user.ID, dataset.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestSubscribe_InvalidIDs(t provider.T) {
	m := newSubscriptionMocks(t)

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Subscribe(context.Background(), 0, 10)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IDs")
	})
}

func (s *SubscriptionServiceSuite) TestSubscribe_AlreadySubscribed(t provider.T) {
	m := newSubscriptionMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, uint64(1), uint64(2)).Return(true, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Subscribe(context.Background(), 1, 2)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrAlreadySubscribed)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestSubscribe_CheckError(t provider.T) {
	m := newSubscriptionMocks(t)
	expected := errors.New("check fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, uint64(1), uint64(2)).Return(false, expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Subscribe(context.Background(), 1, 2)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestSubscribe_CreateError(t provider.T) {
	m := newSubscriptionMocks(t)
	expected := errors.New("insert fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, uint64(1), uint64(2)).Return(false, nil)
		m.repo.On("Create", mock.Anything, mock.Anything).Return(expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Subscribe(context.Background(), 1, 2)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestSubscribe_CreateAlreadyExists(t provider.T) {
	m := newSubscriptionMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, uint64(1), uint64(2)).Return(false, nil)
		m.repo.On("Create", mock.Anything, mock.Anything).Return(repositories.ErrAlreadySubscribed)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Subscribe(context.Background(), 1, 2)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrAlreadySubscribed)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestUnsubscribe_Success(t provider.T) {
	m := newSubscriptionMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, uint64(3), uint64(4)).Return(true, nil)
		m.repo.On("Unsubscribe", mock.Anything, uint64(3), uint64(4)).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Unsubscribe(context.Background(), 3, 4)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestUnsubscribe_InvalidIDs(t provider.T) {
	m := newSubscriptionMocks(t)

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Unsubscribe(context.Background(), 0, 1)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IDs")
	})
}

func (s *SubscriptionServiceSuite) TestUnsubscribe_NotSubscribed(t provider.T) {
	m := newSubscriptionMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, uint64(5), uint64(6)).Return(false, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Unsubscribe(context.Background(), 5, 6)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrNotSubscribed)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestUnsubscribe_CheckError(t provider.T) {
	m := newSubscriptionMocks(t)
	expected := errors.New("check fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, uint64(5), uint64(6)).Return(false, expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Unsubscribe(context.Background(), 5, 6)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestUnsubscribe_UnsubscribeError(t provider.T) {
	m := newSubscriptionMocks(t)
	expected := errors.New("delete fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, uint64(5), uint64(6)).Return(true, nil)
		m.repo.On("Unsubscribe", mock.Anything, uint64(5), uint64(6)).Return(expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Unsubscribe(context.Background(), 5, 6)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestUnsubscribe_RepoNotFound(t provider.T) {
	m := newSubscriptionMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, uint64(5), uint64(6)).Return(true, nil)
		m.repo.On("Unsubscribe", mock.Anything, uint64(5), uint64(6)).Return(repositories.ErrSubscriptionNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.Unsubscribe(context.Background(), 5, 6)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrNotSubscribed)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestListSubscribers_Success(t provider.T) {
	m := newSubscriptionMocks(t)
	expected := []uint64{1, 2, 3}

	var (
		result []uint64
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("GetSubscribers", mock.Anything, uint64(10)).Return(expected, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListSubscribers(context.Background(), 10)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, expected, result)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestListSubscribers_InvalidDataset(t provider.T) {
	m := newSubscriptionMocks(t)

	var (
		result []uint64
		err    error
	)

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListSubscribers(context.Background(), 0)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func (s *SubscriptionServiceSuite) TestListSubscribers_Error(t provider.T) {
	m := newSubscriptionMocks(t)
	expected := errors.New("list fail")

	var err error
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("GetSubscribers", mock.Anything, uint64(10)).Return(nil, expected)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.ListSubscribers(context.Background(), 10)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestListSubscriptions_Success(t provider.T) {
	m := newSubscriptionMocks(t)
	expected := []uint64{11, 22}

	var (
		result []uint64
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("GetByUser", mock.Anything, uint64(7)).Return(expected, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListSubscriptions(context.Background(), 7)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, expected, result)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestListSubscriptions_InvalidUser(t provider.T) {
	m := newSubscriptionMocks(t)

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.ListSubscriptions(context.Background(), 0)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid userID")
	})
}

func (s *SubscriptionServiceSuite) TestListSubscriptions_Error(t provider.T) {
	m := newSubscriptionMocks(t)
	expected := errors.New("list fail")

	var err error
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("GetByUser", mock.Anything, uint64(7)).Return(nil, expected)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.ListSubscriptions(context.Background(), 7)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestIsSubscribed_Success(t provider.T) {
	m := newSubscriptionMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, uint64(8), uint64(9)).Return(true, nil)
	})

	var (
		result bool
		err    error
	)
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.IsSubscribed(context.Background(), 8, 9)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.True(t, result)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestIsSubscribed_Error(t provider.T) {
	m := newSubscriptionMocks(t)
	expected := errors.New("check fail")

	var err error
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("IsSubscribed", mock.Anything, uint64(8), uint64(9)).Return(false, expected)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.IsSubscribed(context.Background(), 8, 9)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *SubscriptionServiceSuite) TestSubscribe_ClassicStyle(t provider.T) {
	repo := newFakeSubscriptionRepo()
	svc := services.NewSubscriptionService(repo, zap.NewNop())

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = svc.Subscribe(context.Background(), 1, 2)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		exists, _ := repo.IsSubscribed(context.Background(), 1, 2)
		assert.True(t, exists)
	})
}

func TestSubscriptionServiceSuite(t *testing.T) {
	suite.RunSuite(t, new(SubscriptionServiceSuite))
}
