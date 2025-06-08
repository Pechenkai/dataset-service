package dto

import (
	"time"

	"ppo/internal/entities"
)

type DatasetDTO struct {
	ID           uint64
	Name         string
	Description  string
	OwnerID      uint64
	OwnerName    string
	CategoryID   uint64
	CategoryName string
	DownloadURL  string
	IsPublic     bool
	CreatedAt    time.Time
	AccessStatus string

	IsSubscribed bool
	Reviews      []*ReviewDTO
	HasReviews   bool
}

//func ToDatasetDTO(d *entities.Dataset) *DatasetDTO {
//	return &DatasetDTO{
//		ID:          d.ID,
//		Name:        d.Name,
//		Description: d.Description,
//		OwnerID:     d.OwnerID,
//		CategoryID:  d.CategoryID,
//		IsPublic:    d.IsPublic,
//		CreatedAt:   d.CreatedAt,
//	}
//}

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

func ToDatasetDTOs(list []*entities.Dataset) []*DatasetDTO {
	result := make([]*DatasetDTO, 0, len(list))
	for _, d := range list {
		result = append(result, ToDatasetDTO(d))
	}
	return result
}

type CreateDatasetForm struct {
	Name        string
	Description string
	CategoryID  uint64
	IsPublic    bool
	MetaFormat  string
	MetaTags    string
	MetaSize    uint64
}

type UpdateDatasetForm struct {
	ID          uint64
	Name        string
	Description string
	CategoryID  uint64
	IsPublic    bool
}

type VersionDTO struct {
	ID          uint64
	Number      string
	ChangeLog   string
	CreatedAt   time.Time
	DownloadURL string
}

func ToVersionDTOs(src []*entities.DatasetVersion) []*VersionDTO {
	out := make([]*VersionDTO, len(src))
	for i, v := range src {
		out[i] = &VersionDTO{
			ID:        v.ID,
			Number:    v.Number,
			ChangeLog: v.ChangeLog,
			CreatedAt: v.UploadDate,
		}
	}
	return out
}
