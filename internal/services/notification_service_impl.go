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

// NewNotificationService создаёт экземпляр NotificationService с привязанным логгером.
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

// NotifySubscribers рассылает сообщение всем подписчикам указанного датасета.
func (s *notificationService) NotifySubscribers(ctx context.Context, cmd NotifySubscribersCmd) (int, error) {
	s.logger.Debug("NotifySubscribers called",
		zap.Uint64("dataset_id", cmd.DatasetID),
		zap.String("message", cmd.Message),
	)

	// 1) Получаем список userID подписчиков из репозитория
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
	for _, userID := range subscribers {
		// 2) Создаём новую сущность Notification
		notif, err := entities.NewNotification(userID, cmd.DatasetID, cmd.Message, time.Now().UTC())
		if err != nil {
			s.logger.Error("failed to construct notification entity",
				zap.Error(err),
				zap.Uint64("user_id", userID),
				zap.Uint64("dataset_id", cmd.DatasetID),
			)
			// Возвращаем количество уже успешно созданных уведомлений
			return count, fmt.Errorf("invalid notification for user %d: %w", userID, err)
		}
		s.logger.Debug("notification entity constructed",
			zap.Uint64("user_id", userID),
			zap.Uint64("dataset_id", cmd.DatasetID),
		)

		// 3) Пытаемся сохранить уведомление через репозиторий
		if err := s.notifRepo.Create(ctx, notif); err != nil {
			s.logger.Error("failed to create notification in repository",
				zap.Error(err),
				zap.Uint64("user_id", userID),
				zap.Uint64("dataset_id", cmd.DatasetID),
			)
			// Возвращаем количество уже созданных (успешных) уведомлений
			return count, fmt.Errorf("create notification for user %d: %w", userID, err)
		}

		count++
		s.logger.Info("notification created",
			zap.Uint64("user_id", userID),
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

// GetNotificationsByUser возвращает все уведомления для данного пользователя.
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

// MarkAsRead устанавливает флаг «прочитано» для конкретного уведомления.
func (s *notificationService) MarkAsRead(ctx context.Context, notificationID uint64) error {
	s.logger.Debug("MarkAsRead called",
		zap.Uint64("notification_id", notificationID),
	)

	// 1) Сначала пытаемся получить уведомление по ID
	notif, err := s.notifRepo.FindByID(ctx, notificationID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotificationNotFound) {
			s.logger.Warn("notification not found when fetching",
				zap.Uint64("notification_id", notificationID),
			)
			return ErrNotificationNotFound
		}
		s.logger.Error("failed to fetch notification",
			zap.Error(err),
			zap.Uint64("notification_id", notificationID),
		)
		return fmt.Errorf("fetch notification: %w", err)
	}
	if notif == nil {
		// На всякий случай: если репозиторий вернул (nil, nil)
		s.logger.Warn("notification is nil after fetch",
			zap.Uint64("notification_id", notificationID),
		)
		return ErrNotificationNotFound
	}

	// 2) Отмечаем «прочитано»
	notif.IsRead = true
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
		return fmt.Errorf("mark as read: %w", err)
	}

	s.logger.Info("notification marked as read successfully",
		zap.Uint64("notification_id", notificationID),
	)
	return nil
}
