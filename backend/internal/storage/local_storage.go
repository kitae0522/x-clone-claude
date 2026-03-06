package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type localStorage struct {
	basePath string
}

// NewLocalStorage creates a new local file system storage implementation.
func NewLocalStorage(basePath string) MediaStorage {
	return &localStorage{basePath: basePath}
}

func (s *localStorage) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	now := time.Now()
	dir := filepath.Join(s.basePath, fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()))

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	filePath := filepath.Join(dir, filename)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, file); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return a URL-style relative path using forward slashes
	relPath := filepath.ToSlash(filePath)
	if !strings.HasPrefix(relPath, "/") {
		relPath = "/" + relPath
	}

	return relPath, nil
}

func (s *localStorage) Delete(ctx context.Context, url string) error {
	path := strings.TrimPrefix(url, "/")
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	absBase, err := filepath.Abs(s.basePath)
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}
	if !strings.HasPrefix(absPath, absBase+string(os.PathSeparator)) {
		return fmt.Errorf("path traversal attempt blocked")
	}

	if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
