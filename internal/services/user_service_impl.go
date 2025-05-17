package services

import (
	"context"
	"errors"
	"fmt"
	"ppo/internal/dataaccess/repositories/postgres"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, cmd RegisterUserCmd) (uint64, error) {
	user, err := entities.NewUser(cmd.Username, cmd.Email, cmd.Password, cmd.Country, cmd.Role, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("invalid user data: %w", err)
	}

	existing, err := s.repo.FindByEmail(ctx, user.Email)
	if err != nil {
		return 0, fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return 0, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}
	user.Password = string(hash)

	if err := s.repo.Create(ctx, user); err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	return user.ID, nil
}

func (s *userService) Authenticate(ctx context.Context, cmd AuthenticateUserCmd) (*entities.User, error) {
	user, err := s.repo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cmd.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, cmd UpdateUserCmd) error {
	user, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("fetch user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if cmd.Username != "" {
		user.Username = cmd.Username
	}
	if cmd.Email != "" && cmd.Email != user.Email {
		if dup, _ := s.repo.FindByEmail(ctx, cmd.Email); dup != nil {
			return ErrUserExists
		}
		user.Email = cmd.Email
	}
	if cmd.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash new password: %w", err)
		}
		user.Password = string(hash)
	}
	user.Country = cmd.Country
	user.IsBlocked = cmd.IsBlocked
	user.Role = cmd.Role

	if err := s.repo.Update(ctx, user); err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint64) (*entities.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
