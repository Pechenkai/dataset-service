package repositories

import (
	"context"
)

type SubscriptionRepository interface {
	Subscribe(ctx context.Context, userID, datasetID uint64) error
	IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error)
	GetSubscribers(ctx context.Context, datasetID uint64) ([]uint64, error)
}
