package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type reviewService struct {
	repo repositories.ReviewRepository
}

func NewReviewService(repo repositories.ReviewRepository) ReviewService {
	return &reviewService{repo: repo}
}

func (s *reviewService) CreateReview(ctx context.Context, cmd CreateReviewCmd) (uint64, error) {
	rev, err := entities.NewReview(cmd.UserID, cmd.DatasetID, cmd.Rating /*createdAt=*/, nowUTC(), cmd.Text)
	if err != nil {
		if errors.Is(err, entities.ErrInvalidRating) {
			return 0, ErrInvalidRating
		}
		return 0, fmt.Errorf("invalid review: %w", err)
	}

	if err := s.repo.Create(ctx, rev); err != nil {
		return 0, fmt.Errorf("create review: %w", err)
	}
	return rev.ID, nil
}

func (s *reviewService) UpdateReview(ctx context.Context, cmd UpdateReviewCmd) error {
	rev, err := s.repo.FindByID(ctx, cmd.ReviewID)
	if err != nil {
		return fmt.Errorf("fetch review: %w", err)
	}
	if rev == nil {
		return ErrReviewNotFound
	}
	if cmd.Rating < entities.Rating1 || cmd.Rating > entities.Rating5 {
		return ErrInvalidRating
	}
	rev.Rating = cmd.Rating
	rev.Text = cmd.Text
	if err := s.repo.Update(ctx, rev); err != nil {
		return fmt.Errorf("update review: %w", err)
	}
	return nil
}

func (s *reviewService) DeleteReview(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repositories.ErrReviewNotFound) {
			return ErrReviewNotFound
		}
		return fmt.Errorf("delete review: %w", err)
	}
	return nil
}

func (s *reviewService) GetReviewByID(ctx context.Context, id uint64) (*entities.Review, error) {
	rev, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}
	if rev == nil {
		return nil, ErrReviewNotFound
	}
	return rev, nil
}

func (s *reviewService) ListByDataset(ctx context.Context, datasetID uint64) ([]*entities.Review, error) {
	list, err := s.repo.FindByDatasetID(ctx, datasetID)
	if err != nil {
		return nil, fmt.Errorf("list reviews by dataset: %w", err)
	}
	return list, nil
}

func (s *reviewService) ListByUser(ctx context.Context, userID uint64) ([]*entities.Review, error) {
	list, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list reviews by user: %w", err)
	}
	return list, nil
}

func (s *reviewService) GetRatingSummary(ctx context.Context, datasetID uint64) (RatingSummary, error) {
	reviews, err := s.repo.FindByDatasetID(ctx, datasetID)
	if err != nil {
		return RatingSummary{}, fmt.Errorf("fetch reviews for summary: %w", err)
	}
	if len(reviews) == 0 {
		return RatingSummary{Average: 0, Count: 0}, nil
	}
	sum := 0
	for _, r := range reviews {
		sum += int(r.Rating)
	}
	avg := float64(sum) / float64(len(reviews))
	return RatingSummary{Average: avg, Count: len(reviews)}, nil
}

func nowUTC() (t time.Time) {
	return time.Now().UTC()
}
