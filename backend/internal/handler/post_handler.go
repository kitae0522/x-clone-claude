package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/service"
)

type PostHandler struct {
	postService service.PostService
}

func NewPostHandler(ps service.PostService) *PostHandler {
	return &PostHandler{postService: ps}
}

func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	authorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	var req dto.CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return respondError(c, apperror.BadRequest("invalid request body"))
	}

	resp, err := h.postService.CreatePost(c.Context(), authorID, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (h *PostHandler) GetPostByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return respondError(c, apperror.BadRequest("invalid post ID"))
	}

	userID := extractOptionalUserID(c)

	resp, err := h.postService.GetPostByID(c.Context(), id, userID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (h *PostHandler) GetPosts(c *fiber.Ctx) error {
	userID := extractOptionalUserID(c)

	posts, err := h.postService.GetPosts(c.Context(), userID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    posts,
	})
}

func extractOptionalUserID(c *fiber.Ctx) *uuid.UUID {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return nil
	}
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil
	}
	return &id
}
