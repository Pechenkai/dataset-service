package services

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"time"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type notificationService struct {
	notifRepo repositories.NotificationRepository
	subRepo   repositories.SubscriptionRepository
	logger    *zap.Logger
}

func NewNotificationService(
	notifRepo repositories.NotificationRepository,
	subRepo repositories.SubscriptionRepository,
	logger *zap.Logger,
) NotificationService {
	logger.Debug("NewNotificationService initialized")
	return &notificationService{
		notifRepo: notifRepo,
		subRepo:   subRepo,
		logger:    logger,
	}
}

func (s *notificationService) NotifySubscribers(ctx context.Context, cmd NotifySubscribersCmd) (int, error) {
	s.logger.Debug("NotifySubscribers called",
		zap.Uint64("dataset_id", cmd.DatasetID),
		zap.String("message", cmd.Message),
	)

	subscribers, err := s.subRepo.GetSubscribers(ctx, cmd.DatasetID)
	if err != nil {
		s.logger.Error("failed to fetch subscribers",
			zap.Error(err),
			zap.Uint64("dataset_id", cmd.DatasetID),
		)
		return 0, fmt.Errorf("fetch subscribers: %w", err)
	}

	if len(subscribers) == 0 {
		s.logger.Info("no subscribers found for dataset",
			zap.Uint64("dataset_id", cmd.DatasetID),
		)
		return 0, ErrNoSubscribers
	}

	count := 0
	for _, sub := range subscribers {
		notif, err := entities.NewNotification(sub.UserID, cmd.DatasetID, cmd.Message, time.Now().UTC())
		if err != nil {
			s.logger.Error("failed to construct notification entity",
				zap.Error(err),
				zap.Uint64("user_id", sub.UserID),
				zap.Uint64("dataset_id", cmd.DatasetID),
			)
			return count, fmt.Errorf("invalid notification for user %d: %w", sub.UserID, err)
		}
		s.logger.Debug("notification entity constructed",
			zap.Uint64("user_id", sub.UserID),
			zap.Uint64("dataset_id", cmd.DatasetID),
		)

		if err := s.notifRepo.Create(ctx, notif); err != nil {
			s.logger.Error("failed to create notification in repository",
				zap.Error(err),
				zap.Uint64("user_id", sub.UserID),
				zap.Uint64("dataset_id", cmd.DatasetID),
			)
			return count, fmt.Errorf("create notification for user %d: %w", sub.UserID, err)
		}

		count++
		s.logger.Info("notification created",
			zap.Uint64("user_id", sub.UserID),
			zap.Uint64("dataset_id", cmd.DatasetID),
			zap.Uint64("notification_id", notif.ID),
		)
	}

	s.logger.Info("NotifySubscribers completed successfully",
		zap.Uint64("dataset_id", cmd.DatasetID),
		zap.Int("notified_count", count),
	)
	return count, nil
}

func (s *notificationService) GetNotificationsByUser(ctx context.Context, userID uint64) ([]*entities.Notification, error) {
	s.logger.Debug("GetNotificationsByUser called",
		zap.Uint64("user_id", userID),
	)

	notifs, err := s.notifRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to list notifications",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	s.logger.Info("notifications fetched successfully",
		zap.Uint64("user_id", userID),
		zap.Int("count", len(notifs)),
	)
	return notifs, nil
}

func (s *notificationService) GetNotificationByID(ctx context.Context, notificationID uint64) (*entities.Notification, error) {
	s.logger.Debug("GetNotificationByID called",
		zap.Uint64("notification_id", notificationID),
	)

	notif, err := s.notifRepo.FindByID(ctx, notificationID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotificationNotFound) {
			s.logger.Warn("notification not found when fetching",
				zap.Uint64("notification_id", notificationID),
			)
			return nil, ErrNotificationNotFound
		}
		s.logger.Error("failed to fetch notification",
			zap.Error(err),
			zap.Uint64("notification_id", notificationID),
		)
		return nil, fmt.Errorf("fetch notification: %w", err)
	}
	if notif == nil {
		s.logger.Warn("notification is nil after fetch",
			zap.Uint64("notification_id", notificationID),
		)
		return nil, ErrNotificationNotFound
	}

	s.logger.Info("notification fetched successfully",
		zap.Uint64("notification_id", notif.ID),
		zap.Uint64("user_id", notif.UserID),
	)
	return notif, nil
}

func (s *notificationService) SetReadStatus(ctx context.Context, notificationID uint64, isRead bool) error {
	s.logger.Debug("SetReadStatus called",
		zap.Uint64("notification_id", notificationID),
		zap.Bool("is_read", isRead),
	)

	notif, err := s.GetNotificationByID(ctx, notificationID)
	if err != nil {
		return err
	}

	if notif.IsRead == isRead {
		s.logger.Debug("notification already has requested state",
			zap.Uint64("notification_id", notificationID),
			zap.Bool("is_read", isRead),
		)
		return nil
	}

	notif.IsRead = isRead
	if err := s.notifRepo.Update(ctx, notif); err != nil {
		if errors.Is(err, repositories.ErrNotificationNotFound) {
			s.logger.Warn("notification not found when updating",
				zap.Uint64("notification_id", notificationID),
			)
			return ErrNotificationNotFound
		}
		s.logger.Error("failed to mark notification as read",
			zap.Error(err),
			zap.Uint64("notification_id", notificationID),
		)
		return fmt.Errorf("set read status: %w", err)
	}

	s.logger.Info("notification read status updated",
		zap.Uint64("notification_id", notificationID),
		zap.Bool("is_read", isRead),
	)
	return nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, notificationID uint64) error {
	return s.SetReadStatus(ctx, notificationID, true)
}

func (s *notificationService) NotifyUser(ctx context.Context, userID, datasetID uint64, message string) error {
	s.logger.Debug("NotifyUser called",
		zap.Uint64("user_id", userID),
		zap.Uint64("dataset_id", datasetID),
		zap.String("message", message),
	)

	notif, err := entities.NewNotification(userID, datasetID, message, time.Now().UTC())
	if err != nil {
		s.logger.Error("failed to construct notification entity",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return err
	}

	if err := s.notifRepo.Create(ctx, notif); err != nil {
		s.logger.Error("failed to create notification in repository",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", datasetID),
		)
		return fmt.Errorf("create notification: %w", err)
	}

	s.logger.Info("notification created",
		zap.Uint64("user_id", userID),
		zap.Uint64("notification_id", notif.ID),
		zap.Uint64("dataset_id", datasetID),
	)
	return nil
}
