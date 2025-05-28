package repositories

import (
	"context"
	"ppo/internal/entities"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, s *entities.Subscription) error
	IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error)
	GetSubscribers(ctx context.Context, datasetID uint64) ([]uint64, error)
	Unsubscribe(ctx context.Context, userID, datasetID uint64) error
	GetByUser(ctx context.Context, userID uint64) ([]uint64, error)
}
