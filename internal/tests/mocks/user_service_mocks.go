package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) Create(ctx context.Context, u *entities.User) error {
	return m.Mock.Called(ctx, u).Error(0)
}

func (m *UserRepository) Update(ctx context.Context, u *entities.User) error {
	return m.Mock.Called(ctx, u).Error(0)
}

func (m *UserRepository) Delete(ctx context.Context, id uint64) error {
	return m.Mock.Called(ctx, id).Error(0)
}

func (m *UserRepository) FindByID(ctx context.Context, id uint64) (*entities.User, error) {
	args := m.Mock.Called(ctx, id)
	var user *entities.User
	if value := args.Get(0); value != nil {
		user, _ = value.(*entities.User)
	}
	return user, args.Error(1)
}

func (m *UserRepository) FindAll(ctx context.Context) ([]*entities.User, error) {
	args := m.Mock.Called(ctx)
	var users []*entities.User
	if value := args.Get(0); value != nil {
		users, _ = value.([]*entities.User)
	}
	return users, args.Error(1)
}

func (m *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Mock.Called(ctx, email)
	var user *entities.User
	if value := args.Get(0); value != nil {
		user, _ = value.(*entities.User)
	}
	return user, args.Error(1)
}

var _ repositories.UserRepository = (*UserRepository)(nil)
