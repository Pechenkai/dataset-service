package services

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"ppo/internal/entities"
	"time"

	"ppo/internal/repositories"
)

type subscriptionService struct {
	repo   repositories.SubscriptionRepository
	logger *zap.Logger
}

func NewSubscriptionService(repo repositories.SubscriptionRepository, logger *zap.Logger) SubscriptionService {
	logger.Debug("NewSubscriptionService initialized")
	return &subscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *subscriptionService) Subscribe(ctx context.Context, userID, datasetID uint64) error {
	s.logger.Debug("Subscribe called",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
	)

	if userID == 0 || datasetID == 0 {
		s.logger.Warn("invalid IDs provided for subscribe",
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return fmt.Errorf("invalid IDs: user=%d, dataset=%d", userID, datasetID)
	}

	exists, err := s.repo.IsSubscribed(ctx, userID, datasetID)
	if err != nil {
		s.logger.Error("failed to check existing subscription",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return fmt.Errorf("check subscription: %w", err)
	}
	if exists {
		s.logger.Info("user already subscribed, aborting",
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return ErrAlreadySubscribed
	}

	sub, err := entities.NewSubscription(userID, datasetID, time.Now().UTC())
	if err != nil {
		s.logger.Error("failed to construct subscription entity",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return err
	}
	s.logger.Debug("subscription entity constructed",
		zap.Uint64("user_id", sub.UserID),
		zap.Uint64("dataset_id", sub.DatasetID),
	)

	if err := s.repo.Create(ctx, sub); err != nil {
		if errors.Is(err, repositories.ErrAlreadySubscribed) {
			s.logger.Info("repository indicates already subscribed",
				zap.Uint64("user_id", userID),
				zap.Uint64("dataset_id", datasetID),
			)
			return ErrAlreadySubscribed
		}
		s.logger.Error("failed to create subscription in repository",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return fmt.Errorf("subscribe: %w", err)
	}

	s.logger.Info("user successfully subscribed to dataset",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
	)
	return nil
}

func (s *subscriptionService) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	s.logger.Debug("Unsubscribe called",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
	)

	if userID == 0 || datasetID == 0 {
		s.logger.Warn("invalid IDs provided for unsubscribe",
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return fmt.Errorf("invalid IDs: user=%d, dataset=%d", userID, datasetID)
	}

	exists, err := s.repo.IsSubscribed(ctx, userID, datasetID)
	if err != nil {
		s.logger.Error("failed to check existing subscription before unsubscribe",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return fmt.Errorf("check subscription: %w", err)
	}
	if !exists {
		s.logger.Info("user is not subscribed, cannot unsubscribe",
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return ErrNotSubscribed
	}

	if err := s.repo.Unsubscribe(ctx, userID, datasetID); err != nil {
		if errors.Is(err, repositories.ErrSubscriptionNotFound) {
			s.logger.Warn("repository indicates subscription not found during unsubscribe",
				zap.Uint64("user_id", userID),
				zap.Uint64("dataset_id", datasetID),
			)
			return ErrNotSubscribed
		}
		s.logger.Error("failed to unsubscribe in repository",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return fmt.Errorf("unsubscribe: %w", err)
	}

	s.logger.Info("user successfully unsubscribed from dataset",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
	)
	return nil
}

func (s *subscriptionService) ListSubscribers(ctx context.Context, datasetID uint64) ([]*entities.Subscription, error) {
	s.logger.Debug("ListSubscribers called",
		zap.Uint64("dataset_id", datasetID),
	)

	if datasetID == 0 {
		s.logger.Warn("invalid datasetID provided for ListSubscribers",
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, fmt.Errorf("invalid datasetID: %d", datasetID)
	}

	subs, err := s.repo.GetSubscribers(ctx, datasetID)
	if err != nil {
		s.logger.Error("failed to list subscribers from repository",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, fmt.Errorf("list subscribers: %w", err)
	}

	s.logger.Info("subscribers list fetched successfully",
		zap.Uint64("dataset_id", datasetID),
		zap.Int("count", len(subs)),
	)
	return subs, nil
}

func (s *subscriptionService) ListSubscriptions(ctx context.Context, userID uint64) ([]*entities.Subscription, error) {
	s.logger.Debug("ListSubscriptions called",
		zap.Uint64("user_id", userID),
	)

	if userID == 0 {
		s.logger.Warn("invalid userID provided for ListSubscriptions",
			zap.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("invalid userID: %d", userID)
	}

	ds, err := s.repo.GetByUser(ctx, userID)
	if err != nil {
		s.logger.Error("failed to list subscriptions from repository",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}

	s.logger.Info("subscriptions list fetched successfully",
		zap.Uint64("user_id", userID),
		zap.Int("count", len(ds)),
	)
	return ds, nil
}

func (s *subscriptionService) ListAllSubscriptions(ctx context.Context) ([]*entities.Subscription, error) {
	s.logger.Debug("ListAllSubscriptions called")

	subs, err := s.repo.GetAll(ctx)
	if err != nil {
		s.logger.Error("failed to list all subscriptions from repository", zap.Error(err))
		return nil, fmt.Errorf("list all subscriptions: %w", err)
	}

	s.logger.Info("all subscriptions fetched", zap.Int("count", len(subs)))
	return subs, nil
}

func (s *subscriptionService) IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error) {
	s.logger.Debug("IsSubscribed called", zap.Uint64("user_id", userID), zap.Uint64("dataset_id", datasetID))

	exists, err := s.repo.IsSubscribed(ctx, userID, datasetID)
	if err != nil {
		s.logger.Error("IsSubscribed failed", zap.Error(err), zap.Uint64("user_id", userID), zap.Uint64("dataset_id", datasetID))
		return false, fmt.Errorf("check subscription: %w", err)
	}
	return exists, nil
}
