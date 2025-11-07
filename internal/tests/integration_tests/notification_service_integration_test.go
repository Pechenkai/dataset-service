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

func TestNotificationService_NotifySubscribersAndMarkAsRead(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	fabric := testdata.NewFabric()

	pg := integration.StartPostgres(t)

	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	categoryRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	datasetRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	subRepo := postqbuild.NewSubscriptionRepo(pg.DB, logger)
	notifRepo := postqbuild.NewNotificationRepo(pg.DB, logger)

	svc := services.NewNotificationService(notifRepo, subRepo, logger)

	user := fabric.RegularUser()
	require.NoError(t, userRepo.Create(ctx, user))

	category := fabric.Category()
	require.NoError(t, categoryRepo.Create(ctx, category))

	dataset := fabric.Dataset(user, category)
	require.NoError(t, datasetRepo.Create(ctx, dataset))

	subscription := fabric.Subscription(user, dataset)
	require.NoError(t, subRepo.Create(ctx, subscription))

	message := "dataset updated with new samples"
	cmd := fabric.NotifySubscribersCommand(dataset.ID, message)
	count, err := svc.NotifySubscribers(ctx, cmd)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	userNotifications, err := svc.GetNotificationsByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, userNotifications, 1)
	require.Equal(t, message, userNotifications[0].Message)
	assert.False(t, userNotifications[0].IsRead)

	err = svc.MarkAsRead(ctx, userNotifications[0].ID)
	require.NoError(t, err)

	notification, err := notifRepo.FindByID(ctx, userNotifications[0].ID)
	require.NoError(t, err)
	assert.True(t, notification.IsRead)

	err = svc.NotifyUser(ctx, user.ID, dataset.ID, "direct alert")
	require.NoError(t, err)

	all, err := svc.GetNotificationsByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, all, 2)
	assert.Contains(t, []string{all[0].Message, all[1].Message}, "direct alert")
}
