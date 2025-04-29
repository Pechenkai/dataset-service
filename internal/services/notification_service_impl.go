package services

import (
	"time"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type notificationService struct {
	notificationRepo repositories.NotificationRepository
	subscriptionRepo repositories.SubscriptionRepository
}

func NewNotificationService(
	notificationRepo repositories.NotificationRepository,
	subscriptionRepo repositories.SubscriptionRepository,
) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *notificationService) NotifySubscribers(datasetID uint64, message string) error {
	subscribers, err := s.getSubscribersForDataset(datasetID)
	if err != nil {
		return err
	}

	for _, userID := range subscribers {
		notification, err := entities.NewNotification(userID, datasetID, message, time.Now())

		if err != nil {
			return err
		}

		if err := s.notificationRepo.Create(notification); err != nil {
			return err
		}
	}
	return nil
}

func (s *notificationService) getSubscribersForDataset(datasetID uint64) ([]uint64, error) {
	return s.subscriptionRepo.GetSubscribers(datasetID)
}

func (s *notificationService) GetNotificationsByUser(userID uint64) ([]*entities.Notification, error) {
	return s.notificationRepo.FindByUserID(userID)
}

func (s *notificationService) MarkAsRead(notificationID uint64) error {
	notification, err := s.notificationRepo.FindByID(notificationID)
	if err != nil {
		return err
	}
	if notification == nil {
		return ErrNotificationFound
	}
	notification.IsRead = true
	return s.notificationRepo.Update(notification)
}
