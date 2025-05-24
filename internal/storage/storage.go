package storage

import (
	"context"
	"io"
)

type Storage interface {
	Upload(ctx context.Context, key string, r io.Reader, size int64) (url string, err error)
	Delete(ctx context.Context, key string) error
	GetURL(ctx context.Context, key string) (url string, err error)
}
