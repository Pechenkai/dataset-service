package entities

import (
	"fmt"
	"time"
)

type AccessStatus string

const (
	AccessStatusPending  AccessStatus = "pending"
	AccessStatusApproved AccessStatus = "approved"
	AccessStatusDenied   AccessStatus = "denied"
)

var ValidAccessStatuses = map[AccessStatus]struct{}{
	AccessStatusPending:  {},
	AccessStatusApproved: {},
	AccessStatusDenied:   {},
}

type AccessRequest struct {
	ID        uint64       `json:"id"`
	DatasetID uint64       `json:"dataset_id"`
	UserID    uint64       `json:"user_id"`
	Status    AccessStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
}

func NewAccessRequest(dsID, userID uint64) (*AccessRequest, error) {
	if dsID == 0 || userID == 0 {
		return nil, fmt.Errorf("invalid dataset or user ID: %d, %d", dsID, userID)
	}
	return &AccessRequest{
		DatasetID: dsID,
		UserID:    userID,
		Status:    AccessStatusPending,
		CreatedAt: time.Now().UTC(),
	}, nil
}
