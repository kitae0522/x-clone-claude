package handler

import (
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/media-service/internal/model"
	"github.com/kitae0522/twitter-clone-claude/media-service/internal/storage"
	"github.com/kitae0522/twitter-clone-claude/media-service/internal/worker"
)

// magicBytes maps file magic bytes to expected MIME type prefixes.
var magicBytes = map[string]string{
	"\xff\xd8\xff":      "image/jpeg",
	"\x89PNG\r\n\x1a\n": "image/png",
	"RIFF":              "image/webp", // also used by AVI, but we check content-type too
	"GIF87a":            "image/gif",
	"GIF89a":            "image/gif",
	"\x00\x00\x00":      "video/", // ftyp box (mp4/webm) starts after size bytes
}

type MediaHandler struct {
	store     storage.ObjectStorage
	processor *worker.Processor
	registry  *worker.Registry
}

func NewMediaHandler(store storage.ObjectStorage, proc *worker.Processor, reg *worker.Registry) *MediaHandler {
	return &MediaHandler{
		store:     store,
		processor: proc,
		registry:  reg,
	}
}

func (h *MediaHandler) Upload(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "not authenticated",
		})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "file is required",
		})
	}

	contentType := fileHeader.Header.Get("Content-Type")
	mediaType := classifyMIME(contentType)
	if mediaType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "unsupported media type: " + contentType,
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to read file",
		})
	}
	defer file.Close()

	// Validate magic bytes against declared content type
	header := make([]byte, 12)
	n, _ := file.Read(header)
	header = header[:n]

	if !validateMagicBytes(header, contentType) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "file content does not match declared content type",
		})
	}

	// Reset reader position
	if _, err := file.Seek(0, 0); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to process file",
		})
	}

	id := uuid.New().String()

	media := &model.Media{
		ID:        id,
		Status:    model.StatusPending,
		MediaType: mediaType,
		MimeType:  contentType,
		Size:      fileHeader.Size,
		CreatedAt: time.Now(),
	}
	h.registry.Set(id, media)

	h.processor.Enqueue(worker.Job{
		ID:          id,
		File:        file,
		ContentType: contentType,
		MediaType:   mediaType,
		Size:        fileHeader.Size,
	})

	slog.Info("media upload accepted", "id", id, "type", mediaType, "size", fileHeader.Size, "user", userID)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"data": model.UploadResponse{
			ID:     id,
			Status: model.StatusPending,
		},
	})
}

func (h *MediaHandler) GetStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid media id",
		})
	}

	media, ok := h.registry.Get(id)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "media not found",
		})
	}

	resp := model.StatusResponse{
		ID:        media.ID,
		Status:    media.Status,
		MediaType: media.MediaType,
		MimeType:  media.MimeType,
		Width:     media.Width,
		Height:    media.Height,
		Size:      media.Size,
		Error:     media.Error,
	}

	if media.Status == model.StatusReady {
		size := model.SizeVariant(c.Query("size", string(model.SizeMedium)))
		if !isValidSize(size) {
			size = model.SizeMedium
		}
		key := buildObjectKey(id, media.MediaType, size)
		url, err := h.store.PresignedURL(c.Context(), key, 1*time.Hour)
		if err != nil {
			slog.Error("failed to generate presigned url", "id", id, "error", err)
		} else {
			resp.URL = url
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *MediaHandler) Serve(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid media id",
		})
	}

	// Check in-memory registry for processing status
	if media, ok := h.registry.Get(id); ok && media.Status != model.StatusReady {
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"success": true,
			"data": model.StatusResponse{
				ID:     media.ID,
				Status: media.Status,
			},
		})
	}

	size := model.SizeVariant(c.Query("size", string(model.SizeMedium)))
	if !isValidSize(size) {
		size = model.SizeMedium
	}

	// Try each media type (image→webp, video→webm) until one succeeds
	extensions := []string{"webp", "webm"}
	for _, ext := range extensions {
		key := string(size) + "/" + id + "." + ext
		body, contentType, err := h.store.Get(c.Context(), key)
		if err != nil {
			continue
		}

		data, readErr := io.ReadAll(body)
		body.Close()
		if readErr != nil {
			continue
		}

		c.Set("Content-Type", contentType)
		c.Set("Cache-Control", "public, max-age=31536000, immutable")

		return c.Send(data)
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"success": false,
		"error":   "media not found",
	})
}

func (h *MediaHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid media id",
		})
	}

	media, ok := h.registry.Get(id)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "media not found",
		})
	}

	for _, v := range model.AllVariants {
		key := buildObjectKey(id, media.MediaType, v)
		if err := h.store.Delete(c.Context(), key); err != nil {
			slog.Error("failed to delete variant", "id", id, "variant", v, "error", err)
		}
	}

	h.registry.Delete(id)
	slog.Info("media deleted", "id", id)

	return c.JSON(fiber.Map{"success": true})
}

func (h *MediaHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "media-service",
	})
}

func classifyMIME(contentType string) model.MediaType {
	switch {
	case contentType == "image/gif":
		return model.MediaTypeGIF
	case strings.HasPrefix(contentType, "image/"):
		return model.MediaTypeImage
	case strings.HasPrefix(contentType, "video/"):
		return model.MediaTypeVideo
	default:
		return ""
	}
}

func validateMagicBytes(header []byte, contentType string) bool {
	if len(header) < 3 {
		return false
	}

	headerStr := string(header)
	for magic, expectedPrefix := range magicBytes {
		if strings.HasPrefix(headerStr, magic) {
			if strings.HasPrefix(contentType, expectedPrefix) {
				return true
			}
		}
	}

	// For video files, check for ftyp box (MP4/WebM container)
	if len(header) >= 8 && string(header[4:8]) == "ftyp" {
		return strings.HasPrefix(contentType, "video/")
	}

	// RIFF container can be WebP or AVI
	if len(header) >= 12 && string(header[0:4]) == "RIFF" {
		if string(header[8:12]) == "WEBP" {
			return contentType == "image/webp"
		}
	}

	// WebM uses EBML header (0x1A45DFA3)
	if len(header) >= 4 && header[0] == 0x1A && header[1] == 0x45 && header[2] == 0xDF && header[3] == 0xA3 {
		return contentType == "video/webm"
	}

	return false
}

func isValidSize(s model.SizeVariant) bool {
	for _, v := range model.AllVariants {
		if v == s {
			return true
		}
	}
	return false
}

func buildObjectKey(id string, mediaType model.MediaType, size model.SizeVariant) string {
	ext := "webp"
	if mediaType == model.MediaTypeVideo {
		ext = "webm"
	}
	return string(size) + "/" + id + "." + ext
}
