//go:build integration || e2e

package integration

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
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

	const (
		accessKey = "minioadmin"
		secretKey = "minioadmin"
		region    = "russia"
	)

	pool, err := dockertest.NewPool("")
	require.NoError(t, err, "failed to connect to Docker daemon")
	pool.MaxWait = 3 * time.Minute

	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "quay.io/minio/minio",
			Tag:        "latest",
			Cmd:        []string{"server", "/data", "--console-address", ":9001"},
			Env: []string{
				"MINIO_ROOT_USER=" + accessKey,
				"MINIO_ROOT_PASSWORD=" + secretKey,
			},
		},
		func(h *docker.HostConfig) { h.AutoRemove = true },
	)
	require.NoError(t, err, "failed to start minio container")

	t.Cleanup(func() { _ = pool.Purge(resource) })

	hostPort := resource.GetHostPort("9000/tcp")
	require.NotEmpty(t, hostPort, "no mapped host port for 9000/tcp")

	require.NoError(t, pool.Retry(func() error {
		d := net.Dialer{Timeout: 2 * time.Second}
		conn, err := d.Dial("tcp", hostPort)
		if err != nil {
			return err
		}
		_ = conn.Close()

		client := &http.Client{Timeout: 3 * time.Second}
		req, _ := http.NewRequest(http.MethodGet, "http://"+hostPort+"/minio/health/ready", nil)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("ready status %d", resp.StatusCode)
		}
		return nil
	}), "minio container did not become ready in time")

	cli, err := minio.New(hostPort, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
		Region: region,
	})
	require.NoError(t, err, "failed to init minio client")

	storageCfg := config.Storage{
		Endpoint:  hostPort,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
		Bucket:    bucket,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = cli.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region})
	if err != nil {
		exists, bucketErr := cli.BucketExists(ctx, bucket)
		require.NoError(t, bucketErr, "failed to check bucket existence")
		if !exists {
			require.NoError(t, err, "failed to create minio bucket")
		}
	}

	return &MinioContainer{
		Pool:     pool,
		Resource: resource,
		Client:   cli,
		Config:   storageCfg,
	}
}
