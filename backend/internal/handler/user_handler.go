package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/service"
)

type UserHandler struct {
	userService service.UserService
	postService service.PostService
}

func NewUserHandler(userService service.UserService, postService service.PostService) *UserHandler {
	return &UserHandler{userService: userService, postService: postService}
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	handle := c.Params("handle")
	if handle == "" {
		return respondError(c, apperror.BadRequest("handle is required"))
	}

	profile, err := h.userService.GetProfile(c.Context(), handle, h.getViewerID(c))
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    profile,
	})
}

func (h *UserHandler) getViewerID(c *fiber.Ctx) *uuid.UUID {
	if userIDStr, ok := c.Locals("userID").(string); ok {
		if id, err := uuid.Parse(userIDStr); err == nil {
			return &id
		}
	}
	return nil
}

func (h *UserHandler) GetUserPosts(c *fiber.Ctx) error {
	handle := c.Params("handle")
	if handle == "" {
		return respondError(c, apperror.BadRequest("handle is required"))
	}

	posts, err := h.postService.ListPostsByHandle(c.Context(), handle, h.getViewerID(c))
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    posts,
	})
}

func (h *UserHandler) GetUserReplies(c *fiber.Ctx) error {
	handle := c.Params("handle")
	if handle == "" {
		return respondError(c, apperror.BadRequest("handle is required"))
	}

	replies, err := h.postService.ListRepliesByHandle(c.Context(), handle, h.getViewerID(c))
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    replies,
	})
}

func (h *UserHandler) GetUserLikes(c *fiber.Ctx) error {
	handle := c.Params("handle")
	if handle == "" {
		return respondError(c, apperror.BadRequest("handle is required"))
	}

	likes, err := h.postService.ListLikedPostsByHandle(c.Context(), handle, h.getViewerID(c))
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    likes,
	})
}

func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	var req dto.UpdateProfileRequest
	if err := parseAndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.userService.UpdateProfile(c.Context(), userID, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    user,
	})
}
