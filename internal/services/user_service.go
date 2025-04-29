package services

import "ppo/internal/entities"

type UserService interface {
	Register(user *entities.User) error

	Authenticate(email, password string) (*entities.User, error)

	UpdateUser(user *entities.User) error

	DeleteUser(id uint64) error

	GetUserByID(id uint64) (*entities.User, error)
}
