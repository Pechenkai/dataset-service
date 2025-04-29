package entities

import (
	"strings"
	"time"
)

type Rating int8

type Review struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	DatasetID uint64    `json:"dataset_id"`
	Rating    Rating    `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	Text      string    `json:"text"`
}

const (
	Rating1 Rating = 1
	Rating2 Rating = 2
	Rating3 Rating = 3
	Rating4 Rating = 4
	Rating5 Rating = 5
)

var validRatings = map[Rating]struct{}{
	Rating1: {},
	Rating2: {},
	Rating3: {},
	Rating4: {},
	Rating5: {},
}

func NewReview(userID, datasetID uint64, rating Rating, createdAt time.Time, text string) (*Review, error) {
	if userID == 0 {
		return nil, ErrMissingUserID
	}
	if datasetID == 0 {
		return nil, ErrMissingDatasetID
	}
	if _, ok := validRatings[rating]; !ok {
		return nil, ErrInvalidRating
	}
	if createdAt.After(time.Now()) {
		return nil, ErrInvalidCreatedAt
	}
	return &Review{
		UserID:    userID,
		DatasetID: datasetID,
		Rating:    rating,
		CreatedAt: createdAt,
		Text:      strings.TrimSpace(text),
	}, nil
}
