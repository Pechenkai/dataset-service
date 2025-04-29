package services

import "ppo/internal/entities"

type ReviewService interface {
	CreateReview(review *entities.Review) error

	UpdateReview(review *entities.Review) error

	DeleteReview(id uint64) error

	GetReviewByID(id uint64) (*entities.Review, error)

	GetReviewsByDataset(datasetID uint64) ([]*entities.Review, error)

	GetReviewsByUser(userID uint64) ([]*entities.Review, error)

	ComputeAverageRating(datasetID uint64) (float64, error)
}
