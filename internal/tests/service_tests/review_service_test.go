package services

import (
	"errors"
	"ppo/internal/entities"
	"ppo/internal/services"
	"testing"
)

type mockReviewRepo struct {
	createErr    error
	updateErr    error
	deleteErr    error
	findErr      error
	findByIDResp *entities.Review
	byDataset    []*entities.Review
	byUser       []*entities.Review
}

func (m *mockReviewRepo) Create(r *entities.Review) error { return m.createErr }
func (m *mockReviewRepo) Update(r *entities.Review) error { return m.updateErr }
func (m *mockReviewRepo) Delete(id int64) error           { return m.deleteErr }
func (m *mockReviewRepo) FindByID(id int64) (*entities.Review, error) {
	return m.findByIDResp, m.findErr
}
func (m *mockReviewRepo) FindByDatasetID(id int64) ([]*entities.Review, error) {
	return m.byDataset, m.findErr
}
func (m *mockReviewRepo) FindByUserID(id int64) ([]*entities.Review, error) {
	return m.byUser, m.findErr
}

func TestReviewService_CreateAndRating(t *testing.T) {
	svc := services.NewReviewService(&mockReviewRepo{})

	// Создание: nil-параметр
	err := svc.CreateReview(nil)
	if !errors.Is(err, services.ErrNilReview) {
		t.Error("ожидается ошибка при создании nil-отзыва")
	}

	// Создание: некорректный рейтинг
	err = svc.CreateReview(&entities.Review{Rating: 0})
	if !errors.Is(err, services.ErrInvalidRating) {
		t.Error("ожидается ошибка при некорректном рейтинге")
	}

	// Валидация создания: ошибка репозитория
	repoErr := &mockReviewRepo{createErr: errors.New("сбой создания")}
	svc = services.NewReviewService(repoErr)
	err = svc.CreateReview(&entities.Review{Rating: 3})
	if err == nil || err.Error() != "сбой создания" {
		t.Errorf("ожидается 'сбой создания', получено: %v", err)
	}

	// Вычисление среднего рейтинга: два отзыва
	reviews := []*entities.Review{{Rating: 3}, {Rating: 5}}
	repoAgg := &mockReviewRepo{byDataset: reviews}
	svc = services.NewReviewService(repoAgg)
	avg, _ := svc.ComputeAverageRating(1)
	if avg != 4.0 {
		t.Errorf("ожидается средний рейтинг 4.0, получено %v", avg)
	}

	// Вычисление среднего рейтинга: нет отзывов
	repoEmpty := &mockReviewRepo{byDataset: []*entities.Review{}}
	svc = services.NewReviewService(repoEmpty)
	avg, _ = svc.ComputeAverageRating(1)
	if avg != 0 {
		t.Errorf("ожидается рейтинг 0 для пустого списка, получено %v", avg)
	}
}
