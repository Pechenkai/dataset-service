package services

import (
	"context"
	"errors"
	"fmt"
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

// Register регистрирует нового пользователя
func (s *userService) Register(ctx context.Context, cmd RegisterUserCmd) (uint64, error) {
	// 1. Валидируем входные данные через фабрику сущности
	user, err := entities.NewUser(cmd.Username, cmd.Email, cmd.Password, cmd.Country, cmd.Role, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("invalid user data: %w", err)
	}

	// 2. Проверяем, нет ли уже пользователя с таким email
	existing, err := s.repo.FindByEmail(ctx, user.Email)
	if err != nil {
		// Если произошла любая ошибка доступа к БД
		return 0, fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return 0, ErrUserExists
	}

	// 3. Хешируем пароль в bcrypt
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}
	user.Password = string(hashed)

	// 4. Сохраняем в БД
	if err := s.repo.Create(ctx, user); err != nil {
		// Если это дублирование по email из репозитория, вернём ErrUserExists
		if errors.Is(err, repositories.ErrEmailAlreadyExists) {
			return 0, ErrUserExists
		}
		return 0, fmt.Errorf("create user: %w", err)
	}

	return user.ID, nil
}

// Authenticate проверяет, что email+пароль совпадают, и возвращает сущность User (без изменения пароля)
func (s *userService) Authenticate(ctx context.Context, cmd AuthenticateUserCmd) (*entities.User, error) {
	user, err := s.repo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Сравниваем хэш пароля с введённым
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cmd.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// UpdateUser обновляет профиль пользователя:
// – если email изменился, проверяем, что новый email не занят;
// – при необходимости хешируем новый пароль.
func (s *userService) UpdateUser(ctx context.Context, cmd UpdateUserCmd) error {
	user, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("fetch user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Обновляем username, если передано
	if cmd.Username != "" {
		user.Username = cmd.Username
	}

	// Если email поменялся, проверяем дублирование
	if cmd.Email != "" && cmd.Email != user.Email {
		dup, err := s.repo.FindByEmail(ctx, cmd.Email)
		if err != nil {
			return fmt.Errorf("check email duplicate: %w", err)
		}
		if dup != nil {
			return ErrUserExists
		}
		user.Email = cmd.Email
	}

	// Если передали новый пароль – хешируем
	if cmd.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash new password: %w", err)
		}
		user.Password = string(hashed)
	}

	// Обновляем остальные поля
	user.Country = cmd.Country
	user.IsBlocked = cmd.IsBlocked
	user.Role = cmd.Role

	// Сохраняем изменения
	if err := s.repo.Update(ctx, user); err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// DeleteUser удаляет пользователя по ID
func (s *userService) DeleteUser(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// GetUserByID возвращает пользователя по ID
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
