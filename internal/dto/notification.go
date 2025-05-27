package dto

// NotificationResponse — для GET /api/v1/notifications
type NotificationResponse struct {
	ID        uint64 `json:"id"`
	DatasetID uint64 `json:"dataset_id"`
	Message   string `json:"message"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

// ListNotificationsResponse
type ListNotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
}
