package repositories

import (
	"context"
	"ppo/internal/entities"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *entities.Notification) error

	FindByID(ctx context.Context, id uint64) (*entities.Notification, error)

	FindByUserID(ctx context.Context, userID uint64) ([]*entities.Notification, error)

	Update(ctx context.Context, n *entities.Notification) error

	Delete(ctx context.Context, id uint64) error
}
