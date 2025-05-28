package mocks

import (
	"context"

	"ppo/internal/entities"

	"github.com/stretchr/testify/mock"
)

type NotificationRepository struct {
	mock.Mock
}

func (m *NotificationRepository) Delete(ctx context.Context, id uint64) error {
	//TODO implement me
	panic("implement me")
}

func (m *NotificationRepository) Create(ctx context.Context, notif *entities.Notification) error {
	args := m.Called(ctx, notif)
	return args.Error(0)
}

func (m *NotificationRepository) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Notification, error) {
	args := m.Called(ctx, userID)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]*entities.Notification), args.Error(1)
}

func (m *NotificationRepository) FindByID(ctx context.Context, id uint64) (*entities.Notification, error) {
	args := m.Called(ctx, id)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*entities.Notification), args.Error(1)
}

func (m *NotificationRepository) Update(ctx context.Context, notif *entities.Notification) error {
	args := m.Called(ctx, notif)
	return args.Error(0)
}

type SubscriptionRepository struct {
	mock.Mock
}

func (m *SubscriptionRepository) Create(ctx context.Context, s *entities.Subscription) error {
	//TODO implement me
	panic("implement me")
}

func (m *SubscriptionRepository) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	panic("implement me")
}

func (m *SubscriptionRepository) GetByUser(ctx context.Context, userID uint64) ([]uint64, error) {
	panic("implement me")
}

func (m *SubscriptionRepository) Subscribe(ctx context.Context, userID, datasetID uint64) error {
	//TODO implement me
	panic("implement me")
}

func (m *SubscriptionRepository) IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m *SubscriptionRepository) GetSubscribers(ctx context.Context, datasetID uint64) ([]uint64, error) {
	args := m.Called(ctx, datasetID)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]uint64), args.Error(1)
}
