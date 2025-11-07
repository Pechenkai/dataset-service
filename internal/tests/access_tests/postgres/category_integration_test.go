//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
	"ppo/internal/repositories"
	"ppo/internal/tests/integration"
)

func TestCategoryRepository_CreateUpdateAndList(t *testing.T) {
	pg := integration.StartPostgres(t)

	repo := postqbuild.NewCategoryRepo(pg.DB, zap.NewNop())
	ctx := context.Background()

	cat := &entities.Category{
		Name:        "Computer Vision",
		Description: "initial description",
	}

	require.NoError(t, repo.Create(ctx, cat))
	require.NotZero(t, cat.ID)

	cat.Description = "updated description"
	require.NoError(t, repo.Update(ctx, cat))

	found, err := repo.FindByID(ctx, cat.ID)
	require.NoError(t, err)
	assert.Equal(t, "Computer Vision", found.Name)
	assert.Equal(t, "updated description", found.Description)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)
	assert.Equal(t, cat.ID, all[0].ID)
}

func TestCategoryRepository_DuplicateName(t *testing.T) {
	pg := integration.StartPostgres(t)

	repo := postqbuild.NewCategoryRepo(pg.DB, zap.NewNop())
	ctx := context.Background()

	first := &entities.Category{
		Name:        "Natural Language Processing",
		Description: "language models",
	}
	require.NoError(t, repo.Create(ctx, first))

	duplicate := &entities.Category{
		Name:        first.Name,
		Description: "should fail",
	}
	err := repo.Create(ctx, duplicate)
	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrCategoryAlreadyExists)
}
