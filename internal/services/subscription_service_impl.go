package services

import (
	"context"
	"errors"
	"fmt"
	"ppo/internal/entities"
	"time"

	"ppo/internal/repositories"
)

type subscriptionService struct {
	repo repositories.SubscriptionRepository
}

func NewSubscriptionService(repo repositories.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{repo: repo}
}

func (s *subscriptionService) Subscribe(ctx context.Context, userID, datasetID uint64) error {
	// Проверяем, что IDs валидны
	if userID == 0 || datasetID == 0 {
		return fmt.Errorf("invalid IDs: user=%d, dataset=%d", userID, datasetID)
	}
	// Сначала проверяем, есть ли уже подписка
	exists, err := s.repo.IsSubscribed(ctx, userID, datasetID)
	if err != nil {
		return fmt.Errorf("check subscription: %w", err)
	}
	if exists {
		return ErrAlreadySubscribed
	}

	// Конструируем объект подписки (проверка внутри NewSubscription)
	sub, err := entities.NewSubscription(userID, datasetID, time.Now().UTC())
	if err != nil {
		// NewSubscription вернёт ошибку, если, например, ID некорректны,
		// но это маловероятно, ведь мы их уже проверили выше.
		// Всё равно просто возвращаем её напрямую.
		return err
	}

	// Пытаемся сохранить подписку в репозитории
	if err := s.repo.Create(ctx, sub); err != nil {
		// Если репозиторий вернул ErrAlreadySubscribed (duplicate key),
		// то пробрасываем сервисную ErrAlreadySubscribed
		if errors.Is(err, repositories.ErrAlreadySubscribed) {
			return ErrAlreadySubscribed
		}
		// Иначе — оборачиваем любую другую ошибку
		return fmt.Errorf("subscribe: %w", err)
	}

	return nil
}

func (s *subscriptionService) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	// Проверяем, что IDs валидны
	if userID == 0 || datasetID == 0 {
		return fmt.Errorf("invalid IDs: user=%d, dataset=%d", userID, datasetID)
	}
	// Проверяем, существует ли подписка
	exists, err := s.repo.IsSubscribed(ctx, userID, datasetID)
	if err != nil {
		return fmt.Errorf("check subscription: %w", err)
	}
	if !exists {
		return ErrNotSubscribed
	}

	// Пытаемся удалить подписку
	if err := s.repo.Unsubscribe(ctx, userID, datasetID); err != nil {
		// Если репозиторий возвращает ErrSubscriptionNotFound, пробрасываем сервисную ErrNotSubscribed
		if errors.Is(err, repositories.ErrSubscriptionNotFound) {
			return ErrNotSubscribed
		}
		// Иначе — оборачиваем любую другую ошибку
		return fmt.Errorf("unsubscribe: %w", err)
	}

	return nil
}

func (s *subscriptionService) ListSubscribers(ctx context.Context, datasetID uint64) ([]uint64, error) {
	// Проверяем корректность datasetID
	if datasetID == 0 {
		return nil, fmt.Errorf("invalid datasetID: %d", datasetID)
	}

	// Запрашиваем список userID, подписанных на датасет
	subs, err := s.repo.GetSubscribers(ctx, datasetID)
	if err != nil {
		return nil, fmt.Errorf("list subscribers: %w", err)
	}
	return subs, nil
}

func (s *subscriptionService) ListSubscriptions(ctx context.Context, userID uint64) ([]uint64, error) {
	// Проверяем корректность userID
	if userID == 0 {
		return nil, fmt.Errorf("invalid userID: %d", userID)
	}

	// Запрашиваем список datasetID, на которые подписан пользователь
	ds, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	return ds, nil
}
