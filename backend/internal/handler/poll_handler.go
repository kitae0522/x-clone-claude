package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/service"
)

type PollHandler struct {
	pollService service.PollService
}

func NewPollHandler(ps service.PollService) *PollHandler {
	return &PollHandler{pollService: ps}
}

func (h *PollHandler) Vote(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	postIDStr := c.Params("id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		return respondError(c, apperror.BadRequest("invalid post ID"))
	}

	var req dto.VoteRequest
	if err := parseAndValidate(c, &req); err != nil {
		return err
	}

	resp, err := h.pollService.Vote(c.Context(), postID, userID, int16(req.OptionIndex))
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    resp,
	})
}
