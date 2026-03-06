package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/storage"
)

var (
	ErrUnsupportedMediaType = apperror.BadRequest("unsupported media type")
	ErrFileTooLarge         = apperror.BadRequest("file size exceeds limit")
)

var allowedMIMETypes = map[string]model.MediaType{
	"image/jpeg": model.MediaTypeImage,
	"image/png":  model.MediaTypeImage,
	"image/webp": model.MediaTypeImage,
	"image/gif":  model.MediaTypeGIF,
	"video/mp4":  model.MediaTypeVideo,
	"video/webm": model.MediaTypeVideo,
}

const (
	maxImageSize = 5 * 1024 * 1024  // 5MB
	maxGIFSize   = 15 * 1024 * 1024 // 15MB
	maxVideoSize = 50 * 1024 * 1024 // 50MB
)

type MediaService interface {
	Upload(ctx context.Context, uploaderID uuid.UUID, file io.Reader, filename string, contentType string, size int64) (*dto.MediaResponse, error)
}

type mediaService struct {
	storage  storage.MediaStorage
	mediaRepo repository.MediaRepository
}

func NewMediaService(s storage.MediaStorage, mr repository.MediaRepository) MediaService {
	return &mediaService{
		storage:   s,
		mediaRepo: mr,
	}
}

func (s *mediaService) Upload(ctx context.Context, uploaderID uuid.UUID, file io.Reader, filename string, contentType string, size int64) (*dto.MediaResponse, error) {
	// Validate MIME type
	mediaType, ok := allowedMIMETypes[contentType]
	if !ok {
		return nil, ErrUnsupportedMediaType
	}

	// Validate file size
	if err := validateFileSize(mediaType, size); err != nil {
		return nil, err
	}

	// Generate UUID-based filename preserving original extension
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = extensionFromMIME(contentType)
	}
	newFilename := uuid.New().String() + ext

	// Upload file to storage
	url, err := s.storage.Upload(ctx, file, newFilename, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Save media record to database
	media := &model.Media{
		UploaderID: uploaderID,
		URL:        url,
		MediaType:  mediaType,
		MimeType:   contentType,
		SizeBytes:  size,
	}

	if err := s.mediaRepo.Create(ctx, media); err != nil {
		return nil, fmt.Errorf("failed to save media record: %w", err)
	}

	return &dto.MediaResponse{
		ID:       media.ID.String(),
		URL:      media.URL,
		Type:     string(media.MediaType),
		MimeType: media.MimeType,
		Width:    media.Width,
		Height:   media.Height,
		Size:     media.SizeBytes,
		Duration: media.DurationSeconds,
	}, nil
}

func validateFileSize(mediaType model.MediaType, size int64) error {
	switch mediaType {
	case model.MediaTypeImage:
		if size > maxImageSize {
			return apperror.BadRequest("image file size exceeds 5MB limit")
		}
	case model.MediaTypeGIF:
		if size > maxGIFSize {
			return apperror.BadRequest("GIF file size exceeds 15MB limit")
		}
	case model.MediaTypeVideo:
		if size > maxVideoSize {
			return apperror.BadRequest("video file size exceeds 50MB limit")
		}
	}
	return nil
}

func extensionFromMIME(contentType string) string {
	mimeToExt := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
		"video/mp4":  ".mp4",
		"video/webm": ".webm",
	}

	if ext, ok := mimeToExt[strings.ToLower(contentType)]; ok {
		return ext
	}
	return ""
}
