package dto

import (
	"time"

	"ppo/internal/entities"
)

type ReviewDTO struct {
	ID             uint64
	UserID         uint64
	DatasetID      uint64
	Rating         entities.Rating
	AuthorUsername string
	Text           string
	CreatedAt      time.Time
}

func ToReviewDTO(r *entities.Review) *ReviewDTO {
	return &ReviewDTO{
		ID:        r.ID,
		UserID:    r.UserID,
		DatasetID: r.DatasetID,
		Rating:    r.Rating,
		Text:      r.Text,
		CreatedAt: r.CreatedAt,
	}
}

func ToReviewDTOs(list []*entities.Review) []*ReviewDTO {
	res := make([]*ReviewDTO, 0, len(list))
	for _, r := range list {
		res = append(res, ToReviewDTO(r))
	}
	return res
}

type CreateReviewForm struct {
	DatasetID uint64
	Rating    int
	Text      string
}

type UpdateReviewForm struct {
	ID     uint64
	Rating int
	Text   string
}
type RatingSummaryDTO struct {
	Average float64
	Count   int
}
