package services

import (
	"context"

	"ppo/internal/entities"
)

type NotifySubscribersCmd struct {
	DatasetID uint64
	Message   string
}

type NotificationService interface {
	NotifySubscribers(ctx context.Context, cmd NotifySubscribersCmd) (int, error)
	GetNotificationsByUser(ctx context.Context, userID uint64) ([]*entities.Notification, error)
	MarkAsRead(ctx context.Context, notificationID uint64) error
}
