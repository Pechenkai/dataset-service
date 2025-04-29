package entities

import (
	"time"
)

type Subscription struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	DatasetID uint64    `json:"dataset_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewSubscription(userID uint64, datasetID uint64, createdAt time.Time) (*Subscription, error) {
	if userID == 0 {
		return nil, ErrMissingUserID
	}
	if datasetID == 0 {
		return nil, ErrInvalidDatasetID
	}
	if createdAt.After(time.Now()) {
		return nil, ErrInvalidCreatedAt
	}
	return &Subscription{
		UserID:    userID,
		DatasetID: datasetID,
		CreatedAt: createdAt,
	}, nil
}
