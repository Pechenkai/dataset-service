package dto

import (
	"time"

	"ppo/internal/entities"
)

// ReviewDTO — данные одного отзыва, передаваемые в шаблоны.
type ReviewDTO struct {
	ID        uint64
	UserID    uint64
	DatasetID uint64
	Rating    entities.Rating
	Text      string
	CreatedAt time.Time
}

// ToReviewDTO конвертирует *entities.Review → *ReviewDTO.
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

// ToReviewDTOs конвертирует срез *entities.Review → срез *ReviewDTO.
func ToReviewDTOs(list []*entities.Review) []*ReviewDTO {
	res := make([]*ReviewDTO, 0, len(list))
	for _, r := range list {
		res = append(res, ToReviewDTO(r))
	}
	return res
}

// CreateReviewForm — поля HTML‐формы для создания отзыва.
type CreateReviewForm struct {
	DatasetID uint64
	Rating    int // вводим как int, но приведём к entities.Rating
	Text      string
}

// UpdateReviewForm — поля формы для редактирования существующего отзыва.
type UpdateReviewForm struct {
	ID     uint64
	Rating int
	Text   string
}

// RatingSummaryDTO — передаётся в шаблон для показа средней оценки и количества отзывов.
type RatingSummaryDTO struct {
	Average float64
	Count   int
}
