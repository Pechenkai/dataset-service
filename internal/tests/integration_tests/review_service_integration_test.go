//go:build integration

package integrationtests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
	"ppo/internal/services"
	"ppo/internal/tests/integration"
	"ppo/internal/tests/testdata"
)

func TestReviewService_CreateUpdateListAndDelete(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	fabric := testdata.NewFabric()

	pg := integration.StartPostgres(t)

	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	categoryRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	datasetRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	reviewRepo := postqbuild.NewReviewRepo(pg.DB, logger)

	svc := services.NewReviewService(reviewRepo, logger)

	owner := fabric.RegularUser()
	require.NoError(t, userRepo.Create(ctx, owner))

	reviewer := fabric.AdminUser()
	require.NoError(t, userRepo.Create(ctx, reviewer))

	secondReviewer := fabric.RegularUser()
	require.NoError(t, userRepo.Create(ctx, secondReviewer))

	category := fabric.Category()
	require.NoError(t, categoryRepo.Create(ctx, category))

	dataset := fabric.Dataset(owner, category)
	require.NoError(t, datasetRepo.Create(ctx, dataset))

	createCmd := fabric.CreateReviewCommand(reviewer, dataset, entities.Rating5)
	reviewID, err := svc.CreateReview(ctx, createCmd)
	require.NoError(t, err)
	require.NotZero(t, reviewID)

	review, err := svc.GetReviewByID(ctx, reviewID)
	require.NoError(t, err)
	require.Equal(t, createCmd.Rating, review.Rating)
	require.Equal(t, createCmd.Text, review.Text)

	updateCmd := fabric.UpdateReviewCommand(review, entities.Rating4)
	err = svc.UpdateReview(ctx, updateCmd)
	require.NoError(t, err)

	updated, err := svc.GetReviewByID(ctx, reviewID)
	require.NoError(t, err)
	assert.Equal(t, entities.Rating4, updated.Rating)
	assert.Equal(t, updateCmd.Text, updated.Text)

	secondCmd := fabric.CreateReviewCommand(secondReviewer, dataset, entities.Rating2)
	secondID, err := svc.CreateReview(ctx, secondCmd)
	require.NoError(t, err)
	require.NotZero(t, secondID)

	byDataset, err := svc.ListByDataset(ctx, dataset.ID)
	require.NoError(t, err)
	require.Len(t, byDataset, 2)

	byUser, err := svc.ListByUser(ctx, reviewer.ID)
	require.NoError(t, err)
	require.Len(t, byUser, 1)
	assert.Equal(t, reviewID, byUser[0].ID)

	summary, err := svc.GetRatingSummary(ctx, dataset.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, summary.Count)
	assert.InDelta(t, float64(3), summary.Average, 0.01)

	err = svc.DeleteReview(ctx, reviewID)
	require.NoError(t, err)

	_, err = svc.GetReviewByID(ctx, reviewID)
	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrReviewNotFound)
}
