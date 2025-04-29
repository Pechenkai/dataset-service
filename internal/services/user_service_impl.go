package services

import (
	"time"

	"ppo/internal/entities"
	"ppo/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) Register(user *entities.User) error {
	if user == nil {
		return ErrNilUser
	}

	existingUser, err := s.userRepo.FindByEmail(user.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)

	if user.RegistrationDate.IsZero() {
		user.RegistrationDate = time.Now()
	}

	return s.userRepo.Create(user)
}

func (s *userService) Authenticate(email, password string) (*entities.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}
	return user, nil
}

func (s *userService) UpdateUser(user *entities.User) error {
	if user == nil {
		return ErrNilUser
	}
	return s.userRepo.Update(user)
}

func (s *userService) DeleteUser(id uint64) error {
	return s.userRepo.Delete(id)
}

func (s *userService) GetUserByID(id uint64) (*entities.User, error) {
	return s.userRepo.FindByID(id)
}
