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

type NotificationServiceSuite struct {
	suite.Suite
}

type notificationMocks struct {
	notifRepo *mocks.NotificationRepository
	subRepo   *mocks.SubscriptionRepository
	svc       services.NotificationService
}

func newNotificationMocks(t provider.T) notificationMocks {
	t.Helper()
	notifRepo := &mocks.NotificationRepository{}
	subRepo := &mocks.SubscriptionRepository{}
	svc := services.NewNotificationService(notifRepo, subRepo, zap.NewNop())
	return notificationMocks{notifRepo: notifRepo, subRepo: subRepo, svc: svc}
}

func (m notificationMocks) AssertExpectations(t provider.T) {
	t.Helper()
	m.notifRepo.AssertExpectations(t)
	m.subRepo.AssertExpectations(t)
}

func (s *NotificationServiceSuite) TestNotifySubscribers_Success(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.NotifySubscribersCommand(42, "dataset updated")
	m := newNotificationMocks(t)

	subscribers := []uint64{101, 202}
	notifID := uint64(500)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.subRepo.On("GetSubscribers", mock.Anything, cmd.DatasetID).Return(subscribers, nil)
		m.notifRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Notification")).Run(func(args mock.Arguments) {
			n := args.Get(1).(*entities.Notification)
			n.ID = notifID
			notifID++
		}).Return(nil)
	})

	var (
		count int
		err   error
	)

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		count, err = m.svc.NotifySubscribers(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, len(subscribers), count)
		m.AssertExpectations(t)
	})
}

func (s *NotificationServiceSuite) TestNotifySubscribers_NoSubscribers(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.NotifySubscribersCommand(42, "update")
	m := newNotificationMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.subRepo.On("GetSubscribers", mock.Anything, cmd.DatasetID).Return([]uint64{}, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.NotifySubscribers(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrNoSubscribers)
		m.AssertExpectations(t)
	})
}

func (s *NotificationServiceSuite) TestNotifySubscribers_ListError(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.NotifySubscribersCommand(42, "update")
	m := newNotificationMocks(t)
	expectedErr := errors.New("db down")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.subRepo.On("GetSubscribers", mock.Anything, cmd.DatasetID).Return(nil, expectedErr)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.NotifySubscribers(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *NotificationServiceSuite) TestNotifySubscribers_InvalidNotification(t provider.T) {
	cmd := testdata.NewNotifySubscribersCmdBuilder().WithDataset(42).WithoutMessage().Build()
	m := newNotificationMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.subRepo.On("GetSubscribers", mock.Anything, cmd.DatasetID).Return([]uint64{1}, nil)
	})

	var err error
	var count int
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		count, err = m.svc.NotifySubscribers(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Equal(t, 0, count)
		m.AssertExpectations(t)
	})
}

func (s *NotificationServiceSuite) TestNotifySubscribers_CreateError(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.NotifySubscribersCommand(42, "update")
	m := newNotificationMocks(t)
	expectedErr := errors.New("insert fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.subRepo.On("GetSubscribers", mock.Anything, cmd.DatasetID).Return([]uint64{1}, nil)
		m.notifRepo.On("Create", mock.Anything, mock.Anything).Return(expectedErr)
	})

	var err error
	var count int
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		count, err = m.svc.NotifySubscribers(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *NotificationServiceSuite) TestNotifySubscribers_ClassicStyle(t provider.T) {
	repo := &inMemoryNotificationRepo{}
	subRepo := newFakeSubscriptionRepo()
	subRepo.subs[42] = []uint64{1, 2}
	svc := services.NewNotificationService(repo, subRepo, zap.NewNop())
	cmd := services.NotifySubscribersCmd{DatasetID: 42, Message: "update"}

	var (
		count int
		err   error
	)
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		count, err = svc.NotifySubscribers(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, repo.notifications, 2)
	})
}

func (s *NotificationServiceSuite) TestGetNotificationsByUser_Success(t provider.T) {
	fabric := testdata.NewFabric()
	notif := fabric.Notification(fabric.RegularUser(), fabric.Dataset(fabric.RegularUser(), fabric.Category()))
	m := newNotificationMocks(t)

	var (
		result []*entities.Notification
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.notifRepo.On("FindByUserID", mock.Anything, notif.UserID).Return([]*entities.Notification{notif}, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetNotificationsByUser(context.Background(), notif.UserID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, notif.ID, result[0].ID)
		m.AssertExpectations(t)
	})
}

func (s *NotificationServiceSuite) TestGetNotificationsByUser_RepoError(t provider.T) {
	m := newNotificationMocks(t)
	expectedErr := errors.New("select fail")

	var err error
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.notifRepo.On("FindByUserID", mock.Anything, uint64(1)).Return(nil, expectedErr)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.GetNotificationsByUser(context.Background(), 1)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *NotificationServiceSuite) TestMarkAsRead_Success(t provider.T) {
	fabric := testdata.NewFabric()
	notif := fabric.Notification(fabric.RegularUser(), fabric.Dataset(fabric.RegularUser(), fabric.Category()))
	notif.ID = 777
	m := newNotificationMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.notifRepo.On("FindByID", mock.Anything, notif.ID).Return(notif, nil)
		m.notifRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Notification")).Run(func(args mock.Arguments) {
			updated := args.Get(1).(*entities.Notification)
			assert.True(t, updated.IsRead)
		}).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.MarkAsRead(context.Background(), notif.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *NotificationServiceSuite) TestMarkAsRead_NotFoundOnFetch(t provider.T) {
	m := newNotificationMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.notifRepo.On("FindByID", mock.Anything, uint64(1)).Return((*entities.Notification)(nil), repositories.ErrNotificationNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.MarkAsRead(context.Background(), 1)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrNotificationNotFound)
		m.AssertExpectations(t)
	})
}

func (s *NotificationServiceSuite) TestMarkAsRead_UpdateError(t provider.T) {
	fabric := testdata.NewFabric()
	notif := fabric.Notification(fabric.RegularUser(), fabric.Dataset(fabric.RegularUser(), fabric.Category()))
	notif.ID = 888
	m := newNotificationMocks(t)
	expectedErr := errors.New("update fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.notifRepo.On("FindByID", mock.Anything, notif.ID).Return(notif, nil)
		m.notifRepo.On("Update", mock.Anything, mock.Anything).Return(expectedErr)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.MarkAsRead(context.Background(), notif.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *NotificationServiceSuite) TestNotifyUser_Success(t provider.T) {
	notifRepo := &mocks.NotificationRepository{}
	svc := services.NewNotificationService(notifRepo, &mocks.SubscriptionRepository{}, zap.NewNop())

	notifRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Notification")).Return(nil)

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = svc.NotifyUser(context.Background(), 1, 2, "direct notify")
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		notifRepo.AssertExpectations(t)
	})
}

func (s *NotificationServiceSuite) TestNotifyUser_InvalidPayload(t provider.T) {
	svc := services.NewNotificationService(&mocks.NotificationRepository{}, &mocks.SubscriptionRepository{}, zap.NewNop())

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = svc.NotifyUser(context.Background(), 1, 2, "")
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, entities.ErrEmptyMessage)
	})
}

func (s *NotificationServiceSuite) TestNotifyUser_CreateError(t provider.T) {
	notifRepo := &mocks.NotificationRepository{}
	svc := services.NewNotificationService(notifRepo, &mocks.SubscriptionRepository{}, zap.NewNop())
	expectedErr := errors.New("insert fail")

	notifRepo.On("Create", mock.Anything, mock.Anything).Return(expectedErr)

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = svc.NotifyUser(context.Background(), 1, 2, "notify")
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		notifRepo.AssertExpectations(t)
	})
}

func TestNotificationServiceSuite(t *testing.T) {
	suite.RunSuite(t, new(NotificationServiceSuite))
}

type inMemoryNotificationRepo struct {
	notifications []*entities.Notification
}

func (r *inMemoryNotificationRepo) Create(ctx context.Context, n *entities.Notification) error {
	r.notifications = append(r.notifications, n)
	if n.ID == 0 {
		n.ID = uint64(len(r.notifications))
	}
	return nil
}

func (r *inMemoryNotificationRepo) Delete(ctx context.Context, id uint64) error { return nil }

func (r *inMemoryNotificationRepo) Update(ctx context.Context, n *entities.Notification) error {
	for i, notif := range r.notifications {
		if notif.ID == n.ID {
			r.notifications[i] = n
			return nil
		}
	}
	return repositories.ErrNotificationNotFound
}

func (r *inMemoryNotificationRepo) FindByID(ctx context.Context, id uint64) (*entities.Notification, error) {
	for _, notif := range r.notifications {
		if notif.ID == id {
			return notif, nil
		}
	}
	return nil, repositories.ErrNotificationNotFound
}

func (r *inMemoryNotificationRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Notification, error) {
	var result []*entities.Notification
	for _, notif := range r.notifications {
		if notif.UserID == userID {
			result = append(result, notif)
		}
	}
	return result, nil
}
