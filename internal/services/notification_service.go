package services

import "ppo/internal/entities"

type NotificationService interface {
	NotifySubscribers(datasetID uint64, message string) error
	GetNotificationsByUser(userID uint64) ([]*entities.Notification, error)
	MarkAsRead(notificationID uint64) error
}
