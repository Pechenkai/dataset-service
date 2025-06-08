package dto

import (
	"ppo/internal/entities"
	"time"
)

type AccessRequestDTO struct {
	ID          uint64
	DatasetID   uint64
	DatasetName string
	UserID      uint64
	Username    string
	Status      string
	CreatedAt   time.Time
}

func ToAccessRequestDTO(ar *entities.AccessRequest, datasetName, username string) *AccessRequestDTO {
	return &AccessRequestDTO{
		ID:          ar.ID,
		DatasetID:   ar.DatasetID,
		DatasetName: datasetName,
		UserID:      ar.UserID,
		Username:    username,
		Status:      string(ar.Status),
		CreatedAt:   ar.CreatedAt,
	}
}
