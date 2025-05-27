package storage

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"ppo/internal/config"
)

type S3Storage struct {
	client *minio.Client
	bucket string
}

func NewS3Storage(cfg config.Storage) (*S3Storage, error) {
	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}
	return &S3Storage{client: cli, bucket: cfg.Bucket}, nil
}

func (s *S3Storage) Upload(ctx context.Context, key string, r io.Reader, size int64) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}
	presigned, err := s.client.PresignedGetObject(ctx, s.bucket, key, time.Hour, url.Values{})
	if err != nil {
		return "", err
	}
	return presigned.String(), nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *S3Storage) GetURL(ctx context.Context, key string) (string, error) {
	presigned, err := s.client.PresignedGetObject(ctx, s.bucket, key, time.Hour, url.Values{})
	if err != nil {
		return "", err
	}
	return presigned.String(), nil
}
