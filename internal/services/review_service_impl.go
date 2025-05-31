package services

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"time"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type reviewService struct {
	repo   repositories.ReviewRepository
	logger *zap.Logger
}

// NewReviewService создаёт экземпляр ReviewService с привязанным логгером.
func NewReviewService(repo repositories.ReviewRepository, logger *zap.Logger) ReviewService {
	logger.Debug("NewReviewService initialized")
	return &reviewService{
		repo:   repo,
		logger: logger,
	}
}

// CreateReview создаёт новый отзыв.
func (s *reviewService) CreateReview(ctx context.Context, cmd CreateReviewCmd) (uint64, error) {
	s.logger.Debug("CreateReview called",
		zap.Uint64("user_id", cmd.UserID),
		zap.Uint64("dataset_id", cmd.DatasetID),
		zap.Int("rating", int(cmd.Rating)),
	)

	// Проверяем корректность полей через конструктор entity
	rev, err := entities.NewReview(cmd.UserID, cmd.DatasetID, cmd.Rating, nowUTC(), cmd.Text)
	if err != nil {
		if errors.Is(err, entities.ErrInvalidRating) {
			s.logger.Warn("invalid rating provided",
				zap.Uint64("user_id", cmd.UserID),
				zap.Uint64("dataset_id", cmd.DatasetID),
				zap.Int("rating", int(cmd.Rating)),
			)
			return 0, ErrInvalidRating
		}
		s.logger.Error("failed to construct review entity",
			zap.Error(err),
			zap.Uint64("user_id", cmd.UserID),
			zap.Uint64("dataset_id", cmd.DatasetID),
		)
		return 0, fmt.Errorf("invalid review: %w", err)
	}
	s.logger.Debug("review entity constructed",
		zap.Uint64("user_id", rev.UserID),
		zap.Uint64("dataset_id", rev.DatasetID),
		zap.Int("rating", int(rev.Rating)),
	)

	// Пытаемся сохранить
	if err := s.repo.Create(ctx, rev); err != nil {
		s.logger.Error("failed to create review in repository",
			zap.Error(err),
			zap.Uint64("user_id", rev.UserID),
			zap.Uint64("dataset_id", rev.DatasetID),
		)
		return 0, fmt.Errorf("create review: %w", err)
	}

	s.logger.Info("review created successfully",
		zap.Uint64("review_id", rev.ID),
		zap.Uint64("user_id", rev.UserID),
		zap.Uint64("dataset_id", rev.DatasetID),
		zap.Int("rating", int(rev.Rating)),
	)
	return rev.ID, nil
}

// UpdateReview обновляет существующий отзыв.
func (s *reviewService) UpdateReview(ctx context.Context, cmd UpdateReviewCmd) error {
	s.logger.Debug("UpdateReview called",
		zap.Uint64("review_id", cmd.ReviewID),
		zap.Int("new_rating", int(cmd.Rating)),
	)

	// Находим существующий отзыв
	rev, err := s.repo.FindByID(ctx, cmd.ReviewID)
	if err != nil {
		if errors.Is(err, repositories.ErrReviewNotFound) {
			s.logger.Warn("review not found during fetch",
				zap.Uint64("review_id", cmd.ReviewID),
			)
			return ErrReviewNotFound
		}
		s.logger.Error("failed to fetch review",
			zap.Error(err),
			zap.Uint64("review_id", cmd.ReviewID),
		)
		return fmt.Errorf("fetch review: %w", err)
	}
	if rev == nil {
		s.logger.Warn("review is nil after fetch",
			zap.Uint64("review_id", cmd.ReviewID),
		)
		return ErrReviewNotFound
	}
	s.logger.Debug("fetched review for update",
		zap.Uint64("review_id", rev.ID),
		zap.Int("old_rating", int(rev.Rating)),
	)

	// Проверяем валидность нового рейтинга
	if cmd.Rating < entities.Rating1 || cmd.Rating > entities.Rating5 {
		s.logger.Warn("invalid rating value for update",
			zap.Uint64("review_id", cmd.ReviewID),
			zap.Int("rating", int(cmd.Rating)),
		)
		return ErrInvalidRating
	}

	// Обновляем поля
	rev.Rating = cmd.Rating
	rev.Text = cmd.Text

	// Сохраняем изменения
	if err := s.repo.Update(ctx, rev); err != nil {
		if errors.Is(err, repositories.ErrReviewNotFound) {
			s.logger.Warn("review not found during update",
				zap.Uint64("review_id", cmd.ReviewID),
			)
			return ErrReviewNotFound
		}
		s.logger.Error("failed to update review in repository",
			zap.Error(err),
			zap.Uint64("review_id", rev.ID),
		)
		return fmt.Errorf("update review: %w", err)
	}

	s.logger.Info("review updated successfully",
		zap.Uint64("review_id", rev.ID),
		zap.Int("new_rating", int(rev.Rating)),
	)
	return nil
}

// DeleteReview удаляет отзыв по ID.
func (s *reviewService) DeleteReview(ctx context.Context, id uint64) error {
	s.logger.Debug("DeleteReview called",
		zap.Uint64("review_id", id),
	)

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repositories.ErrReviewNotFound) {
			s.logger.Warn("review not found during delete",
				zap.Uint64("review_id", id),
			)
			return ErrReviewNotFound
		}
		s.logger.Error("failed to delete review from repository",
			zap.Error(err),
			zap.Uint64("review_id", id),
		)
		return fmt.Errorf("delete review: %w", err)
	}

	s.logger.Info("review deleted successfully",
		zap.Uint64("review_id", id),
	)
	return nil
}

// GetReviewByID возвращает отзыв по его ID.
func (s *reviewService) GetReviewByID(ctx context.Context, id uint64) (*entities.Review, error) {
	s.logger.Debug("GetReviewByID called",
		zap.Uint64("review_id", id),
	)

	rev, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrReviewNotFound) {
			s.logger.Warn("review not found during fetch",
				zap.Uint64("review_id", id),
			)
			return nil, ErrReviewNotFound
		}
		s.logger.Error("failed to fetch review",
			zap.Error(err),
			zap.Uint64("review_id", id),
		)
		return nil, fmt.Errorf("get review: %w", err)
	}
	if rev == nil {
		s.logger.Warn("review is nil after fetch",
			zap.Uint64("review_id", id),
		)
		return nil, ErrReviewNotFound
	}

	s.logger.Info("review fetched successfully",
		zap.Uint64("review_id", rev.ID),
		zap.Uint64("user_id", rev.UserID),
		zap.Uint64("dataset_id", rev.DatasetID),
		zap.Int("rating", int(rev.Rating)),
	)
	return rev, nil
}

// ListByDataset возвращает все отзывы для указанного набора данных.
func (s *reviewService) ListByDataset(ctx context.Context, datasetID uint64) ([]*entities.Review, error) {
	s.logger.Debug("ListByDataset called",
		zap.Uint64("dataset_id", datasetID),
	)

	list, err := s.repo.FindByDatasetID(ctx, datasetID)
	if err != nil {
		s.logger.Error("failed to list reviews by dataset",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return nil, fmt.Errorf("list reviews by dataset: %w", err)
	}

	s.logger.Info("reviews list fetched for dataset",
		zap.Uint64("dataset_id", datasetID),
		zap.Int("count", len(list)),
	)
	return list, nil
}

// ListByUser возвращает все отзывы, оставленные конкретным пользователем.
func (s *reviewService) ListByUser(ctx context.Context, userID uint64) ([]*entities.Review, error) {
	s.logger.Debug("ListByUser called",
		zap.Uint64("user_id", userID),
	)

	list, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to list reviews by user",
			zap.Error(err),
			zap.Uint64("user_id", userID),
		)
		return nil, fmt.Errorf("list reviews by user: %w", err)
	}

	s.logger.Info("reviews list fetched for user",
		zap.Uint64("user_id", userID),
		zap.Int("count", len(list)),
	)
	return list, nil
}

// GetRatingSummary вычисляет среднюю оценку и количество отзывов для набора данных.
func (s *reviewService) GetRatingSummary(ctx context.Context, datasetID uint64) (RatingSummary, error) {
	s.logger.Debug("GetRatingSummary called",
		zap.Uint64("dataset_id", datasetID),
	)

	reviews, err := s.repo.FindByDatasetID(ctx, datasetID)
	if err != nil {
		s.logger.Error("failed to fetch reviews for summary",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return RatingSummary{}, fmt.Errorf("fetch reviews for summary: %w", err)
	}
	if len(reviews) == 0 {
		s.logger.Info("no reviews found for rating summary",
			zap.Uint64("dataset_id", datasetID),
		)
		return RatingSummary{Average: 0, Count: 0}, nil
	}

	sum := 0
	for _, r := range reviews {
		sum += int(r.Rating)
	}
	avg := float64(sum) / float64(len(reviews))

	s.logger.Info("rating summary calculated",
		zap.Uint64("dataset_id", datasetID),
		zap.Float64("average_rating", avg),
		zap.Int("review_count", len(reviews)),
	)
	return RatingSummary{Average: avg, Count: len(reviews)}, nil
}

// nowUTC возвращает текущее время в UTC.
func nowUTC() time.Time {
	return time.Now().UTC()
}
