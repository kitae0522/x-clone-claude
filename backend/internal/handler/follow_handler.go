package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/service"
)

type FollowHandler struct {
	followService service.FollowService
}

func NewFollowHandler(followService service.FollowService) *FollowHandler {
	return &FollowHandler{followService: followService}
}

func (h *FollowHandler) Follow(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	followerID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	handle := c.Params("handle")
	if handle == "" {
		return respondError(c, apperror.BadRequest("handle is required"))
	}

	resp, err := h.followService.Follow(c.Context(), followerID, handle)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (h *FollowHandler) Unfollow(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	followerID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	handle := c.Params("handle")
	if handle == "" {
		return respondError(c, apperror.BadRequest("handle is required"))
	}

	resp, err := h.followService.Unfollow(c.Context(), followerID, handle)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (h *FollowHandler) GetFollowing(c *fiber.Ctx) error {
	handle := c.Params("handle")
	if handle == "" {
		return respondError(c, apperror.BadRequest("handle is required"))
	}

	resp, err := h.followService.GetFollowing(c.Context(), handle)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (h *FollowHandler) GetFollowers(c *fiber.Ctx) error {
	handle := c.Params("handle")
	if handle == "" {
		return respondError(c, apperror.BadRequest("handle is required"))
	}

	resp, err := h.followService.GetFollowers(c.Context(), handle)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}
