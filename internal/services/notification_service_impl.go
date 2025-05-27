package services

import (
	"context"
	"fmt"
	"time"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type notificationService struct {
	notifRepo repositories.NotificationRepository
	subRepo   repositories.SubscriptionRepository
	clock     Clock
}

func NewNotificationService(
	notifRepo repositories.NotificationRepository,
	subRepo repositories.SubscriptionRepository,
) NotificationService {
	return &notificationService{
		notifRepo: notifRepo,
		subRepo:   subRepo,
	}
}

func (s *notificationService) NotifySubscribers(ctx context.Context, datasetID uint64, message string) (int, error) {
	subscribers, err := s.subRepo.GetSubscribers(ctx, datasetID)
	if err != nil {
		return 0, fmt.Errorf("fetch subscribers: %w", err)
	}
	if len(subscribers) == 0 {
		return 0, ErrNoSubscribers
	}

	count := 0
	for _, userID := range subscribers {
		notif, err := entities.NewNotification(userID, datasetID, message, time.Now().UTC())
		if err != nil {
			return count, fmt.Errorf("invalid notification for user %d: %w", userID, err)
		}

		if err := s.notifRepo.Create(ctx, notif); err != nil {
			return count, fmt.Errorf("create notification for user %d: %w", userID, err)
		}
		count++
	}
	return count, nil
}

func (s *notificationService) GetNotificationsByUser(ctx context.Context, userID uint64) ([]*entities.Notification, error) {
	notifs, err := s.notifRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	return notifs, nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, notificationID uint64) error {
	notif, err := s.notifRepo.FindByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("fetch notification: %w", err)
	}
	if notif == nil {
		return ErrNotificationNotFound
	}
	notif.IsRead = true
	if err := s.notifRepo.Update(ctx, notif); err != nil {
		return fmt.Errorf("mark as read: %w", err)
	}
	return nil
}
