package storage

import (
	"context"
	"io"
)

type Client interface {
	Upload(ctx context.Context, path string, contentType string, reader io.Reader) (string, error)
	Delete(ctx context.Context, path string) error
}
