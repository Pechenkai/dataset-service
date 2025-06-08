package services

import (
	"context"
	"ppo/internal/entities"
)

type RequestAccessCmd struct {
	DatasetID uint64
	UserID    uint64
}

type AccessService interface {
	Request(ctx context.Context, cmd RequestAccessCmd) error
	Approve(ctx context.Context, requestID uint64, ownerID uint64) error
	Deny(ctx context.Context, requestID uint64, ownerID uint64) error
	ListPending(ctx context.Context, ownerID uint64) ([]*entities.AccessRequest, error)
	FindByRequestID(ctx context.Context, requestID uint64) (*entities.AccessRequest, error)
	Find(ctx context.Context, datasetID, userID uint64) (*entities.AccessRequest, error)
}
