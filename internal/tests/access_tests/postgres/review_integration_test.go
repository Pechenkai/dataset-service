//go:build integration

package postgres_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
	"ppo/internal/repositories"
	"ppo/internal/tests/integration"
)

func TestReviewRepository_CRUDAndQueries(t *testing.T) {
	pg := integration.StartPostgres(t)
	logger := zap.NewNop()

	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	categoryRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	datasetRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	reviewRepo := postqbuild.NewReviewRepo(pg.DB, logger)

	user := mustCreateUser(t, userRepo)
	category := mustCreateCategory(t, categoryRepo)
	dataset := mustCreateDataset(t, datasetRepo, user.ID, category.ID, true)

	review := mustCreateReview(t, reviewRepo, user.ID, dataset.ID)
	require.NotZero(t, review.ID)

	review.Rating = entities.Rating4
	review.Text = "updated text"
	require.NoError(t, reviewRepo.Update(ctx, review))

	found, err := reviewRepo.FindByID(ctx, review.ID)
	require.NoError(t, err)
	assert.Equal(t, entities.Rating4, found.Rating)
	assert.Equal(t, "updated text", found.Text)

	byDataset, err := reviewRepo.FindByDatasetID(ctx, dataset.ID)
	require.NoError(t, err)
	require.Len(t, byDataset, 1)
	assert.Equal(t, review.ID, byDataset[0].ID)

	byUser, err := reviewRepo.FindByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, byUser, 1)
	assert.Equal(t, review.ID, byUser[0].ID)

	require.NoError(t, reviewRepo.Delete(ctx, review.ID))

	_, err = reviewRepo.FindByID(ctx, review.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrReviewNotFound)
}
