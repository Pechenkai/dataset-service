package services

import (
	"context"
	"io"

	"ppo/internal/entities"
)

type CreateDatasetCmd struct {
	ActorID     uint64
	Name        string
	Description string
	CategoryID  uint64
	FileName    string
	IsPublic    bool

	MetaFormat string
	MetaTags   string
	MetaSize   uint64
}

type AddVersionCmd struct {
	ActorID   uint64
	DatasetID uint64
	ChangeLog string
	FileName  string

	MetaFormat string
	MetaTags   string
	MetaSize   uint64
}

type DatasetService interface {
	CreateDataset(ctx context.Context, cmd CreateDatasetCmd, r io.Reader, size int64) (uint64, error)
	AddDatasetVersion(ctx context.Context, cmd AddVersionCmd, r io.Reader, size int64) (uint64, error)
	GetDataset(ctx context.Context, id uint64) (*entities.Dataset, error)
	ListDatasets(ctx context.Context, onlyPublic bool, ownerID *uint64) ([]*entities.Dataset, error)
	GetVersion(ctx context.Context, versionID uint64) (*entities.DatasetVersion, error)
	ListVersions(ctx context.Context, datasetID uint64) ([]*entities.DatasetVersion, error)
	ListByCategory(ctx context.Context, categoryID uint64) ([]*entities.Dataset, error)
	DeleteDataset(ctx context.Context, datasetID uint64) error
}
