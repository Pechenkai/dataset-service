package entities

import (
	"strings"
	"time"
)

type Notification struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	DatasetID uint64    `json:"dataset_id"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

func NewNotification(userID, datasetID uint64, message string, createdAt time.Time) (*Notification, error) {
	message = strings.TrimSpace(message)
	if userID == 0 {
		return nil, ErrMissingUserID
	}
	if datasetID == 0 {
		return nil, ErrMissingDatasetID
	}
	if message == "" {
		return nil, ErrEmptyMessage
	}
	if createdAt.After(time.Now()) {
		return nil, ErrNotificationInFuture
	}

	return &Notification{
		UserID:    userID,
		DatasetID: datasetID,
		Message:   message,
		IsRead:    false,
		CreatedAt: createdAt,
	}, nil
}
