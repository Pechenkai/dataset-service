package services

import (
	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type reviewService struct {
	reviewRepo repositories.ReviewRepository
}

func NewReviewService(reviewRepo repositories.ReviewRepository) ReviewService {
	return &reviewService{
		reviewRepo: reviewRepo,
	}
}

func (s *reviewService) CreateReview(review *entities.Review) error {
	if review == nil {
		return ErrNilReview
	}
	if review.Rating < 1 || review.Rating > 5 {
		return ErrInvalidRating
	}

	return s.reviewRepo.Create(review)
}

func (s *reviewService) UpdateReview(review *entities.Review) error {
	if review == nil {
		return ErrNilReview
	}
	return s.reviewRepo.Update(review)
}

func (s *reviewService) DeleteReview(id uint64) error {
	return s.reviewRepo.Delete(id)
}

func (s *reviewService) GetReviewByID(id uint64) (*entities.Review, error) {
	return s.reviewRepo.FindByID(id)
}

func (s *reviewService) GetReviewsByDataset(datasetID uint64) ([]*entities.Review, error) {
	return s.reviewRepo.FindByDatasetID(datasetID)
}

func (s *reviewService) GetReviewsByUser(userID uint64) ([]*entities.Review, error) {
	return s.reviewRepo.FindByUserID(userID)
}

func (s *reviewService) ComputeAverageRating(datasetID uint64) (float64, error) {
	reviews, err := s.reviewRepo.FindByDatasetID(datasetID)
	if err != nil {
		return 0, err
	}
	if len(reviews) == 0 {
		return 0, nil
	}

	sum := 0
	for _, r := range reviews {
		sum += int(r.Rating)
	}
	average := float64(sum) / float64(len(reviews))
	return average, nil
}
