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
	if err := parseAndValidate(c, &req); err != nil {
		return err
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

func (h *PostHandler) CreateReply(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	authorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	parentIDStr := c.Params("id")
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		return respondError(c, apperror.BadRequest("invalid post ID"))
	}

	var req dto.CreateReplyRequest
	if err := parseAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.postService.CreateReply(c.Context(), parentID, authorID, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (h *PostHandler) ListReplies(c *fiber.Ctx) error {
	parentIDStr := c.Params("id")
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		return respondError(c, apperror.BadRequest("invalid post ID"))
	}

	userID := extractOptionalUserID(c)

	replies, err := h.postService.ListReplies(c.Context(), parentID, userID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    replies,
	})
}

func (h *PostHandler) UpdatePost(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	requesterID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	postIDStr := c.Params("id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		return respondError(c, apperror.BadRequest("invalid post ID"))
	}

	var req dto.UpdatePostRequest
	if err := parseAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.postService.UpdatePost(c.Context(), postID, requesterID, req)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (h *PostHandler) DeletePost(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	requesterID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	postIDStr := c.Params("id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		return respondError(c, apperror.BadRequest("invalid post ID"))
	}

	if err := h.postService.DeletePost(c.Context(), postID, requesterID); err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    dto.DeletePostResponse{Message: "post deleted successfully"},
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
