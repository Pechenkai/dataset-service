//go:build integration

package integrationtests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/services"
	"ppo/internal/tests/integration"
	"ppo/internal/tests/testdata"
)

func TestSubscriptionService_SubscribeAndUnsubscribeFlow(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	fabric := testdata.NewFabric()

	pg := integration.StartPostgres(t)

	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	categoryRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	datasetRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	subRepo := postqbuild.NewSubscriptionRepo(pg.DB, logger)

	svc := services.NewSubscriptionService(subRepo, logger)

	user := fabric.RegularUser()
	require.NoError(t, userRepo.Create(ctx, user))

	category := fabric.Category()
	require.NoError(t, categoryRepo.Create(ctx, category))

	dataset := fabric.Dataset(user, category)
	require.NoError(t, datasetRepo.Create(ctx, dataset))

	err := svc.Subscribe(ctx, user.ID, dataset.ID)
	require.NoError(t, err)

	subscribed, err := svc.IsSubscribed(ctx, user.ID, dataset.ID)
	require.NoError(t, err)
	assert.True(t, subscribed)

	subscribers, err := svc.ListSubscribers(ctx, dataset.ID)
	require.NoError(t, err)
	require.Contains(t, subscribers, user.ID)

	datasets, err := svc.ListSubscriptions(ctx, user.ID)
	require.NoError(t, err)
	require.Contains(t, datasets, dataset.ID)

	err = svc.Unsubscribe(ctx, user.ID, dataset.ID)
	require.NoError(t, err)

	subscribed, err = svc.IsSubscribed(ctx, user.ID, dataset.ID)
	require.NoError(t, err)
	assert.False(t, subscribed)

	err = svc.Unsubscribe(ctx, user.ID, dataset.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrNotSubscribed)
}
