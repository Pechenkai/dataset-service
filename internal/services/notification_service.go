package services

import (
	"context"

	"ppo/internal/entities"
)

type NotificationService interface {
	NotifySubscribers(ctx context.Context, datasetID uint64, message string) (int, error)
	GetNotificationsByUser(ctx context.Context, userID uint64) ([]*entities.Notification, error)
	MarkAsRead(ctx context.Context, notificationID uint64) error
}
