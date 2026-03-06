package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

// --- Mock implementations ---

type mockMediaStorage struct {
	uploadURL string
	uploadErr error
}

func (m *mockMediaStorage) Upload(_ context.Context, _ io.Reader, _ string, _ string) (string, error) {
	if m.uploadErr != nil {
		return "", m.uploadErr
	}
	return m.uploadURL, nil
}

func (m *mockMediaStorage) Delete(_ context.Context, _ string) error {
	return nil
}

type mockMediaRepo struct {
	created *model.Media
	err     error
}

func (m *mockMediaRepo) Create(_ context.Context, media *model.Media) error {
	if m.err != nil {
		return m.err
	}
	media.ID = uuid.New()
	m.created = media
	return nil
}

func (m *mockMediaRepo) FindByID(_ context.Context, _ uuid.UUID) (*model.Media, error) {
	return nil, nil
}

func (m *mockMediaRepo) FindByPostID(_ context.Context, _ uuid.UUID) ([]model.Media, error) {
	return nil, nil
}

func (m *mockMediaRepo) LinkToPost(_ context.Context, _ []uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockMediaRepo) FindByIDs(_ context.Context, _ []uuid.UUID) ([]model.Media, error) {
	return nil, nil
}

func (m *mockMediaRepo) UnlinkByPostID(_ context.Context, _ uuid.UUID) error {
	return nil
}

// --- Tests ---

func TestMediaService_Upload_MIMETypeValidation(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantErr     bool
	}{
		{name: "image/jpeg allowed", contentType: "image/jpeg", wantErr: false},
		{name: "image/png allowed", contentType: "image/png", wantErr: false},
		{name: "image/webp allowed", contentType: "image/webp", wantErr: false},
		{name: "image/gif allowed", contentType: "image/gif", wantErr: false},
		{name: "video/mp4 allowed", contentType: "video/mp4", wantErr: false},
		{name: "video/webm allowed", contentType: "video/webm", wantErr: false},
		{name: "application/pdf rejected", contentType: "application/pdf", wantErr: true},
		{name: "text/plain rejected", contentType: "text/plain", wantErr: true},
		{name: "audio/mpeg rejected", contentType: "audio/mpeg", wantErr: true},
		{name: "image/svg+xml rejected", contentType: "image/svg+xml", wantErr: true},
		{name: "empty string rejected", contentType: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &mockMediaStorage{uploadURL: "https://cdn.example.com/test.jpg"}
			repo := &mockMediaRepo{}
			svc := NewMediaService(storage, repo)

			_, err := svc.Upload(
				context.Background(),
				uuid.New(),
				strings.NewReader("fake-file-data"),
				"test.jpg",
				tt.contentType,
				1024, // 1KB, well under all limits
			)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error for unsupported MIME type, got nil")
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T: %v", err, err)
				}
				if appErr.Code != 400 {
					t.Errorf("expected status 400, got %d", appErr.Code)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestMediaService_Upload_FileSizeLimit(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		size        int64
		wantErr     bool
	}{
		// Image: 5MB limit
		{name: "image exactly 5MB", contentType: "image/jpeg", size: 5 * 1024 * 1024, wantErr: false},
		{name: "image over 5MB", contentType: "image/png", size: 5*1024*1024 + 1, wantErr: true},
		{name: "image 1KB", contentType: "image/webp", size: 1024, wantErr: false},

		// GIF: 15MB limit
		{name: "gif exactly 15MB", contentType: "image/gif", size: 15 * 1024 * 1024, wantErr: false},
		{name: "gif over 15MB", contentType: "image/gif", size: 15*1024*1024 + 1, wantErr: true},
		{name: "gif 1KB", contentType: "image/gif", size: 1024, wantErr: false},

		// Video: 50MB limit
		{name: "video exactly 50MB", contentType: "video/mp4", size: 50 * 1024 * 1024, wantErr: false},
		{name: "video over 50MB", contentType: "video/mp4", size: 50*1024*1024 + 1, wantErr: true},
		{name: "video 1KB", contentType: "video/webm", size: 1024, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &mockMediaStorage{uploadURL: "https://cdn.example.com/test.file"}
			repo := &mockMediaRepo{}
			svc := NewMediaService(storage, repo)

			_, err := svc.Upload(
				context.Background(),
				uuid.New(),
				strings.NewReader("fake"),
				"test.file",
				tt.contentType,
				tt.size,
			)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error for oversized file, got nil")
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T: %v", err, err)
				}
				if appErr.Code != 400 {
					t.Errorf("expected status 400, got %d", appErr.Code)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestMediaService_Upload_Success(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		filename    string
		wantType    string
	}{
		{name: "jpeg image", contentType: "image/jpeg", filename: "photo.jpg", wantType: "image"},
		{name: "png image", contentType: "image/png", filename: "screenshot.png", wantType: "image"},
		{name: "gif", contentType: "image/gif", filename: "animation.gif", wantType: "gif"},
		{name: "mp4 video", contentType: "video/mp4", filename: "clip.mp4", wantType: "video"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedURL := fmt.Sprintf("https://cdn.example.com/%s", tt.filename)
			storage := &mockMediaStorage{uploadURL: expectedURL}
			repo := &mockMediaRepo{}
			svc := NewMediaService(storage, repo)

			resp, err := svc.Upload(
				context.Background(),
				uuid.New(),
				strings.NewReader("fake-file-data"),
				tt.filename,
				tt.contentType,
				2048,
			)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if resp.URL != expectedURL {
				t.Errorf("expected URL %q, got %q", expectedURL, resp.URL)
			}
			if resp.Type != tt.wantType {
				t.Errorf("expected type %q, got %q", tt.wantType, resp.Type)
			}
			if resp.MimeType != tt.contentType {
				t.Errorf("expected mimeType %q, got %q", tt.contentType, resp.MimeType)
			}
			if resp.Size != 2048 {
				t.Errorf("expected size 2048, got %d", resp.Size)
			}
			if resp.ID == "" {
				t.Error("expected non-empty ID")
			}
		})
	}
}

func TestMediaService_Upload_StorageError(t *testing.T) {
	storage := &mockMediaStorage{uploadErr: fmt.Errorf("storage unavailable")}
	repo := &mockMediaRepo{}
	svc := NewMediaService(storage, repo)

	_, err := svc.Upload(
		context.Background(),
		uuid.New(),
		strings.NewReader("data"),
		"test.jpg",
		"image/jpeg",
		1024,
	)

	if err == nil {
		t.Fatal("expected error when storage fails, got nil")
	}
}

func TestMediaService_Upload_RepoError(t *testing.T) {
	storage := &mockMediaStorage{uploadURL: "https://cdn.example.com/test.jpg"}
	repo := &mockMediaRepo{err: fmt.Errorf("db error")}
	svc := NewMediaService(storage, repo)

	_, err := svc.Upload(
		context.Background(),
		uuid.New(),
		strings.NewReader("data"),
		"test.jpg",
		"image/jpeg",
		1024,
	)

	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}
