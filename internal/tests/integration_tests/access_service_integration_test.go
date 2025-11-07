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

func TestAccessService_RequestApproveAndDenyFlow(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	fabric := testdata.NewFabric()

	pg := integration.StartPostgres(t)

	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	categoryRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	datasetRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	accessRepo := postqbuild.NewAccessRequestRepo(pg.DB, logger)

	service := services.NewAccessService(datasetRepo, accessRepo, logger)

	owner := fabric.RegularUser()
	require.NoError(t, userRepo.Create(ctx, owner))

	requester := fabric.RegularUser()
	require.NoError(t, userRepo.Create(ctx, requester))

	category := fabric.Category()
	require.NoError(t, categoryRepo.Create(ctx, category))

	dataset := fabric.Dataset(owner, category)
	dataset.IsPublic = false
	require.NoError(t, datasetRepo.Create(ctx, dataset))

	requestCmd := fabric.RequestAccessCommand(dataset.ID, requester.ID)
	err := service.Request(ctx, requestCmd)
	require.NoError(t, err)

	pending, err := service.ListPending(ctx, owner.ID)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Equal(t, dataset.ID, pending[0].DatasetID)
	assert.Equal(t, requester.ID, pending[0].UserID)

	found, err := service.Find(ctx, dataset.ID, requester.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, entities.AccessStatusPending, found.Status)

	err = service.Request(ctx, requestCmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrRequestAlreadyExists)

	err = service.Approve(ctx, pending[0].ID, owner.ID)
	require.NoError(t, err)

	approved, err := service.FindByRequestID(ctx, pending[0].ID)
	require.NoError(t, err)
	require.Equal(t, entities.AccessStatusApproved, approved.Status)

	err = service.Request(ctx, requestCmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrRequestAlreadyExists)

	anotherUser := fabric.AdminUser()
	require.NoError(t, userRepo.Create(ctx, anotherUser))

	secondCmd := fabric.RequestAccessCommand(dataset.ID, anotherUser.ID)
	err = service.Request(ctx, secondCmd)
	require.NoError(t, err)

	secondPending, err := service.ListPending(ctx, owner.ID)
	require.NoError(t, err)
	require.Len(t, secondPending, 1)

	err = service.Deny(ctx, secondPending[0].ID, owner.ID)
	require.NoError(t, err)

	denied, err := service.FindByRequestID(ctx, secondPending[0].ID)
	require.NoError(t, err)
	require.Equal(t, entities.AccessStatusDenied, denied.Status)

	result, err := service.Find(ctx, dataset.ID, anotherUser.ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entities.AccessStatusDenied, result.Status)

	publicDataset := fabric.Dataset(owner, category)
	publicDataset.IsPublic = true
	require.NoError(t, datasetRepo.Create(ctx, publicDataset))

	publicRequest := fabric.RequestAccessCommand(publicDataset.ID, requester.ID)
	err = service.Request(ctx, publicRequest)
	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrBadRequest)
}
