package storage

import (
	"context"
	"io"
	"time"
)

type Client interface {
	Upload(ctx context.Context, path string, contentType string, reader io.Reader) (string, error)
	Delete(ctx context.Context, path string) error
	PresignGet(ctx context.Context, path string, ttl time.Duration) (string, error)
}
