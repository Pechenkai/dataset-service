package services_test

import (
	"context"
	"errors"
	"ppo/internal/entities"
	"ppo/internal/repositories"
	"ppo/internal/services"
	"ppo/internal/tests/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateReview_Success(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	cmd := services.CreateReviewCmd{
		UserID:    1,
		DatasetID: 2,
		Rating:    entities.Rating4,
		Text:      "Great!",
	}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(r *entities.Review) bool {
		return r.UserID == 1 && r.DatasetID == 2 && r.Rating == entities.Rating4 &&
			r.Text == "Great!"
	})).Run(func(args mock.Arguments) {
		args.Get(1).(*entities.Review).ID = 42
	}).Return(nil)

	id, err := svc.CreateReview(context.Background(), cmd)
	assert.NoError(t, err)
	assert.Equal(t, uint64(42), id)
	repo.AssertExpectations(t)
}

func TestCreateReview_InvalidRating(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	_, err := svc.CreateReview(context.Background(), services.CreateReviewCmd{
		UserID:    1,
		DatasetID: 1,
		Rating:    entities.Rating(10),
		Text:      "X",
	})
	assert.ErrorIs(t, err, services.ErrInvalidRating)
}

func TestUpdateReview_Success(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	existing := &entities.Review{ID: 5, UserID: 1, DatasetID: 2, Rating: entities.Rating3}
	repo.On("FindByID", mock.Anything, uint64(5)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(r *entities.Review) bool {
		return r.ID == 5 && r.Rating == entities.Rating5 && r.Text == "Upd"
	})).Return(nil)

	err := svc.UpdateReview(context.Background(), services.UpdateReviewCmd{
		ReviewID: 5,
		Rating:   entities.Rating5,
		Text:     "Upd",
	})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateReview_NotFound(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	repo.On("FindByID", mock.Anything, uint64(9)).Return(nil, nil)

	err := svc.UpdateReview(context.Background(), services.UpdateReviewCmd{
		ReviewID: 9,
		Rating:   entities.Rating3,
		Text:     "X",
	})
	assert.ErrorIs(t, err, services.ErrReviewNotFound)
}

func TestUpdateReview_InvalidRating(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	existing := &entities.Review{ID: 7, UserID: 1, DatasetID: 2, Rating: entities.Rating2}
	repo.On("FindByID", mock.Anything, uint64(7)).Return(existing, nil)

	svc := services.NewReviewService(repo)
	err := svc.UpdateReview(context.Background(), services.UpdateReviewCmd{
		ReviewID: 7,
		Rating:   entities.Rating(9),
		Text:     "X",
	})
	assert.ErrorIs(t, err, services.ErrInvalidRating)
}

func TestDeleteReview_Success(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	repo.On("Delete", mock.Anything, uint64(3)).Return(nil)
	err := svc.DeleteReview(context.Background(), 3)
	assert.NoError(t, err)
}

func TestDeleteReview_NotFound(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	repo.On("Delete", mock.Anything, uint64(4)).Return(repositories.ErrReviewNotFound)
	err := svc.DeleteReview(context.Background(), 4)
	assert.ErrorIs(t, err, services.ErrReviewNotFound)
}

func TestGetReviewByID_Success(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	expected := &entities.Review{ID: 10}
	repo.On("FindByID", mock.Anything, uint64(10)).Return(expected, nil)

	rev, err := svc.GetReviewByID(context.Background(), 10)
	assert.NoError(t, err)
	assert.Equal(t, expected, rev)
}

func TestGetReviewByID_NotFound(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	repo.On("FindByID", mock.Anything, uint64(11)).Return(nil, nil)
	_, err := svc.GetReviewByID(context.Background(), 11)
	assert.ErrorIs(t, err, services.ErrReviewNotFound)
}

func TestListByDataset_Success(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	list := []*entities.Review{{ID: 1}, {ID: 2}}
	repo.On("FindByDatasetID", mock.Anything, uint64(2)).Return(list, nil)

	res, err := svc.ListByDataset(context.Background(), 2)
	assert.NoError(t, err)
	assert.Equal(t, list, res)
}

func TestListByUser_Error(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	repo.On("FindByUserID", mock.Anything, uint64(3)).Return(nil, errors.New("fail"))
	_, err := svc.ListByUser(context.Background(), 3)
	assert.Error(t, err)
}

func TestGetRatingSummary_Empty(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	repo.On("FindByDatasetID", mock.Anything, uint64(5)).Return([]*entities.Review{}, nil)
	sum, err := svc.GetRatingSummary(context.Background(), 5)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, sum.Average)
	assert.Equal(t, 0, sum.Count)
}

func TestGetRatingSummary_Calc(t *testing.T) {
	repo := new(mocks.ReviewRepository)
	svc := services.NewReviewService(repo)

	reviews := []*entities.Review{
		{Rating: entities.Rating1},
		{Rating: entities.Rating5},
		{Rating: entities.Rating3},
	}
	repo.On("FindByDatasetID", mock.Anything, uint64(8)).Return(reviews, nil)

	sum, err := svc.GetRatingSummary(context.Background(), 8)
	assert.NoError(t, err)
	// (1+5+3)/3 = 3.0
	assert.Equal(t, 3.0, sum.Average)
	assert.Equal(t, 3, sum.Count)
}
