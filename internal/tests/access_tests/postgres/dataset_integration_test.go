//go:build integration

package postgres_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/repositories"
	"ppo/internal/tests/integration"
)

func TestDatasetRepository_CRUDAndQueries(t *testing.T) {
	pg := integration.StartPostgres(t)
	logger := zap.NewNop()

	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	categoryRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	datasetRepo := postqbuild.NewDatasetRepo(pg.DB, logger)

	user := mustCreateUser(t, userRepo)
	category := mustCreateCategory(t, categoryRepo)

	dataset := mustCreateDataset(t, datasetRepo, user.ID, category.ID, true)
	require.NotZero(t, dataset.ID)

	dataset.Description = "updated description"
	dataset.IsPublic = false
	require.NoError(t, datasetRepo.Update(ctx, dataset))

	found, err := datasetRepo.FindByID(ctx, dataset.ID)
	require.NoError(t, err)
	assert.Equal(t, dataset.Description, found.Description)
	assert.False(t, found.IsPublic)

	owned, err := datasetRepo.FindByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, owned, 1)
	assert.Equal(t, dataset.ID, owned[0].ID)

	all, err := datasetRepo.FindAll(ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)

	public, err := datasetRepo.FindPublic(ctx)
	require.NoError(t, err)
	assert.Len(t, public, 0, "dataset no longer public after update")

	categoryDatasets, err := datasetRepo.FindByCategoryID(ctx, category.ID)
	require.NoError(t, err)
	require.Len(t, categoryDatasets, 1)

	require.NoError(t, datasetRepo.Delete(ctx, dataset.ID))

	_, err = datasetRepo.FindByID(ctx, dataset.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrDatasetNotFound)
}
