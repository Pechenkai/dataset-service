package services

import (
	"context"

	"ppo/internal/entities"
)

type SubscriptionService interface {
	Subscribe(ctx context.Context, userID, datasetID uint64) error
	Unsubscribe(ctx context.Context, userID, datasetID uint64) error
	ListSubscribers(ctx context.Context, datasetID uint64) ([]*entities.Subscription, error)
	ListSubscriptions(ctx context.Context, userID uint64) ([]*entities.Subscription, error)
	ListAllSubscriptions(ctx context.Context) ([]*entities.Subscription, error)
	IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error)
}
