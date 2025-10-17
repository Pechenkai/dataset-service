//go:build integration || e2e

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"ppo/internal/config"
)

type MinioContainer struct {
	Pool     *dockertest.Pool
	Resource *dockertest.Resource
	Client   *minio.Client
	Config   config.Storage
}

func StartMinio(t *testing.T, bucket string) *MinioContainer {
	t.Helper()

	pool, err := dockertest.NewPool("")
	require.NoError(t, err, "failed to connect to Docker daemon")

	const (
		accessKey = "minioadmin"
		secretKey = "minioadmin"
		region    = "ru-central"
	)

	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "quay.io/minio/minio",
			Tag:        "latest",
			Cmd:        []string{"server", "/data", "--console-address", ":9001"},
			Env: []string{
				fmt.Sprintf("MINIO_ROOT_USER=%s", accessKey),
				fmt.Sprintf("MINIO_ROOT_PASSWORD=%s", secretKey),
			},
		},
		func(h *docker.HostConfig) {
			h.AutoRemove = true
		},
	)
	require.NoError(t, err, "failed to start minio container")

	t.Cleanup(func() {
		_ = pool.Purge(resource)
	})

	endpoint := fmt.Sprintf("localhost:%s", resource.GetPort("9000/tcp"))

	var client *minio.Client
	require.NoError(t, pool.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var err error
		client, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: false,
			Region: region,
		})
		if err != nil {
			return err
		}

		_, err = client.ListBuckets(ctx)
		return err
	}), "minio container did not become ready in time")

	storageCfg := config.Storage{
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
		Bucket:    bucket,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region})
	if err != nil {
		exists, bucketErr := client.BucketExists(ctx, bucket)
		require.NoError(t, bucketErr, "failed to check existing bucket")
		if !exists {
			require.NoError(t, err, "failed to create minio bucket")
		}
	}

	return &MinioContainer{
		Pool:     pool,
		Resource: resource,
		Client:   client,
		Config:   storageCfg,
	}
}
