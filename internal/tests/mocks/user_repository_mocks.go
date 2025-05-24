package mocks

import (
	"context"

	"ppo/internal/entities"

	"github.com/stretchr/testify/mock"
)

type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) FindAll(ctx context.Context) ([]*entities.User, error) {
	//TODO implement me
	panic("implement me")
}

func (m *UserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*entities.User), args.Error(1)
}

func (m *UserRepository) FindByID(ctx context.Context, id uint64) (*entities.User, error) {
	args := m.Called(ctx, id)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*entities.User), args.Error(1)
}

func (m *UserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *UserRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
