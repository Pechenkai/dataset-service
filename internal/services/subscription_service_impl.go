package services

import (
	"context"
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
	if userID == 0 || datasetID == 0 {
		return fmt.Errorf("invalid IDs: user=%d, dataset=%d", userID, datasetID)
	}
	exists, err := s.repo.IsSubscribed(ctx, userID, datasetID)
	if err != nil {
		return fmt.Errorf("check subscription: %w", err)
	}
	if exists {
		return ErrAlreadySubscribed
	}

	sub, err := entities.NewSubscription(userID, datasetID, time.Now().UTC())
	if err != nil {
		return err
	}
	return s.repo.Create(ctx, sub)
}

func (s *subscriptionService) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	if userID == 0 || datasetID == 0 {
		return fmt.Errorf("invalid IDs: user=%d, dataset=%d", userID, datasetID)
	}
	exists, err := s.repo.IsSubscribed(ctx, userID, datasetID)
	if err != nil {
		return fmt.Errorf("check subscription: %w", err)
	}
	if !exists {
		return ErrNotSubscribed
	}
	if err := s.repo.Unsubscribe(ctx, userID, datasetID); err != nil {
		return fmt.Errorf("unsubscribe: %w", err)
	}
	return nil
}

func (s *subscriptionService) ListSubscribers(ctx context.Context, datasetID uint64) ([]uint64, error) {
	if datasetID == 0 {
		return nil, fmt.Errorf("invalid datasetID: %d", datasetID)
	}
	subs, err := s.repo.GetSubscribers(ctx, datasetID)
	if err != nil {
		return nil, fmt.Errorf("list subscribers: %w", err)
	}
	return subs, nil
}

func (s *subscriptionService) ListSubscriptions(ctx context.Context, userID uint64) ([]uint64, error) {
	if userID == 0 {
		return nil, fmt.Errorf("invalid userID: %d", userID)
	}
	ds, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	return ds, nil
}
