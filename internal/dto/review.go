package dto

import (
	"ppo/internal/entities"
	"ppo/internal/services"
)

type CreateReviewRequest struct {
	DatasetID uint64 `json:"dataset_id" validate:"required"`
	Rating    int8   `json:"rating" validate:"required,min=1,max=5"`
	Text      string `json:"text"`
}

func (r *CreateReviewRequest) ToCommand(userID uint64) services.CreateReviewCmd {
	return services.CreateReviewCmd{
		UserID:    userID,
		DatasetID: r.DatasetID,
		Rating:    entities.Rating(r.Rating),
		Text:      r.Text,
	}
}

type ReviewResponse struct {
	ID        uint64 `json:"id"`
	UserID    uint64 `json:"user_id"`
	DatasetID uint64 `json:"dataset_id"`
	Rating    int8   `json:"rating"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type ListReviewsResponse struct {
	Reviews []ReviewResponse `json:"reviews"`
}

type RatingSummaryResponse struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}
