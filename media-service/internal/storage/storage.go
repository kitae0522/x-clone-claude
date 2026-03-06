package storage

import (
	"context"
	"io"
	"time"
)

// ObjectStorage is the interface for S3-compatible storage backends.
// Implementations: AWS S3, Cloudflare R2, MinIO.
type ObjectStorage interface {
	Put(ctx context.Context, key string, body io.Reader, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, string, error)
	Delete(ctx context.Context, key string) error
	PresignedURL(ctx context.Context, key string, expires time.Duration) (string, error)
}
