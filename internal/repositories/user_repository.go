package repositories

import "ppo/internal/entities"

type UserRepository interface {
	Create(*entities.User) error
	Update(*entities.User) error
	Delete(id uint64) error
	FindByID(id uint64) (*entities.User, error)
	FindAll() ([]*entities.User, error)
	FindByEmail(email string) (*entities.User, error)
}
