package services

import (
	"context"

	"ppo/internal/entities"
)

type RegisterUserCmd struct {
	Username string
	Email    string
	Password string
	Country  string
	Role     string
}

type AuthenticateUserCmd struct {
	Email    string
	Password string
}

type UpdateUserCmd struct {
	ID        uint64
	Username  string
	Email     string
	Password  string // optional: current or new password
	Country   string
	IsBlocked bool
	Role      string
}

type UserService interface {
	Register(ctx context.Context, cmd RegisterUserCmd) (uint64, error)
	Authenticate(ctx context.Context, cmd AuthenticateUserCmd) (*entities.User, error)
	UpdateUser(ctx context.Context, cmd UpdateUserCmd) error
	DeleteUser(ctx context.Context, id uint64) error
	GetUserByID(ctx context.Context, id uint64) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
}
