//go:build integration
// +build integration

package minio

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"ppo/internal/config"
	"testing"
	"time"

	"ppo/internal/storage"
)

var (
	s3       storage.Storage
	endpoint string
	bucket   = "test-bucket"
	access   = "minioadmin"
	secret   = "minioadmin"
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("docker pool error: %v", err)
	}
	opts := &dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "latest",
		Cmd:        []string{"server", "--address", ":9000", "--console-address", ":9001", "/data"},
		Env: []string{
			"MINIO_ROOT_USER=" + access,
			"MINIO_ROOT_PASSWORD=" + secret,
		},
	}
	resource, err := pool.RunWithOptions(opts, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("could not start minio: %v", err)
	}
	resource.Expire(120)

	pool.MaxWait = 60 * time.Second
	ip := resource.GetBoundIP("9000/tcp")
	if ip == "0.0.0.0" || ip == "" {
		ip = "127.0.0.1"
	}
	endpoint = fmt.Sprintf("%s:%s", ip, resource.GetPort("9000/tcp"))

	var minioClient *minio.Client
	if err := pool.Retry(func() error {
		cli, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(access, secret, ""),
			Secure: false,
		})
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := cli.ListBuckets(ctx); err != nil {
			return err
		}
		minioClient = cli
		return nil
	}); err != nil {
		log.Fatalf("could not connect to minio: %v", err)
	}

	ctx := context.Background()
	if err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		exists, err2 := minioClient.BucketExists(ctx, bucket)
		if err2 != nil || !exists {
			log.Fatalf("could not create bucket: %v", err)
		}
	}

	cfg := config.Storage{
		Endpoint:  endpoint,
		AccessKey: access,
		SecretKey: secret,
		Bucket:    bucket,
		Region:    "us-east-1",
	}
	s3, err = storage.NewS3Storage(cfg)
	if err != nil {
		log.Fatalf("NewS3Storage error: %v", err)
	}

	code := m.Run()

	pool.Purge(resource)
	os.Exit(code)
}

func TestS3Storage_UploadGetURLAndDelete(t *testing.T) {
	ctx := context.Background()
	key := "datasets/123/testfile.txt"

	data := []byte("hello minio")
	size := int64(len(data))
	reader := bytes.NewReader(data)

	url, err := s3.Upload(ctx, key, reader, size)
	assert.NoError(t, err)
	assert.Contains(t, url, key)

	gotURL, err := s3.GetURL(ctx, key)
	assert.NoError(t, err)
	assert.Contains(t, gotURL, key)

	err = s3.Delete(ctx, key)
	assert.NoError(t, err)
}
