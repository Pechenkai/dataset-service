package entities

import (
	"strings"
	"time"
)

type Dataset struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uint64    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	CategoryID  uint64    `json:"category_id"`
	IsPublic    bool      `json:"is_public"`
}

func NewDataset(name, description string, ownerID, categoryID uint64, isPublic bool, createdAt time.Time) (*Dataset, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyDatasetName
	}
	if len(name) > 120 {
		return nil, ErrDatasetNameTooLong
	}
	if ownerID == 0 {
		return nil, ErrInvalidOwnerID
	}
	if categoryID == 0 {
		return nil, ErrInvalidCategoryID
	}
	if createdAt.After(time.Now()) {
		return nil, ErrInvalidCreatedAt
	}

	return &Dataset{
		Name:        name,
		Description: strings.TrimSpace(description),
		OwnerID:     ownerID,
		CategoryID:  categoryID,
		IsPublic:    isPublic,
		CreatedAt:   createdAt,
	}, nil
}
