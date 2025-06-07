package services

import (
	"context"
)

type SubscriptionService interface {
	Subscribe(ctx context.Context, userID, datasetID uint64) error
	Unsubscribe(ctx context.Context, userID, datasetID uint64) error
	ListSubscribers(ctx context.Context, datasetID uint64) ([]uint64, error)
	ListSubscriptions(ctx context.Context, userID uint64) ([]uint64, error)
	IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error)
}
