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

func TestCategoryService_CreateUpdateListAndDelete(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	fabric := testdata.NewFabric()

	pg := integration.StartPostgres(t)
	repo := postqbuild.NewCategoryRepo(pg.DB, logger)
	svc := services.NewCategoryService(repo, logger)

	createCmd := fabric.CreateCategoryCommand()
	categoryID, err := svc.CreateCategory(ctx, createCmd)
	require.NoError(t, err)
	require.NotZero(t, categoryID)

	list, err := svc.ListCategories(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, categoryID, list[0].ID)
	assert.Equal(t, createCmd.Name, list[0].Name)

	stored, err := svc.GetCategoryByID(ctx, categoryID)
	require.NoError(t, err)
	require.NotNil(t, stored)

	updateCmd := fabric.UpdateCategoryCommand(stored)
	err = svc.UpdateCategory(ctx, updateCmd)
	require.NoError(t, err)

	updated, err := svc.GetCategoryByID(ctx, categoryID)
	require.NoError(t, err)
	require.Equal(t, updateCmd.Name, updated.Name)
	require.Equal(t, updateCmd.Description, updated.Description)

	err = svc.DeleteCategory(ctx, categoryID)
	require.NoError(t, err)

	_, err = svc.GetCategoryByID(ctx, categoryID)
	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrCategoryNotFound)
}
