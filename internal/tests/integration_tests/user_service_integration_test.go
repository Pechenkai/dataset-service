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

func TestUserService_RegisterAuthenticateUpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	fabric := testdata.NewFabric()

	pg := integration.StartPostgres(t)
	repo := postqbuild.NewUserRepo(pg.DB, logger)
	svc := services.NewUserService(repo, logger)

	registerCmd := fabric.RegisterUserCommand()
	userID, err := svc.Register(ctx, registerCmd)
	require.NoError(t, err)
	require.NotZero(t, userID)

	authCmd := fabric.AuthenticateUserCommand(registerCmd.Email, registerCmd.Password)
	authUser, err := svc.Authenticate(ctx, authCmd)
	require.NoError(t, err)
	require.Equal(t, userID, authUser.ID)
	require.Equal(t, registerCmd.Email, authUser.Email)

	stored, err := svc.GetUserByID(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, registerCmd.Email, stored.Email)

	updateCmd := fabric.UpdateUserCommand(stored)
	err = svc.UpdateUser(ctx, updateCmd)
	require.NoError(t, err)

	updated, err := svc.GetUserByID(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, updateCmd.Username, updated.Username)
	require.Equal(t, updateCmd.Country, updated.Country)
	assert.Equal(t, updateCmd.Email, updated.Email)

	_, err = svc.Authenticate(ctx, authCmd)
	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)

	authAfterUpdate := fabric.AuthenticateUserCommand(updateCmd.Email, updateCmd.Password)
	authUpdatedUser, err := svc.Authenticate(ctx, authAfterUpdate)
	require.NoError(t, err)
	require.Equal(t, userID, authUpdatedUser.ID)

	err = svc.DeleteUser(ctx, userID)
	require.NoError(t, err)

	_, err = svc.GetUserByID(ctx, userID)
	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrUserNotFound)
}
