package dto

import (
	"time"

	"ppo/internal/entities"
)

// NotificationDTO — то, что рендерится в шаблонах списка.
type NotificationDTO struct {
	ID        uint64
	DatasetID uint64
	Message   string
	IsRead    bool
	CreatedAt time.Time
}

// ToNotificationDTO конвертирует entities.Notification в NotificationDTO.
func ToNotificationDTO(n *entities.Notification) *NotificationDTO {
	return &NotificationDTO{
		ID:        n.ID,
		DatasetID: n.DatasetID,
		Message:   n.Message,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt,
	}
}

// ToNotificationDTOs конвертирует срез сущностей в срез DTO.
func ToNotificationDTOs(list []*entities.Notification) []*NotificationDTO {
	res := make([]*NotificationDTO, 0, len(list))
	for _, n := range list {
		res = append(res, ToNotificationDTO(n))
	}
	return res
}

// CreateNotificationForm — данные формы для рассылки уведомления подписчикам.
type CreateNotificationForm struct {
	DatasetID uint64
	Message   string
}
