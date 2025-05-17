package services

import (
	"context"

	"ppo/internal/entities"
)

type CreateReviewCmd struct {
	UserID    uint64
	DatasetID uint64
	Rating    entities.Rating
	Text      string
}

type UpdateReviewCmd struct {
	ReviewID uint64
	Rating   entities.Rating
	Text     string
}

type RatingSummary struct {
	Average float64
	Count   int
}

type ReviewService interface {
	CreateReview(ctx context.Context, cmd CreateReviewCmd) (uint64, error)

	UpdateReview(ctx context.Context, cmd UpdateReviewCmd) error

	DeleteReview(ctx context.Context, id uint64) error

	GetReviewByID(ctx context.Context, id uint64) (*entities.Review, error)

	ListByDataset(ctx context.Context, datasetID uint64) ([]*entities.Review, error)

	ListByUser(ctx context.Context, userID uint64) ([]*entities.Review, error)

	GetRatingSummary(ctx context.Context, datasetID uint64) (RatingSummary, error)
}
