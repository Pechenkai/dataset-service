//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	delivhttp "ppo/internal/delivery/http"
	"ppo/internal/delivery/http/dto"
	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
	"ppo/internal/services"
	"ppo/internal/storage"
	"ppo/internal/tests/integration"
	"ppo/internal/tests/testdata"
)

func TestE2E_PublicDatasetJourney(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	pg := integration.StartPostgres(t)
	minioContainer := integration.StartMinio(t, "e2e-datasets")

	catRepo := postqbuild.NewCategoryRepo(pg.DB, logger)
	dsRepo := postqbuild.NewDatasetRepo(pg.DB, logger)
	verRepo := postqbuild.NewVersionRepo(pg.DB, logger)
	mdRepo := postqbuild.NewMetadataRepo(pg.DB, logger)
	userRepo := postqbuild.NewUserRepo(pg.DB, logger)
	subRepo := postqbuild.NewSubscriptionRepo(pg.DB, logger)
	notifRepo := postqbuild.NewNotificationRepo(pg.DB, logger)
	revRepo := postqbuild.NewReviewRepo(pg.DB, logger)

	stor, err := storage.NewS3Storage(minioContainer.Config)
	require.NoError(t, err)

	catSvc := services.NewCategoryService(catRepo, logger)
	dsSvc := services.NewDatasetService(dsRepo, verRepo, mdRepo, stor, logger)
	notifSvc := services.NewNotificationService(notifRepo, subRepo, logger)
	revSvc := services.NewReviewService(revRepo, logger)
	userSvc := services.NewUserService(userRepo, logger)
	subSvc := services.NewSubscriptionService(subRepo, logger)

	router := delivhttp.NewRouter(catSvc, dsSvc, notifSvc, revSvc, userSvc, subSvc, logger)
	server := httptest.NewServer(router)
	defer server.Close()

	fabric := testdata.NewFabric()

	user := fabric.RegularUser()
	require.NoError(t, userRepo.Create(ctx, user))

	category := fabric.Category()
	require.NoError(t, catRepo.Create(ctx, category))

	createCmd := fabric.CreateDatasetCommand(category, user)
	reader, size := fabric.DatasetFile("e2e initial payload")
	datasetID, err := dsSvc.CreateDataset(ctx, createCmd, reader, size)
	require.NoError(t, err)

	client := server.Client()

	status, categories := httpGetJSON[[]dto.CategoryResponse](t, client, fmt.Sprintf("%s/categories", server.URL))
	require.Equal(t, http.StatusOK, status)
	require.Len(t, categories, 1)
	assert.Equal(t, category.ID, categories[0].ID)

	status, datasets := httpGetJSON[dto.DatasetsResponse](t, client, fmt.Sprintf("%s/datasets?public=true", server.URL))
	require.Equal(t, http.StatusOK, status)
	require.Len(t, datasets.Datasets, 1)
	assert.Equal(t, datasetID, datasets.Datasets[0].ID)

	status, versions := httpGetJSON[dto.VersionsResponse](t, client, fmt.Sprintf("%s/datasets/%d/versions", server.URL, datasetID))
	require.Equal(t, http.StatusOK, status)
	require.Len(t, versions.Versions, 1)
	assert.Equal(t, "v0.1", versions.Versions[0].Number)

	subReq := dto.SubscribeRequest{
		UserID:    user.ID,
		DatasetID: datasetID,
	}
	status = httpPostJSONExpectNoContent(t, client, fmt.Sprintf("%s/subscriptions", server.URL), subReq)
	require.Equal(t, http.StatusNoContent, status)

	reviewReq := dto.CreateReviewRequest{
		UserID:    user.ID,
		DatasetID: datasetID,
		Rating:    entities.Rating(5),
		Text:      "Great dataset for prototyping models",
	}
	status, review := httpPostJSON[dto.ReviewResponse](t, client, fmt.Sprintf("%s/reviews", server.URL), reviewReq)
	require.Equal(t, http.StatusCreated, status)
	assert.Equal(t, reviewReq.Rating, review.Rating)
	assert.Equal(t, reviewReq.Text, review.Text)

	status, reviews := httpGetJSON[dto.ReviewsResponse](t, client, fmt.Sprintf("%s/datasets/%d/reviews", server.URL, datasetID))
	require.Equal(t, http.StatusOK, status)
	require.Len(t, reviews.Reviews, 1)
	assert.Equal(t, review.ID, reviews.Reviews[0].ID)

	status, summary := httpGetJSON[dto.RatingSummaryResponse](t, client, fmt.Sprintf("%s/datasets/%d/reviews/summary", server.URL, datasetID))
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, 1, summary.Count)
	assert.Equal(t, 5.0, summary.Average)
}

func httpGetJSON[T any](t *testing.T, client *http.Client, url string) (int, T) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var out T
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		if len(body) > 0 {
			require.NoError(t, json.Unmarshal(body, &out))
		}
	}
	return resp.StatusCode, out
}

func httpPostJSON[T any](t *testing.T, client *http.Client, url string, payload any) (int, T) {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var out T
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && resp.ContentLength != 0 {
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		if len(body) > 0 {
			require.NoError(t, json.Unmarshal(body, &out))
		}
	}

	return resp.StatusCode, out
}

func httpPostJSONExpectNoContent(t *testing.T, client *http.Client, url string, payload any) int {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)

	return resp.StatusCode
}
