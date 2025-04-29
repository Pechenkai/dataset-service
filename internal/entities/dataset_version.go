package entities

import (
	"strings"
	"time"
)

type DatasetVersion struct {
	ID         uint64    `json:"id"`
	Number     string    `json:"number"`
	UploadDate time.Time `json:"upload_date"`
	Filepath   string    `json:"filepath"`
	DatasetID  uint64    `json:"dataset_id"`
	ChangeLog  string    `json:"change_log"`
}

func NewDatasetVersion(number, filepath, changelog string, datasetID uint64, uploadDate time.Time) (*DatasetVersion, error) {
	number = strings.TrimSpace(number)
	filepath = strings.TrimSpace(filepath)
	changelog = strings.TrimSpace(changelog)

	if number == "" {
		return nil, ErrEmptyVersionNumber
	}
	if len(number) > 20 {
		return nil, ErrVersionNumberTooLong
	}
	if filepath == "" {
		return nil, ErrEmptyFilepath
	}
	if datasetID == 0 {
		return nil, ErrInvalidDatasetID
	}
	if uploadDate.After(time.Now()) {
		return nil, ErrUploadDateInFuture
	}

	return &DatasetVersion{
		Number:     number,
		Filepath:   filepath,
		ChangeLog:  changelog,
		DatasetID:  datasetID,
		UploadDate: uploadDate,
	}, nil
}
