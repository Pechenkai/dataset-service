package repositories

import "ppo/internal/entities"

type NotificationRepository interface {
	Create(notification *entities.Notification) error

	FindByID(id uint64) (*entities.Notification, error)

	FindByUserID(userID uint64) ([]*entities.Notification, error)

	Update(notification *entities.Notification) error

	Delete(id uint64) error
}
