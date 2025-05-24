package mocks

import (
	"context"

	"ppo/internal/entities"

	"github.com/stretchr/testify/mock"
)

type ReviewRepository struct {
	mock.Mock
}

func (m *ReviewRepository) Create(ctx context.Context, rev *entities.Review) error {
	args := m.Called(ctx, rev)
	return args.Error(0)
}

func (m *ReviewRepository) FindByID(ctx context.Context, id uint64) (*entities.Review, error) {
	args := m.Called(ctx, id)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*entities.Review), args.Error(1)
}

func (m *ReviewRepository) Update(ctx context.Context, rev *entities.Review) error {
	args := m.Called(ctx, rev)
	return args.Error(0)
}

func (m *ReviewRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *ReviewRepository) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Review, error) {
	args := m.Called(ctx, datasetID)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]*entities.Review), args.Error(1)
}

func (m *ReviewRepository) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Review, error) {
	args := m.Called(ctx, userID)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]*entities.Review), args.Error(1)
}
