package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ppo/internal/entities"
	"ppo/internal/repositories"
	"ppo/internal/services"
	"ppo/internal/tests/mocks"
	"ppo/internal/tests/testdata"
)

type ReviewServiceSuite struct {
	suite.Suite
}

type reviewMocks struct {
	repo *mocks.ReviewRepository
	svc  services.ReviewService
}

func newReviewMocks(t provider.T) reviewMocks {
	t.Helper()
	repo := &mocks.ReviewRepository{}
	svc := services.NewReviewService(repo, zap.NewNop())
	return reviewMocks{repo: repo, svc: svc}
}

func (m reviewMocks) AssertExpectations(t provider.T) {
	t.Helper()
	m.repo.AssertExpectations(t)
}

func (s *ReviewServiceSuite) TestCreateReview_Success(t provider.T) {
	fabric := testdata.NewFabric()
	user := fabric.RegularUser()
	dataset := fabric.Dataset(user, fabric.Category())
	cmd := fabric.CreateReviewCommand(user, dataset, entities.Rating4)
	m := newReviewMocks(t)

	expectedID := uint64(321)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Review")).Run(func(args mock.Arguments) {
			rev := args.Get(1).(*entities.Review)
			rev.ID = expectedID
		}).Return(nil)
	})

	var (
		id  uint64
		err error
	)

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		id, err = m.svc.CreateReview(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, expectedID, id)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestCreateReview_InvalidRating(t provider.T) {
	fabric := testdata.NewFabric()
	user := fabric.RegularUser()
	dataset := fabric.Dataset(user, fabric.Category())
	cmd := fabric.InvalidCreateReviewCommand(user, dataset)
	m := newReviewMocks(t)

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.CreateReview(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrInvalidRating)
	})
}

func (s *ReviewServiceSuite) TestCreateReview_RepoError(t provider.T) {
	fabric := testdata.NewFabric()
	user := fabric.RegularUser()
	dataset := fabric.Dataset(user, fabric.Category())
	cmd := fabric.CreateReviewCommand(user, dataset, entities.Rating3)
	m := newReviewMocks(t)
	expectedErr := errors.New("insert fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Create", mock.Anything, mock.Anything).Return(expectedErr)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.CreateReview(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestUpdateReview_Success(t provider.T) {
	fabric := testdata.NewFabric()
	review := fabric.ReviewPositive(fabric.RegularUser(), fabric.Dataset(fabric.RegularUser(), fabric.Category()))
	review.ID = 11
	cmd := fabric.UpdateReviewCommand(review, entities.Rating5)
	cmd.ReviewID = review.ID
	m := newReviewMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, review.ID).Return(review, nil)
		m.repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Review")).Run(func(args mock.Arguments) {
			updated := args.Get(1).(*entities.Review)
			assert.Equal(t, cmd.Rating, updated.Rating)
			assert.Equal(t, cmd.Text, updated.Text)
		}).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateReview(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestUpdateReview_NotFound(t provider.T) {
	cmd := services.UpdateReviewCmd{ReviewID: 42, Rating: entities.Rating3}
	m := newReviewMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, cmd.ReviewID).Return((*entities.Review)(nil), repositories.ErrReviewNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateReview(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrReviewNotFound)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestUpdateReview_InvalidRating(t provider.T) {
	review := &entities.Review{ID: 7, Rating: entities.Rating3}
	cmd := services.UpdateReviewCmd{ReviewID: review.ID, Rating: entities.Rating(10)}
	m := newReviewMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, review.ID).Return(review, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateReview(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrInvalidRating)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestUpdateReview_UpdateError(t provider.T) {
	review := &entities.Review{ID: 9, Rating: entities.Rating2}
	cmd := services.UpdateReviewCmd{ReviewID: review.ID, Rating: entities.Rating4, Text: "updated"}
	m := newReviewMocks(t)
	expectedErr := errors.New("update fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, review.ID).Return(review, nil)
		m.repo.On("Update", mock.Anything, mock.Anything).Return(expectedErr)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateReview(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestDeleteReview_Success(t provider.T) {
	m := newReviewMocks(t)
	reviewID := uint64(77)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Delete", mock.Anything, reviewID).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.DeleteReview(context.Background(), reviewID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestDeleteReview_NotFound(t provider.T) {
	m := newReviewMocks(t)
	reviewID := uint64(101)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Delete", mock.Anything, reviewID).Return(repositories.ErrReviewNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.DeleteReview(context.Background(), reviewID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrReviewNotFound)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestGetReviewByID_Success(t provider.T) {
	review := &entities.Review{ID: 55, Rating: entities.Rating5}
	m := newReviewMocks(t)

	var (
		result *entities.Review
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, review.ID).Return(review, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetReviewByID(context.Background(), review.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, review.ID, result.ID)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestGetReviewByID_NotFound(t provider.T) {
	m := newReviewMocks(t)
	reviewID := uint64(404)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, reviewID).Return((*entities.Review)(nil), repositories.ErrReviewNotFound)
	})

	var (
		rev *entities.Review
		err error
	)
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		rev, err = m.svc.GetReviewByID(context.Background(), reviewID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, rev)
		assert.ErrorIs(t, err, services.ErrReviewNotFound)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestGetReviewByID_RepoError(t provider.T) {
	m := newReviewMocks(t)
	expectedErr := errors.New("query fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, uint64(1)).Return((*entities.Review)(nil), expectedErr)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.GetReviewByID(context.Background(), 1)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestListByDataset_Success(t provider.T) {
	review := &entities.Review{ID: 1, DatasetID: 55}
	m := newReviewMocks(t)

	var (
		result []*entities.Review
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByDatasetID", mock.Anything, review.DatasetID).Return([]*entities.Review{review}, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListByDataset(context.Background(), review.DatasetID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		require.Len(t, result, 1)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestListByDataset_Error(t provider.T) {
	m := newReviewMocks(t)
	expectedErr := errors.New("select fail")

	var err error
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByDatasetID", mock.Anything, uint64(1)).Return(nil, expectedErr)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.ListByDataset(context.Background(), 1)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestListByUser_Success(t provider.T) {
	review := &entities.Review{ID: 1, UserID: 99}
	m := newReviewMocks(t)

	var (
		result []*entities.Review
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByUserID", mock.Anything, review.UserID).Return([]*entities.Review{review}, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListByUser(context.Background(), review.UserID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		require.Len(t, result, 1)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestListByUser_Error(t provider.T) {
	m := newReviewMocks(t)
	expectedErr := errors.New("list fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByUserID", mock.Anything, uint64(1)).Return(nil, expectedErr)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.ListByUser(context.Background(), 1)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestGetRatingSummary_WithReviews(t provider.T) {
	reviews := []*entities.Review{{Rating: entities.Rating4}, {Rating: entities.Rating2}}
	m := newReviewMocks(t)

	var (
		summary services.RatingSummary
		err     error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByDatasetID", mock.Anything, uint64(5)).Return(reviews, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		summary, err = m.svc.GetRatingSummary(context.Background(), 5)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, 2, summary.Count)
		assert.InDelta(t, 3.0, summary.Average, 0.01)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestGetRatingSummary_NoReviews(t provider.T) {
	m := newReviewMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByDatasetID", mock.Anything, uint64(6)).Return([]*entities.Review{}, nil)
	})

	var summary services.RatingSummary
	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		summary, err = m.svc.GetRatingSummary(context.Background(), 6)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, 0, summary.Count)
		assert.Equal(t, 0.0, summary.Average)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestGetRatingSummary_Error(t provider.T) {
	m := newReviewMocks(t)
	expectedErr := errors.New("summary fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByDatasetID", mock.Anything, uint64(7)).Return(nil, expectedErr)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.GetRatingSummary(context.Background(), 7)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *ReviewServiceSuite) TestCreateReview_ClassicStyle(t provider.T) {
	repo := &inMemoryReviewRepo{}
	svc := services.NewReviewService(repo, zap.NewNop())
	revCmd := services.CreateReviewCmd{UserID: 1, DatasetID: 2, Rating: entities.Rating5, Text: "great"}

	var (
		id  uint64
		err error
	)

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		id, err = svc.CreateReview(context.Background(), revCmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, uint64(1), id)
		assert.Len(t, repo.reviews, 1)
	})
}

func TestReviewServiceSuite(t *testing.T) {
	suite.RunSuite(t, new(ReviewServiceSuite))
}

type inMemoryReviewRepo struct {
	reviews []*entities.Review
}

func (r *inMemoryReviewRepo) Create(ctx context.Context, rev *entities.Review) error {
	r.reviews = append(r.reviews, rev)
	if rev.ID == 0 {
		rev.ID = uint64(len(r.reviews))
	}
	return nil
}

func (r *inMemoryReviewRepo) Delete(ctx context.Context, id uint64) error {
	for i, rev := range r.reviews {
		if rev.ID == id {
			r.reviews = append(r.reviews[:i], r.reviews[i+1:]...)
			return nil
		}
	}
	return repositories.ErrReviewNotFound
}

func (r *inMemoryReviewRepo) Update(ctx context.Context, rev *entities.Review) error {
	for i, stored := range r.reviews {
		if stored.ID == rev.ID {
			r.reviews[i] = rev
			return nil
		}
	}
	return repositories.ErrReviewNotFound
}

func (r *inMemoryReviewRepo) FindByID(ctx context.Context, id uint64) (*entities.Review, error) {
	for _, rev := range r.reviews {
		if rev.ID == id {
			return rev, nil
		}
	}
	return nil, repositories.ErrReviewNotFound
}

func (r *inMemoryReviewRepo) FindByDatasetID(ctx context.Context, datasetID uint64) ([]*entities.Review, error) {
	var list []*entities.Review
	for _, rev := range r.reviews {
		if rev.DatasetID == datasetID {
			list = append(list, rev)
		}
	}
	return list, nil
}

func (r *inMemoryReviewRepo) FindByUserID(ctx context.Context, userID uint64) ([]*entities.Review, error) {
	var list []*entities.Review
	for _, rev := range r.reviews {
		if rev.UserID == userID {
			list = append(list, rev)
		}
	}
	return list, nil
}
