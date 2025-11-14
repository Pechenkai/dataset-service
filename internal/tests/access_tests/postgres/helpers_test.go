//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
)

var seq uint64

var ctx = context.Background()

func uniqueSuffix() string {
	id := atomic.AddUint64(&seq, 1)
	return fmt.Sprintf("it-%d", id)
}

func mustCreateUser(t *testing.T, repo *postqbuild.UserRepo) *entities.User {
	t.Helper()

	suffix := uniqueSuffix()
	email := fmt.Sprintf("%s@example.com", suffix)
	user, err := entities.NewUser(
		"user-"+suffix,
		email,
		"password123",
		"RU",
		entities.RoleUser,
		time.Now().Add(-time.Hour),
	)
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, user))
	return user
}

func mustCreateCategory(t *testing.T, repo *postqbuild.CategoryRepo) *entities.Category {
	t.Helper()

	cat, err := entities.NewCategory("category-"+uniqueSuffix(), "integration category")
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, cat))
	return cat
}

func mustCreateDataset(t *testing.T, repo *postqbuild.DatasetRepo, ownerID, categoryID uint64, isPublic bool) *entities.Dataset {
	t.Helper()

	ds, err := entities.NewDataset(
		"dataset-"+uniqueSuffix(),
		"dataset created in integration test",
		ownerID,
		categoryID,
		isPublic,
		time.Now().Add(-2*time.Hour),
	)
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, ds))
	return ds
}

func mustCreateVersion(t *testing.T, repo *postqbuild.VersionRepo, datasetID uint64, number string) *entities.DatasetVersion {
	t.Helper()

	if number == "" {
		number = "v1.0"
	}

	filepath := fmt.Sprintf("datasets/%d/%s.bin", datasetID, uniqueSuffix())
	version, err := entities.NewDatasetVersion(number, filepath, "initial changelog", datasetID, time.Now().Add(-time.Hour))
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, version))
	return version
}

func mustCreateMetadata(t *testing.T, repo *postqbuild.MetadataRepo, versionID uint64) *entities.Metadata {
	t.Helper()

	md, err := entities.NewMetadata("json", "integration,metadata", 1024, versionID)
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, md))
	return md
}

func mustCreateSubscription(t *testing.T, repo *postqbuild.SubscriptionRepo, userID, datasetID uint64) *entities.Subscription {
	t.Helper()

	sub, err := entities.NewSubscription(userID, datasetID, time.Now())
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, sub))
	return sub
}

func mustCreateNotification(t *testing.T, repo *postqbuild.NotificationRepo, userID, datasetID uint64) *entities.Notification {
	t.Helper()

	notif, err := entities.NewNotification(userID, datasetID, "dataset updated", time.Now().Add(-5*time.Minute))
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, notif))
	return notif
}

func mustCreateReview(t *testing.T, repo *postqbuild.ReviewRepo, userID, datasetID uint64) *entities.Review {
	t.Helper()

	review, err := entities.NewReview(userID, datasetID, entities.Rating5, time.Now().Add(-30*time.Minute), "Excellent dataset for tests")
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, review))
	return review
}

func mustCreateAccessRequest(t *testing.T, repo *postqbuild.AccessRequestRepo, datasetID, userID uint64) *entities.AccessRequest {
	t.Helper()

	req, err := entities.NewAccessRequest(datasetID, userID)
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, req))
	return req
}
