package dto

import (
	"errors"
	"time"

	"ppo/internal/entities"
	"ppo/internal/services"
)

type CreateReviewRequest struct {
	UserID    uint64          `json:"user_id" validate:"required"`
	DatasetID uint64          `json:"dataset_id" validate:"required"`
	Rating    entities.Rating `json:"rating" validate:"required,oneof=1 2 3 4 5"`
	Text      string          `json:"text" validate:"max=1000"`
}

func (r *CreateReviewRequest) ToCommand() services.CreateReviewCmd {
	return services.CreateReviewCmd{
		UserID:    r.UserID,
		DatasetID: r.DatasetID,
		Rating:    r.Rating,
		Text:      r.Text,
	}
}

type UpdateReviewRequest struct {
	Rating entities.Rating `json:"rating" validate:"required,oneof=1 2 3 4 5"`
	Text   string          `json:"text" validate:"max=1000"`
}

func (r *UpdateReviewRequest) ToCommand(id uint64) services.UpdateReviewCmd {
	return services.UpdateReviewCmd{
		ReviewID: id,
		Rating:   r.Rating,
		Text:     r.Text,
	}
}

type ReviewResponse struct {
	ID        uint64          `json:"id"`
	UserID    uint64          `json:"user_id"`
	DatasetID uint64          `json:"dataset_id"`
	Rating    entities.Rating `json:"rating"`
	Text      string          `json:"text"`
	CreatedAt time.Time       `json:"created_at"`
}

func FromEntityReview(r *entities.Review) ReviewResponse {
	return ReviewResponse{
		ID:        r.ID,
		UserID:    r.UserID,
		DatasetID: r.DatasetID,
		Rating:    r.Rating,
		Text:      r.Text,
		CreatedAt: r.CreatedAt,
	}
}

type ReviewsResponse struct {
	Reviews []ReviewResponse `json:"reviews"`
}

func FromEntityReviewList(list []*entities.Review) ReviewsResponse {
	out := make([]ReviewResponse, len(list))
	for i, r := range list {
		out[i] = FromEntityReview(r)
	}
	return ReviewsResponse{Reviews: out}
}

type RatingSummaryResponse struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

func FromServiceSummary(s services.RatingSummary) RatingSummaryResponse {
	return RatingSummaryResponse{
		Average: s.Average,
		Count:   s.Count,
	}
}

func errorsIs(err, target error) bool {
	return err != nil && errors.Is(err, target)
}
