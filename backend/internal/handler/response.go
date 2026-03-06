package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
)

func respondError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*apperror.AppError); ok {
		if appErr.Code >= 500 {
			slog.ErrorContext(c.Context(), "internal error",
				slog.String("error", appErr.Message),
				slog.String("path", c.Path()),
			)
		}
		return c.Status(appErr.Code).JSON(dto.APIResponse{
			Success: false,
			Error:   &appErr.Message,
		})
	}

	slog.ErrorContext(c.Context(), "unexpected error",
		slog.String("error", err.Error()),
		slog.String("path", c.Path()),
	)
	msg := "internal server error"
	return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse{
		Success: false,
		Error:   &msg,
	})
}
