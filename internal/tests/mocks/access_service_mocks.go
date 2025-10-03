package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type AccessRequestRepository struct {
	mock.Mock
}

func (m *AccessRequestRepository) Create(ctx context.Context, ar *entities.AccessRequest) error {
	return m.Called(ctx, ar).Error(0)
}

func (m *AccessRequestRepository) Find(ctx context.Context, datasetID, userID uint64) (*entities.AccessRequest, error) {
	args := m.Called(ctx, datasetID, userID)
	var ar *entities.AccessRequest
	if value := args.Get(0); value != nil {
		ar, _ = value.(*entities.AccessRequest)
	}
	return ar, args.Error(1)
}

func (m *AccessRequestRepository) ListPendingByOwner(ctx context.Context, ownerID uint64) ([]*entities.AccessRequest, error) {
	args := m.Called(ctx, ownerID)
	var list []*entities.AccessRequest
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.AccessRequest)
	}
	return list, args.Error(1)
}

func (m *AccessRequestRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return m.Called(ctx, id, status).Error(0)
}

func (m *AccessRequestRepository) FindByRequestID(ctx context.Context, requestID uint64) (*entities.AccessRequest, error) {
	args := m.Called(ctx, requestID)
	var ar *entities.AccessRequest
	if value := args.Get(0); value != nil {
		ar, _ = value.(*entities.AccessRequest)
	}
	return ar, args.Error(1)
}

var _ repositories.AccessRequestRepository = (*AccessRequestRepository)(nil)
