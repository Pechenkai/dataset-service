//go:build integration

package minio_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ppo/internal/storage"
	"ppo/internal/tests/integration"
)

func TestS3Storage_UploadAndGetURL(t *testing.T) {
	minioContainer := integration.StartMinio(t, "integration-datasets")
	cfg := minioContainer.Config

	s3, err := storage.NewS3Storage(cfg)
	require.NoError(t, err)

	key := "datasets/42/v0.1/test.bin"
	payload := []byte("integration payload")

	_, err = s3.Upload(context.Background(), key, bytes.NewReader(payload), int64(len(payload)))
	require.NoError(t, err)

	obj, err := minioContainer.Client.GetObject(context.Background(), cfg.Bucket, key, minio.GetObjectOptions{})
	require.NoError(t, err)
	defer obj.Close()

	data, err := io.ReadAll(obj)
	require.NoError(t, err)
	assert.Equal(t, payload, data)

	url, err := s3.GetURL(context.Background(), key)
	require.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, key)

	require.NoError(t, s3.Delete(context.Background(), key))

	_, err = minioContainer.Client.StatObject(context.Background(), cfg.Bucket, key, minio.StatObjectOptions{})
	require.Error(t, err, "object should be deleted after test cleanup")
}
