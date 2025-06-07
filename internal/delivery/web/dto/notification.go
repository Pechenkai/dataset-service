package dto

import (
	"time"

	"ppo/internal/entities"
)

type NotificationDTO struct {
	ID          uint64
	DatasetID   uint64
	DatasetName string
	Message     string
	IsRead      bool
	CreatedAt   time.Time
}

func ToNotificationDTO(n *entities.Notification) *NotificationDTO {
	return &NotificationDTO{
		ID:        n.ID,
		DatasetID: n.DatasetID,
		Message:   n.Message,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt,
	}
}

func ToNotificationDTOs(list []*entities.Notification) []*NotificationDTO {
	res := make([]*NotificationDTO, 0, len(list))
	for _, n := range list {
		res = append(res, ToNotificationDTO(n))
	}
	return res
}

type CreateNotificationForm struct {
	DatasetID uint64
	Message   string
}
