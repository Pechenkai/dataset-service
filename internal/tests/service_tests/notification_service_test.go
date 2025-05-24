package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"ppo/internal/entities"
	"ppo/internal/services"
	"ppo/internal/tests/mocks"
)

type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }

func TestNotifySubscribers_Success(t *testing.T) {
	notifRepo := new(mocks.NotificationRepository)
	subRepo := new(mocks.SubscriptionRepository)
	clk := fakeClock{now: time.Date(2025, 5, 24, 12, 0, 0, 0, time.UTC)}

	svc := services.NewNotificationService(notifRepo, subRepo, clk)

	subRepo.On("GetSubscribers", mock.Anything, uint64(7)).
		Return([]uint64{11, 22, 33}, nil)

	notifRepo.
		On("Create", mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
			return n.UserID != 0 && n.DatasetID == 7 && n.Message == "MSG"
		})).
		Return(nil).Times(3)

	count, err := svc.NotifySubscribers(context.Background(), 7, "MSG")
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	subRepo.AssertExpectations(t)
	notifRepo.AssertExpectations(t)
}

func TestNotifySubscribers_NoSubscribers(t *testing.T) {
	notifRepo := new(mocks.NotificationRepository)
	subRepo := new(mocks.SubscriptionRepository)
	clk := fakeClock{now: time.Now()}

	svc := services.NewNotificationService(notifRepo, subRepo, clk)

	subRepo.On("GetSubscribers", mock.Anything, uint64(5)).
		Return([]uint64{}, nil)

	cnt, err := svc.NotifySubscribers(context.Background(), 5, "hello")
	assert.Zero(t, cnt)
	assert.ErrorIs(t, err, services.ErrNoSubscribers)
}

func TestNotifySubscribers_FetchSubsError(t *testing.T) {
	notifRepo := new(mocks.NotificationRepository)
	subRepo := new(mocks.SubscriptionRepository)
	svc := services.NewNotificationService(notifRepo, subRepo, fakeClock{now: time.Now()})

	subRepo.On("GetSubscribers", mock.Anything, uint64(99)).
		Return(nil, errors.New("db fail"))

	_, err := svc.NotifySubscribers(context.Background(), 99, "MSG")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fetch subscribers")
}

func TestNotifySubscribers_InvalidMessage(t *testing.T) {
	notifRepo := new(mocks.NotificationRepository)
	subRepo := new(mocks.SubscriptionRepository)
	svc := services.NewNotificationService(notifRepo, subRepo, fakeClock{now: time.Now()})

	subRepo.On("GetSubscribers", mock.Anything, uint64(1)).
		Return([]uint64{1}, nil)

	// пустое сообщение => entities.NewNotification вернёт ErrEmptyMessage
	_, err := svc.NotifySubscribers(context.Background(), 1, "   ")
	assert.ErrorIs(t, err, entities.ErrEmptyMessage)
}

func TestNotifySubscribers_CreateError(t *testing.T) {
	notifRepo := new(mocks.NotificationRepository)
	subRepo := new(mocks.SubscriptionRepository)
	clk := fakeClock{now: time.Now()}

	svc := services.NewNotificationService(notifRepo, subRepo, clk)

	subRepo.On("GetSubscribers", mock.Anything, uint64(2)).
		Return([]uint64{2, 3}, nil)

	notifRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
	notifRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("io fail")).Once()

	cnt, err := svc.NotifySubscribers(context.Background(), 2, "MSG")
	assert.Equal(t, 1, cnt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create notification for user")
}

func TestGetNotificationsByUser_Success(t *testing.T) {
	notifRepo := new(mocks.NotificationRepository)
	subRepo := new(mocks.SubscriptionRepository)

	svc := services.NewNotificationService(notifRepo, subRepo, fakeClock{now: time.Now()})

	expected := []*entities.Notification{
		{ID: 1, UserID: 5, Message: "X"},
	}
	notifRepo.On("FindByUserID", mock.Anything, uint64(5)).
		Return(expected, nil)

	res, err := svc.GetNotificationsByUser(context.Background(), 5)
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestGetNotificationsByUser_Error(t *testing.T) {
	notifRepo := new(mocks.NotificationRepository)
	svc := services.NewNotificationService(notifRepo, new(mocks.SubscriptionRepository), fakeClock{now: time.Now()})

	notifRepo.On("FindByUserID", mock.Anything, uint64(6)).
		Return(nil, errors.New("db err"))

	_, err := svc.GetNotificationsByUser(context.Background(), 6)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list notifications")
}

func TestMarkAsRead_Success(t *testing.T) {
	notifRepo := new(mocks.NotificationRepository)
	svc := services.NewNotificationService(notifRepo, new(mocks.SubscriptionRepository), fakeClock{now: time.Now()})

	n := &entities.Notification{ID: 9, IsRead: false}
	notifRepo.On("FindByID", mock.Anything, uint64(9)).Return(n, nil)
	notifRepo.On("Update", mock.Anything, mock.MatchedBy(func(n *entities.Notification) bool {
		return n.ID == 9 && n.IsRead
	})).Return(nil)

	err := svc.MarkAsRead(context.Background(), 9)
	assert.NoError(t, err)
}

func TestMarkAsRead_NotFound(t *testing.T) {
	notifRepo := new(mocks.NotificationRepository)
	svc := services.NewNotificationService(notifRepo, new(mocks.SubscriptionRepository), fakeClock{now: time.Now()})

	notifRepo.On("FindByID", mock.Anything, uint64(10)).Return((*entities.Notification)(nil), nil)

	err := svc.MarkAsRead(context.Background(), 10)
	assert.ErrorIs(t, err, services.ErrNotificationNotFound)
}

func TestMarkAsRead_FetchError(t *testing.T) {
	notifRepo := new(mocks.NotificationRepository)
	svc := services.NewNotificationService(notifRepo, new(mocks.SubscriptionRepository), fakeClock{now: time.Now()})

	notifRepo.On("FindByID", mock.Anything, uint64(11)).Return(nil, errors.New("db fail"))

	err := svc.MarkAsRead(context.Background(), 11)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fetch notification")
}

func TestMarkAsRead_UpdateError(t *testing.T) {
	notifRepo := new(mocks.NotificationRepository)
	svc := services.NewNotificationService(notifRepo, new(mocks.SubscriptionRepository), fakeClock{now: time.Now()})

	n := &entities.Notification{ID: 12}
	notifRepo.On("FindByID", mock.Anything, uint64(12)).Return(n, nil)
	notifRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db update"))

	err := svc.MarkAsRead(context.Background(), 12)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mark as read")
}
