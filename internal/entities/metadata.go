package entities

import (
	"strings"
)

type Metadata struct {
	ID               uint64 `json:"id"`
	Format           string `json:"format"`
	Size             uint64 `json:"size"`
	Tags             string `json:"tags"`
	DatasetVersionID uint64 `json:"dataset_version_id"`
}

func NewMetadata(format, tags string, size uint64, versionID uint64) (*Metadata, error) {
	format = strings.TrimSpace(format)
	tags = strings.TrimSpace(tags)

	if format == "" {
		return nil, ErrEmptyFormat
	}
	if size == 0 {
		return nil, ErrZeroSize
	}
	if versionID == 0 {
		return nil, ErrMissingVersionID
	}

	return &Metadata{
		Format:           format,
		Size:             size,
		Tags:             tags,
		DatasetVersionID: versionID,
	}, nil
}
