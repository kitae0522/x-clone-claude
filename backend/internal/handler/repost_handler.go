package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/service"
)

type RepostHandler struct {
	repostService service.RepostService
}

func NewRepostHandler(rs service.RepostService) *RepostHandler {
	return &RepostHandler{repostService: rs}
}

func (h *RepostHandler) Repost(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	postID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respondError(c, apperror.BadRequest("invalid post ID"))
	}

	resp, err := h.repostService.Repost(c.Context(), userID, postID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (h *RepostHandler) Unrepost(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	postID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respondError(c, apperror.BadRequest("invalid post ID"))
	}

	resp, err := h.repostService.Unrepost(c.Context(), userID, postID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}
