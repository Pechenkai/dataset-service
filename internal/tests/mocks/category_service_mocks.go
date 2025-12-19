package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type CategoryRepository struct {
	mock.Mock
}

func (m *CategoryRepository) Create(ctx context.Context, c *entities.Category) error {
	return m.Mock.Called(ctx, c).Error(0)
}

func (m *CategoryRepository) Delete(ctx context.Context, id uint64) error {
	return m.Mock.Called(ctx, id).Error(0)
}

func (m *CategoryRepository) Update(ctx context.Context, c *entities.Category) error {
	return m.Mock.Called(ctx, c).Error(0)
}

func (m *CategoryRepository) FindByID(ctx context.Context, id uint64) (*entities.Category, error) {
	args := m.Mock.Called(ctx, id)
	var cat *entities.Category
	if value := args.Get(0); value != nil {
		cat, _ = value.(*entities.Category)
	}
	return cat, args.Error(1)
}

func (m *CategoryRepository) FindAll(ctx context.Context) ([]*entities.Category, error) {
	args := m.Mock.Called(ctx)
	var list []*entities.Category
	if value := args.Get(0); value != nil {
		list, _ = value.([]*entities.Category)
	}
	return list, args.Error(1)
}

var _ repositories.CategoryRepository = (*CategoryRepository)(nil)
