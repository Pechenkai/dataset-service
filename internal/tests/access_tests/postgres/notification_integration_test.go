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

func TestNotificationRepository_CRUD(t *testing.T) {
	pg := integration.StartPostgres(t)
	logger := zap.NewNop()

	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	categoryRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	datasetRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	notificationRepo := postqbuild.NewNotificationRepo(pg.DB, logger)

	user := mustCreateUser(t, userRepo)
	category := mustCreateCategory(t, categoryRepo)
	dataset := mustCreateDataset(t, datasetRepo, user.ID, category.ID, true)

	notification := mustCreateNotification(t, notificationRepo, user.ID, dataset.ID)
	require.NotZero(t, notification.ID)

	notification.Message = "dataset is now public"
	notification.IsRead = true
	require.NoError(t, notificationRepo.Update(ctx, notification))

	found, err := notificationRepo.FindByID(ctx, notification.ID)
	require.NoError(t, err)
	assert.True(t, found.IsRead)
	assert.Equal(t, "dataset is now public", found.Message)

	userNotifications, err := notificationRepo.FindByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, userNotifications, 1)
	assert.Equal(t, notification.ID, userNotifications[0].ID)

	require.NoError(t, notificationRepo.Delete(ctx, notification.ID))

	_, err = notificationRepo.FindByID(ctx, notification.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrNotificationNotFound)
}
