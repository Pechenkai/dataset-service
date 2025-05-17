package repositories

import (
	"context"
	"ppo/internal/entities"
)

type UserRepository interface {
	Create(ctx context.Context, u *entities.User) error
	Update(ctx context.Context, u *entities.User) error
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*entities.User, error)
	FindAll(ctx context.Context) ([]*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
}
