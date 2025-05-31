package dto

import (
	"time"

	"ppo/internal/entities"
)

type DatasetResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uint64    `json:"owner_id"`
	CategoryID  uint64    `json:"category_id"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
}

func FromDataset(e *entities.Dataset) DatasetResponse {
	return DatasetResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		OwnerID:     e.OwnerID,
		CategoryID:  e.CategoryID,
		IsPublic:    e.IsPublic,
		CreatedAt:   e.CreatedAt,
	}
}

type DatasetsResponse struct {
	Datasets []DatasetResponse `json:"datasets"`
}

func FromDatasetList(list []*entities.Dataset) DatasetsResponse {
	resp := make([]DatasetResponse, len(list))
	for i, e := range list {
		resp[i] = FromDataset(e)
	}
	return DatasetsResponse{Datasets: resp}
}

type VersionResponse struct {
	ID         uint64    `json:"id"`
	Number     string    `json:"number"`
	Filepath   string    `json:"filepath"`
	ChangeLog  string    `json:"change_log"`
	UploadDate time.Time `json:"upload_date"`
}

func FromVersion(v *entities.DatasetVersion) VersionResponse {
	return VersionResponse{
		ID:         v.ID,
		Number:     v.Number,
		Filepath:   v.Filepath,
		ChangeLog:  v.ChangeLog,
		UploadDate: v.UploadDate,
	}
}

type VersionsResponse struct {
	Versions []VersionResponse `json:"versions"`
}

func FromVersionList(list []*entities.DatasetVersion) VersionsResponse {
	resp := make([]VersionResponse, len(list))
	for i, v := range list {
		resp[i] = FromVersion(v)
	}
	return VersionsResponse{Versions: resp}
}
