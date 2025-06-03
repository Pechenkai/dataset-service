package services

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type userService struct {
	repo   repositories.UserRepository
	logger *zap.Logger
}

// NewUserService создаёт экземпляр UserService с привязанным логгером.
func NewUserService(repo repositories.UserRepository, logger *zap.Logger) UserService {
	logger.Debug("NewUserService initialized")
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

// Register регистрирует нового пользователя.
func (s *userService) Register(ctx context.Context, cmd RegisterUserCmd) (uint64, error) {
	s.logger.Debug("Register called",
		zap.String("username", cmd.Username),
		zap.String("email", cmd.Email),
		zap.String("country", cmd.Country),
		zap.String("role", cmd.Role),
	)

	// 1. Валидируем входные данные через фабрику сущности
	user, err := entities.NewUser(cmd.Username, cmd.Email, cmd.Password, cmd.Country, cmd.Role, time.Now().UTC())
	if err != nil {
		s.logger.Error("failed to validate user data",
			zap.Error(err),
			zap.String("username", cmd.Username),
			zap.String("email", cmd.Email),
		)
		return 0, fmt.Errorf("invalid user data: %w", err)
	}
	s.logger.Debug("user entity constructed",
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)

	// 2. Проверяем, нет ли уже пользователя с таким email
	existing, err := s.repo.FindByEmail(ctx, user.Email)
	if err != nil {
		s.logger.Error("error checking existing user by email",
			zap.Error(err),
			zap.String("email", user.Email),
		)
		return 0, fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		s.logger.Info("user with this email already exists",
			zap.String("email", user.Email),
		)
		return 0, ErrUserExists
	}

	// 3. Хешируем пароль в bcrypt
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password",
			zap.Error(err),
		)
		return 0, fmt.Errorf("hash password: %w", err)
	}
	user.Password = string(hashed)
	s.logger.Debug("password hashed for user", zap.String("email", user.Email))

	// 4. Сохраняем в БД
	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, repositories.ErrEmailAlreadyExists) {
			s.logger.Warn("repository indicates email already exists",
				zap.String("email", user.Email),
			)
			return 0, ErrUserExists
		}
		s.logger.Error("failed to create user in repository",
			zap.Error(err),
			zap.String("email", user.Email),
		)
		return 0, fmt.Errorf("create user: %w", err)
	}

	s.logger.Info("user registered successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("email", user.Email),
	)
	return user.ID, nil
}

// Authenticate проверяет, что email+пароль совпадают, и возвращает сущность User.
func (s *userService) Authenticate(ctx context.Context, cmd AuthenticateUserCmd) (*entities.User, error) {
	s.logger.Debug("Authenticate called",
		zap.String("email", cmd.Email),
	)

	// 1. Получаем пользователя по email
	user, err := s.repo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		s.logger.Error("error fetching user by email",
			zap.Error(err),
			zap.String("email", cmd.Email),
		)
		return nil, fmt.Errorf("fetch user: %w", err)
	}
	if user == nil {
		s.logger.Warn("user not found during authentication",
			zap.String("email", cmd.Email),
		)
		return nil, ErrUserNotFound
	}

	// 2. Сравниваем хэш пароля с введённым
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cmd.Password)); err != nil {
		s.logger.Warn("invalid password attempt",
			zap.String("email", cmd.Email),
		)
		return nil, ErrInvalidCredentials
	}

	s.logger.Info("user authenticated successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("email", user.Email),
	)
	return user, nil
}

// UpdateUser обновляет профиль пользователя:
// – если email изменился, проверяем, что новый email не занят;
// – при необходимости хешируем новый пароль.
func (s *userService) UpdateUser(ctx context.Context, cmd UpdateUserCmd) error {
	s.logger.Debug("UpdateUser called",
		zap.Uint64("user_id", cmd.ID),
		zap.String("new_email", cmd.Email),
		zap.String("new_username", cmd.Username),
	)

	// 1. Получаем существующего пользователя
	user, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		s.logger.Error("error fetching user by ID",
			zap.Error(err),
			zap.Uint64("user_id", cmd.ID),
		)
		return fmt.Errorf("fetch user: %w", err)
	}
	if user == nil {
		s.logger.Warn("user not found for update",
			zap.Uint64("user_id", cmd.ID),
		)
		return ErrUserNotFound
	}
	s.logger.Debug("fetched user for update",
		zap.Uint64("user_id", user.ID),
		zap.String("current_email", user.Email),
	)

	// 2. Обновляем username, если передано
	if cmd.Username != "" {
		user.Username = cmd.Username
	}

	// 3. Если email поменялся, проверяем дублирование
	if cmd.Email != "" && cmd.Email != user.Email {
		s.logger.Debug("email change detected, checking duplicate",
			zap.Uint64("user_id", user.ID),
			zap.String("old_email", user.Email),
			zap.String("new_email", cmd.Email),
		)
		dup, err := s.repo.FindByEmail(ctx, cmd.Email)
		if err != nil {
			s.logger.Error("error checking duplicate email",
				zap.Error(err),
				zap.String("new_email", cmd.Email),
			)
			return fmt.Errorf("check email duplicate: %w", err)
		}
		if dup != nil {
			s.logger.Info("another user already has this email",
				zap.String("email", cmd.Email),
			)
			return ErrUserExists
		}
		user.Email = cmd.Email
	}

	// 4. Если передали новый пароль – хешируем
	if cmd.Password != "" {
		s.logger.Debug("hashing new password for user", zap.Uint64("user_id", user.ID))
		hashed, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("failed to hash new password",
				zap.Error(err),
				zap.Uint64("user_id", user.ID),
			)
			return fmt.Errorf("hash new password: %w", err)
		}
		user.Password = string(hashed)
	}

	// 5. Обновляем остальные поля
	user.Country = cmd.Country
	user.IsBlocked = cmd.IsBlocked
	user.Role = cmd.Role

	// 6. Сохраняем изменения
	if err := s.repo.Update(ctx, user); err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			s.logger.Warn("user not found during update",
				zap.Uint64("user_id", user.ID),
			)
			return ErrUserNotFound
		}
		s.logger.Error("failed to update user in repository",
			zap.Error(err),
			zap.Uint64("user_id", user.ID),
		)
		return fmt.Errorf("update user: %w", err)
	}

	s.logger.Info("user updated successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("email", user.Email),
	)
	return nil
}

// DeleteUser удаляет пользователя по ID.
func (s *userService) DeleteUser(ctx context.Context, id uint64) error {
	s.logger.Debug("DeleteUser called", zap.Uint64("user_id", id))

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			s.logger.Warn("user not found during delete", zap.Uint64("user_id", id))
			return ErrUserNotFound
		}
		s.logger.Error("failed to delete user from repository",
			zap.Error(err),
			zap.Uint64("user_id", id),
		)
		return fmt.Errorf("delete user: %w", err)
	}

	s.logger.Info("user deleted successfully", zap.Uint64("user_id", id))
	return nil
}

// GetUserByID возвращает пользователя по его ID.
func (s *userService) GetUserByID(ctx context.Context, id uint64) (*entities.User, error) {
	s.logger.Debug("GetUserByID called", zap.Uint64("user_id", id))

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("error fetching user by ID",
			zap.Error(err),
			zap.Uint64("user_id", id),
		)
		return nil, fmt.Errorf("fetch user: %w", err)
	}
	if user == nil {
		s.logger.Warn("user not found", zap.Uint64("user_id", id))
		return nil, ErrUserNotFound
	}

	s.logger.Info("user fetched successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("username", user.Username),
	)
	return user, nil
}

// GetUserByID возвращает пользователя по его ID.
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	s.logger.Debug("GetUserByEmail called", zap.String("user_email", email))

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Error("error fetching user by Email",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, fmt.Errorf("fetch user: %w", err)
	}
	if user == nil {
		s.logger.Warn("user not found", zap.String("email", email))
		return nil, ErrUserNotFound
	}

	s.logger.Info("user fetched successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("username", user.Username),
	)
	return user, nil
}
