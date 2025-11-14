//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ppo/internal/config"
	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/delivery/http/dto"
	"ppo/internal/entities"
	"ppo/internal/services"
	"ppo/internal/storage"
	"ppo/internal/tests/testdata"
)

func TestE2E_PublicDatasetJourney(t *testing.T) {
	baseURL := strings.TrimRight(getBaseURL(t), "/")

	seed := seedDataset(t)

	client := &http.Client{Timeout: 10 * time.Second}

	status, categories := httpGetJSON[dto.CategoriesResponse](t, client, fmt.Sprintf("%s/categories", baseURL))
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, categories.Categories)
	assert.Contains(t, extractIDs(categories.Categories), seed.category.ID)

	status, datasets := httpGetJSON[dto.DatasetsResponse](t, client, fmt.Sprintf("%s/datasets?public=true", baseURL))
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, datasets.Datasets)
	assert.Contains(t, extractDatasetIDs(datasets.Datasets), seed.datasetID)

	status, versions := httpGetJSON[dto.VersionsResponse](t, client, fmt.Sprintf("%s/datasets/%d/versions", baseURL, seed.datasetID))
	require.Equal(t, http.StatusOK, status)
	require.Len(t, versions.Versions, 1)
	assert.Equal(t, "v0.1", versions.Versions[0].Number)

	subReq := dto.SubscribeRequest{
		UserID:    seed.user.ID,
		DatasetID: seed.datasetID,
	}
	status = httpPostJSONExpectNoContent(t, client, fmt.Sprintf("%s/subscriptions", baseURL), subReq)
	require.Equal(t, http.StatusNoContent, status)

	reviewReq := dto.CreateReviewRequest{
		UserID:    seed.user.ID,
		DatasetID: seed.datasetID,
		Rating:    entities.Rating(5),
		Text:      "Great dataset for prototyping models",
	}
	status, review := httpPostJSON[dto.ReviewResponse](t, client, fmt.Sprintf("%s/reviews", baseURL), reviewReq)
	require.Equal(t, http.StatusCreated, status)
	assert.Equal(t, reviewReq.Rating, review.Rating)
	assert.Equal(t, reviewReq.Text, review.Text)

	status, reviews := httpGetJSON[dto.ReviewsResponse](t, client, fmt.Sprintf("%s/datasets/%d/reviews", baseURL, seed.datasetID))
	require.Equal(t, http.StatusOK, status)
	require.Len(t, reviews.Reviews, 1)
	assert.Equal(t, review.ID, reviews.Reviews[0].ID)

	status, summary := httpGetJSON[dto.RatingSummaryResponse](t, client, fmt.Sprintf("%s/datasets/%d/reviews/summary", baseURL, seed.datasetID))
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

func getBaseURL(t *testing.T) string {
	t.Helper()
	base := os.Getenv("E2E_BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	require.NotEmpty(t, base, "E2E_BASE_URL must point to a running API instance")
	return base
}

type seedResult struct {
	user      *entities.User
	category  *entities.Category
	datasetID uint64
}

func seedDataset(t *testing.T) seedResult {
	t.Helper()
	ctx := context.Background()
	logger := zap.NewNop()

	cfg, err := config.Load()
	require.NoError(t, err, "failed to load config.yaml")

	db, err := postqbuild.NewPool(ctx, cfg.Database)
	require.NoError(t, err, "failed to connect to postgres")
	t.Cleanup(func() { db.Close() })

	ensureBucket(t, &cfg.Storage)
	stor, err := storage.NewS3Storage(cfg.Storage)
	require.NoError(t, err, "failed to init storage client")

	catRepo := postqbuild.NewCategoryRepo(db, logger)
	dsRepo := postqbuild.NewDatasetRepo(db, logger)
	verRepo := postqbuild.NewVersionRepo(db, logger)
	mdRepo := postqbuild.NewMetadataRepo(db, logger)
	userRepo := postqbuild.NewUserRepo(db, logger)

	dsSvc := services.NewDatasetService(dsRepo, verRepo, mdRepo, stor, logger)

	fabric := testdata.NewFabric()
	now := time.Now().UTC().UnixNano()

	user := fabric.RegularUser()
	user.Email = fmt.Sprintf("e2e-user-%d@example.com", now)
	user.Username = fmt.Sprintf("e2e_user_%d", now)
	require.NoError(t, userRepo.Create(ctx, user))

	category := fabric.Category()
	category.Name = fmt.Sprintf("E2E Category %d", now)
	require.NoError(t, catRepo.Create(ctx, category))

	createCmd := fabric.CreateDatasetCommand(category, user)
	createCmd.Name = fmt.Sprintf("E2E Dataset %d", now)
	createCmd.ActorID = user.ID

	reader, size := fabric.DatasetFile("e2e initial payload")
	datasetID, err := dsSvc.CreateDataset(ctx, createCmd, reader, size)
	require.NoError(t, err)

	return seedResult{
		user:      user,
		category:  category,
		datasetID: datasetID,
	}
}

func ensureBucket(t *testing.T, cfg *config.Storage) {
	t.Helper()
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: strings.HasPrefix(cfg.Endpoint, "https"),
		Region: cfg.Region,
	})
	require.NoError(t, err, "failed to init minio client")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	require.NoError(t, err, "failed to check bucket")
	if exists {
		return
	}
	require.NoError(t, client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region}))
}

func extractIDs(categories []dto.CategoryResponse) []uint64 {
	out := make([]uint64, 0, len(categories))
	for _, c := range categories {
		out = append(out, c.ID)
	}
	return out
}

func extractDatasetIDs(datasets []dto.DatasetResponse) []uint64 {
	out := make([]uint64, 0, len(datasets))
	for _, d := range datasets {
		out = append(out, d.ID)
	}
	return out
}
