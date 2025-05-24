package mocks

import (
	"context"
	"ppo/internal/entities"

	"github.com/stretchr/testify/mock"
)

type CategoryRepository struct {
	mock.Mock
}

func (m *CategoryRepository) Create(ctx context.Context, cat *entities.Category) error {
	args := m.Called(ctx, cat)
	return args.Error(0)
}

func (m *CategoryRepository) Update(ctx context.Context, cat *entities.Category) error {
	args := m.Called(ctx, cat)
	return args.Error(0)
}

func (m *CategoryRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *CategoryRepository) FindByID(ctx context.Context, id uint64) (*entities.Category, error) {
	args := m.Called(ctx, id)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*entities.Category), args.Error(1)
}

func (m *CategoryRepository) FindAll(ctx context.Context) ([]*entities.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.Category), args.Error(1)
}
