package repositories

import (
	"context"

	"ppo/internal/entities"
)

type AccessRequestRepository interface {
	Create(ctx context.Context, ar *entities.AccessRequest) error
	Find(ctx context.Context, datasetID, userID uint64) (*entities.AccessRequest, error)
	ListPendingByOwner(ctx context.Context, ownerID uint64) ([]*entities.AccessRequest, error)
	UpdateStatus(ctx context.Context, id uint64, status string) error
	FindByRequestID(ctx context.Context, requestID uint64) (*entities.AccessRequest, error)
}
