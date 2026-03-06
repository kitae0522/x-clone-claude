package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/service"
)

type MediaHandler struct {
	mediaService service.MediaService
}

func NewMediaHandler(ms service.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: ms}
}

func (h *MediaHandler) Upload(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return respondError(c, apperror.BadRequest("file is required"))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return respondError(c, apperror.Internal("failed to read uploaded file"))
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	resp, err := h.mediaService.Upload(
		c.Context(),
		userID,
		file,
		fileHeader.Filename,
		contentType,
		fileHeader.Size,
	)
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}
