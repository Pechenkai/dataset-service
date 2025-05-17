package services

import (
	"context"

	"ppo/internal/entities"
)

type CreateDatasetCmd struct {
	ActorID     uint64
	Name        string
	Description string
	CategoryID  uint64
	IsPublic    bool

	MetaFormat string
	MetaTags   string
	MetaSize   uint64
}

type AddVersionCmd struct {
	ActorID   uint64
	DatasetID uint64
	ChangeLog string
	FilePath  string

	MetaFormat string
	MetaTags   string
	MetaSize   uint64
}

type DatasetService interface {
	CreateDataset(ctx context.Context, cmd CreateDatasetCmd) (uint64, error)
	AddDatasetVersion(ctx context.Context, cmd AddVersionCmd) (uint64, error)
	GetDataset(ctx context.Context, id uint64) (*entities.Dataset, error)
}
