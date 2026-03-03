package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/service"
)

type PostHandler struct {
	postService service.PostService
}

func NewPostHandler(ps service.PostService) *PostHandler {
	return &PostHandler{postService: ps}
}

func (h *PostHandler) GetPosts(c *fiber.Ctx) error {
	posts, err := h.postService.GetPosts()
	if err != nil {
		errMsg := err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
			Success: false,
			Data:    nil,
			Error:   &errMsg,
		})
	}

	responses := make([]dto.PostResponse, len(posts))
	for i, p := range posts {
		responses[i] = dto.ToPostResponse(p)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    responses,
		Error:   nil,
	})
}
