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

func TestAccessRequestRepository_Workflow(t *testing.T) {
	pg := integration.StartPostgres(t)
	logger := zap.NewNop()

	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	categoryRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	datasetRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	requestRepo := postqbuild.NewAccessRequestRepo(pg.DB, logger)

	owner := mustCreateUser(t, userRepo)
	requester := mustCreateUser(t, userRepo)
	category := mustCreateCategory(t, categoryRepo)
	dataset := mustCreateDataset(t, datasetRepo, owner.ID, category.ID, true)

	request := mustCreateAccessRequest(t, requestRepo, dataset.ID, requester.ID)
	require.NotZero(t, request.ID)

	found, err := requestRepo.Find(ctx, dataset.ID, requester.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, entities.AccessStatusPending, found.Status)

	pending, err := requestRepo.ListPendingByOwner(ctx, owner.ID)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Equal(t, request.ID, pending[0].ID)

	require.NoError(t, requestRepo.UpdateStatus(ctx, request.ID, string(entities.AccessStatusApproved)))

	byID, err := requestRepo.FindByRequestID(ctx, request.ID)
	require.NoError(t, err)
	assert.Equal(t, entities.AccessStatusApproved, byID.Status)

	pending, err = requestRepo.ListPendingByOwner(ctx, owner.ID)
	require.NoError(t, err)
	assert.Len(t, pending, 0, "request should no longer be pending after approval")

	err = requestRepo.UpdateStatus(ctx, request.ID+999, string(entities.AccessStatusDenied))
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrRequestNotFound)
}
