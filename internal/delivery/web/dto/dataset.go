package dto

import (
	"time"

	"ppo/internal/entities"
)

// DatasetDTO — то, что передаётся в шаблоны при отображении списка или деталей.
type DatasetDTO struct {
	ID          uint64
	Name        string
	Description string
	OwnerID     uint64
	CategoryID  uint64
	IsPublic    bool
	CreatedAt   time.Time
}

// ToDatasetDTO конвертирует entities.Dataset в DatasetDTO.
func ToDatasetDTO(d *entities.Dataset) *DatasetDTO {
	return &DatasetDTO{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		OwnerID:     d.OwnerID,
		CategoryID:  d.CategoryID,
		IsPublic:    d.IsPublic,
		CreatedAt:   d.CreatedAt,
	}
}

// ToDatasetDTOs конвертирует срез entities.Dataset в срез DTO.
func ToDatasetDTOs(list []*entities.Dataset) []*DatasetDTO {
	result := make([]*DatasetDTO, 0, len(list))
	for _, d := range list {
		result = append(result, ToDatasetDTO(d))
	}
	return result
}

// CreateDatasetForm представляет поля формы создания нового Dataset.
type CreateDatasetForm struct {
	Name        string
	Description string
	CategoryID  uint64
	IsPublic    bool
}
