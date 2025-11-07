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

func TestMetadataRepository_CRUD(t *testing.T) {
	pg := integration.StartPostgres(t)
	logger := zap.NewNop()

	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	categoryRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	datasetRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	versionRepo := postqbuild.NewVersionRepo(pg.DB, logger)
	metadataRepo := postqbuild.NewMetadataRepo(pg.DB, logger)

	user := mustCreateUser(t, userRepo)
	category := mustCreateCategory(t, categoryRepo)
	dataset := mustCreateDataset(t, datasetRepo, user.ID, category.ID, true)
	version := mustCreateVersion(t, versionRepo, dataset.ID, "v0.1")

	metadata := mustCreateMetadata(t, metadataRepo, version.ID)
	require.NotZero(t, metadata.ID)

	metadata.Format = "parquet"
	metadata.Size = 2048
	metadata.Tags = "updated,tags"
	require.NoError(t, metadataRepo.Update(ctx, metadata))

	found, err := metadataRepo.FindByID(ctx, metadata.ID)
	require.NoError(t, err)
	assert.Equal(t, "parquet", found.Format)
	assert.EqualValues(t, 2048, found.Size)
	assert.Equal(t, "updated,tags", found.Tags)

	require.NoError(t, metadataRepo.Delete(ctx, metadata.ID))

	_, err = metadataRepo.FindByID(ctx, metadata.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrMetadataNotFound)
}
