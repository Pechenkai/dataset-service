package dto

import (
	"io"
	"ppo/internal/services"
)

type CreateDatasetRequest struct {
	Name        string    `json:"name" form:"name" validate:"required,max=100"`
	Description string    `json:"description" form:"description"`
	CategoryID  uint64    `json:"category_id" form:"category_id" validate:"required"`
	IsPublic    bool      `json:"is_public" form:"is_public"`
	File        io.Reader `json:"-" form:"file"` // файл как поток
	FileSize    int64     `json:"-" form:"-"`    // размер из заголовка multipart
	FileName    string    `json:"-" form:"-"`    // имя файла из header
	MetaFormat  string    `json:"format" form:"format"`
	MetaTags    string    `json:"tags" form:"tags"`
	MetaSize    uint64    `json:"size" form:"size"`
}

func (r *CreateDatasetRequest) ToCommand() services.CreateDatasetCmd {
	return services.CreateDatasetCmd{
		Name:        r.Name,
		Description: r.Description,
		CategoryID:  r.CategoryID,
		IsPublic:    r.IsPublic,
		FileName:    r.FileName,
		MetaFormat:  r.MetaFormat,
		MetaTags:    r.MetaTags,
		MetaSize:    r.MetaSize,
	}
}

type DatasetResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     uint64 `json:"owner_id"`
	CategoryID  uint64 `json:"category_id"`
	IsPublic    bool   `json:"is_public"`
	CreatedAt   string `json:"created_at"`
}

type ListDatasetsResponse struct {
	Datasets []DatasetResponse `json:"datasets"`
}
