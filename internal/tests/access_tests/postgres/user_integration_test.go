//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
	"ppo/internal/repositories"
	"ppo/internal/tests/integration"
)

func TestUserRepository_CRUDAndQueries(t *testing.T) {
	pg := integration.StartPostgres(t)
	logger := zap.NewNop()

	repo := postqbuild.NewUserRepo(pg.DB, logger)

	user := mustCreateUser(t, repo)
	require.NotZero(t, user.ID)

	byEmail, err := repo.FindByEmail(ctx, user.Email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, byEmail.ID)

	user.Username = "updated-" + user.Username
	user.Password = "new-password"
	user.Country = "US"
	user.IsBlocked = true
	user.Role = entities.RoleAdmin
	require.NoError(t, repo.Update(ctx, user))

	byID, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "US", byID.Country)
	assert.True(t, byID.IsBlocked)
	assert.Equal(t, entities.RoleAdmin, byID.Role)

	duplicate, err := entities.NewUser(
		"duplicate",
		user.Email,
		"password123",
		"RU",
		entities.RoleUser,
		time.Now().Add(-time.Hour),
	)
	require.NoError(t, err)
	err = repo.Create(ctx, duplicate)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrEmailAlreadyExists)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(all), 1)

	require.NoError(t, repo.Delete(ctx, user.ID))

	_, err = repo.FindByID(ctx, user.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrUserNotFound)
}
