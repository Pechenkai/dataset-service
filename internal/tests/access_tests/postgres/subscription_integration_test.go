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

func TestSubscriptionRepository_Flow(t *testing.T) {
	pg := integration.StartPostgres(t)
	logger := zap.NewNop()

	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	categoryRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	datasetRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	subscriptionRepo := postqbuild.NewSubscriptionRepo(pg.DB, logger)

	user := mustCreateUser(t, userRepo)
	category := mustCreateCategory(t, categoryRepo)
	dataset := mustCreateDataset(t, datasetRepo, user.ID, category.ID, true)

	sub := mustCreateSubscription(t, subscriptionRepo, user.ID, dataset.ID)
	require.NotNil(t, sub)

	subscribed, err := subscriptionRepo.IsSubscribed(ctx, user.ID, dataset.ID)
	require.NoError(t, err)
	assert.True(t, subscribed)

	otherDataset := mustCreateDataset(t, datasetRepo, user.ID, category.ID, true)
	subscribed, err = subscriptionRepo.IsSubscribed(ctx, user.ID, otherDataset.ID)
	require.NoError(t, err)
	assert.False(t, subscribed)

	subscribers, err := subscriptionRepo.GetSubscribers(ctx, dataset.ID)
	require.NoError(t, err)
	assert.Equal(t, []uint64{user.ID}, subscribers)

	userSubs, err := subscriptionRepo.GetByUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Contains(t, userSubs, dataset.ID)

	err = subscriptionRepo.Create(ctx, sub)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrAlreadySubscribed)

	require.NoError(t, subscriptionRepo.Unsubscribe(ctx, user.ID, dataset.ID))

	subscribed, err = subscriptionRepo.IsSubscribed(ctx, user.ID, dataset.ID)
	require.NoError(t, err)
	assert.False(t, subscribed)

	err = subscriptionRepo.Unsubscribe(ctx, user.ID, dataset.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrSubscriptionNotFound)
}
