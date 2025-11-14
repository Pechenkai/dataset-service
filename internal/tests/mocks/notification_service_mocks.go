package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type NotificationRepository struct {
	mock.Mock
}

func (m *NotificationRepository) Create(ctx context.Context, n *entities.Notification) error {
	return m.Called(ctx, n).Error(0)
}

func (m *NotificationRepository) Delete(ctx context.Context, id uint64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *NotificationRepository) Update(ctx context.Context, n *entities.Notification) error {
	return m.Called(ctx, n).Error(0)
}

func (m *NotificationRepository) FindByID(ctx context.Context, id uint64) (*entities.Notification, error) {
	args := m.Called(ctx, id)
	var notif *entities.Notification
	if value := args.Get(0); value != nil {
		notif, _ = value.(*entities.Notification)
	}
	return notif, args.Error(1)
}

func (m *NotificationRepository) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Notification, error) {
	args := m.Called(ctx, userID)
	var list []*entities.Notification
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Notification)
	}
	return list, args.Error(1)
}

var _ repositories.NotificationRepository = (*NotificationRepository)(nil)

type SubscriptionRepository struct {
	mock.Mock
}

func (m *SubscriptionRepository) Create(ctx context.Context, s *entities.Subscription) error {
	return m.Called(ctx, s).Error(0)
}

func (m *SubscriptionRepository) IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error) {
	args := m.Called(ctx, userID, datasetID)
	return args.Bool(0), args.Error(1)
}

func (m *SubscriptionRepository) GetSubscribers(ctx context.Context, datasetID uint64) ([]*entities.Subscription, error) {
	args := m.Called(ctx, datasetID)
	var list []*entities.Subscription
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Subscription)
	}
	return list, args.Error(1)
}

func (m *SubscriptionRepository) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	return m.Called(ctx, userID, datasetID).Error(0)
}

func (m *SubscriptionRepository) GetByUser(ctx context.Context, userID uint64) ([]*entities.Subscription, error) {
	args := m.Called(ctx, userID)
	var list []*entities.Subscription
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Subscription)
	}
	return list, args.Error(1)
}

func (m *SubscriptionRepository) GetAll(ctx context.Context) ([]*entities.Subscription, error) {
	args := m.Called(ctx)
	var list []*entities.Subscription
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Subscription)
	}
	return list, args.Error(1)
}

var _ repositories.SubscriptionRepository = (*SubscriptionRepository)(nil)
