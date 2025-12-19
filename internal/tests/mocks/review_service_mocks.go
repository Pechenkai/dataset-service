package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type ReviewRepository struct {
	mock.Mock
}

func (m *ReviewRepository) Create(ctx context.Context, r *entities.Review) error {
	return m.Mock.Called(ctx, r).Error(0)
}

func (m *ReviewRepository) Delete(ctx context.Context, id uint64) error {
	return m.Mock.Called(ctx, id).Error(0)
}

func (m *ReviewRepository) Update(ctx context.Context, r *entities.Review) error {
	return m.Mock.Called(ctx, r).Error(0)
}

func (m *ReviewRepository) FindByID(ctx context.Context, id uint64) (*entities.Review, error) {
	args := m.Mock.Called(ctx, id)
	var rev *entities.Review
	if value := args.Get(0); value != nil {
		rev, _ = value.(*entities.Review)
	}
	return rev, args.Error(1)
}

func (m *ReviewRepository) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Review, error) {
	args := m.Mock.Called(ctx, datasetID)
	var list []*entities.Review
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Review)
	}
	return list, args.Error(1)
}

func (m *ReviewRepository) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Review, error) {
	args := m.Mock.Called(ctx, userID)
	var list []*entities.Review
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Review)
	}
	return list, args.Error(1)
}

var _ repositories.ReviewRepository = (*ReviewRepository)(nil)
