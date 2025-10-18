//go:build integration

package integrationtests

import (
	"context"
	"io"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
	"ppo/internal/services"
	"ppo/internal/storage"
	"ppo/internal/tests/integration"
	"ppo/internal/tests/testdata"
)

func TestDatasetService_CreateDatasetAndAddVersion(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	pg := integration.StartPostgres(t)
	minioContainer := integration.StartMinio(t, "datasets-integration")

	catRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	dsRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	verRepo := postqbuild.NewVersionRepo(pg.DB, logger)
	mdRepo := postqbuild.NewMetadataRepo(pg.DB, logger)
	userRepo := postqbuild.NewUserRepo(pg.DB, logger)

	stor, err := storage.NewS3Storage(minioContainer.Config)
	require.NoError(t, err)

	svc := services.NewDatasetService(dsRepo, verRepo, mdRepo, stor, logger)

	fabric := testdata.NewFabric()

	user := fabric.RegularUser()
	require.NoError(t, userRepo.Create(ctx, user))

	category := fabric.Category()
	require.NoError(t, catRepo.Create(ctx, category))

	createCmd := fabric.CreateDatasetCommand(category, user)
	const initialDatasetPayload = "initial dataset payload"
	reader, size := fabric.DatasetFile(initialDatasetPayload)

	datasetID, err := svc.CreateDataset(ctx, createCmd, reader, size)
	require.NoError(t, err)
	assert.NotZero(t, datasetID)

	dataset, err := dsRepo.FindByID(ctx, datasetID)
	require.NoError(t, err)
	require.NotNil(t, dataset)
	assert.Equal(t, createCmd.Name, dataset.Name)
	assert.Equal(t, user.ID, dataset.OwnerID)
	assert.Equal(t, category.ID, dataset.CategoryID)

	versions, err := verRepo.FindByDatasetID(ctx, datasetID)
	require.NoError(t, err)
	require.Len(t, versions, 1)

	initialVersion := versions[0]
	assert.Equal(t, "v0.1", initialVersion.Number)
	assert.Equal(t, datasetID, initialVersion.DatasetID)

	obj, err := minioContainer.Client.GetObject(ctx, minioContainer.Config.Bucket, initialVersion.Filepath, minio.GetObjectOptions{})
	require.NoError(t, err)
	initialPayload, err := io.ReadAll(obj)
	require.NoError(t, err)
	assert.Equal(t, initialDatasetPayload, string(initialPayload))
	require.NoError(t, obj.Close())

	var (
		metaFormat string
		metaTags   string
		metaSize   int64
	)
	err = pg.DB.QueryRow(ctx, "SELECT format, tags, size FROM metadata WHERE dataset_version_id = $1", initialVersion.ID).
		Scan(&metaFormat, &metaTags, &metaSize)
	require.NoError(t, err)
	assert.Equal(t, createCmd.MetaFormat, metaFormat)
	assert.Equal(t, createCmd.MetaTags, metaTags)
	assert.EqualValues(t, createCmd.MetaSize, metaSize)

	addCmd := fabric.AddVersionCommand(dataset, user)
	const newVersionPayload = "new version payload"
	verReader, verSize := fabric.DatasetVersionFile(newVersionPayload)

	versionID, err := svc.AddDatasetVersion(ctx, addCmd, verReader, verSize)
	require.NoError(t, err)
	assert.NotZero(t, versionID)

	allVersions, err := verRepo.FindByDatasetID(ctx, datasetID)
	require.NoError(t, err)
	assert.Len(t, allVersions, 2)

	var newVersion *entities.DatasetVersion
	for _, v := range allVersions {
		if v.ID == versionID {
			newVersion = v
			break
		}
	}
	require.NotNil(t, newVersion)
	assert.Equal(t, "v0.2", newVersion.Number)

	obj2, err := minioContainer.Client.GetObject(ctx, minioContainer.Config.Bucket, newVersion.Filepath, minio.GetObjectOptions{})
	require.NoError(t, err)
	verPayload, err := io.ReadAll(obj2)
	require.NoError(t, err)
	assert.Equal(t, newVersionPayload, string(verPayload))
	require.NoError(t, obj2.Close())

	err = pg.DB.QueryRow(ctx, "SELECT format, tags, size FROM metadata WHERE dataset_version_id = $1", newVersion.ID).
		Scan(&metaFormat, &metaTags, &metaSize)
	require.NoError(t, err)
	assert.Equal(t, addCmd.MetaFormat, metaFormat)
	assert.Equal(t, addCmd.MetaTags, metaTags)
	assert.EqualValues(t, addCmd.MetaSize, metaSize)
}
