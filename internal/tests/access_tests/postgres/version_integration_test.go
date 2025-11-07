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

func TestVersionRepository_CRUD(t *testing.T) {
	pg := integration.StartPostgres(t)
	logger := zap.NewNop()

	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	categoryRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	datasetRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	versionRepo := postqbuild.NewVersionRepo(pg.DB, logger)

	user := mustCreateUser(t, userRepo)
	category := mustCreateCategory(t, categoryRepo)
	dataset := mustCreateDataset(t, datasetRepo, user.ID, category.ID, true)

	version := mustCreateVersion(t, versionRepo, dataset.ID, "v0.1")
	require.NotZero(t, version.ID)

	version.ChangeLog = "updated changelog"
	version.Number = "v0.2"
	require.NoError(t, versionRepo.Update(ctx, version))

	found, err := versionRepo.FindByID(ctx, version.ID)
	require.NoError(t, err)
	assert.Equal(t, "v0.2", found.Number)
	assert.Equal(t, "updated changelog", found.ChangeLog)

	list, err := versionRepo.FindByDatasetID(ctx, dataset.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, version.ID, list[0].ID)

	require.NoError(t, versionRepo.Delete(ctx, version.ID))

	_, err = versionRepo.FindByID(ctx, version.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrVersionNotFound)
}
