package dto

import (
	"time"

	"ppo/internal/entities"
	"ppo/internal/services"
)

type NotifyRequest struct {
	Message string `json:"message" validate:"required"`
}

func (r *NotifyRequest) ToCommand(datasetID uint64) services.NotifySubscribersCmd {
	return services.NotifySubscribersCmd{
		DatasetID: datasetID,
		Message:   r.Message,
	}
}

type NotifyResponse struct {
	Sent int `json:"sent"`
}

type NotificationResponse struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	DatasetID uint64    `json:"dataset_id"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

func FromEntityNotification(n *entities.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		DatasetID: n.DatasetID,
		Message:   n.Message,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt,
	}
}

type NotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
}

func FromEntityList(notifs []*entities.Notification) NotificationsResponse {
	out := make([]NotificationResponse, len(notifs))
	for i, n := range notifs {
		out[i] = FromEntityNotification(n)
	}
	return NotificationsResponse{Notifications: out}
}
