package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
)

func respondError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*apperror.AppError); ok {
		return c.Status(appErr.Code).JSON(dto.APIResponse{
			Success: false,
			Error:   &appErr.Message,
		})
	}

	msg := "internal server error"
	return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
		Success: false,
		Error:   &msg,
	})
}
