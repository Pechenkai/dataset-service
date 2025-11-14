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
	GetNotificationByID(ctx context.Context, notificationID uint64) (*entities.Notification, error)
	MarkAsRead(ctx context.Context, notificationID uint64) error
	SetReadStatus(ctx context.Context, notificationID uint64, isRead bool) error
	NotifyUser(ctx context.Context, userID, datasetID uint64, message string) error
}
