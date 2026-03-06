package storage

import (
	"context"
	"io"
)

// MediaStorage defines the interface for media file storage operations.
type MediaStorage interface {
	Upload(ctx context.Context, file io.Reader, filename string, contentType string) (url string, err error)
	Delete(ctx context.Context, url string) error
}
