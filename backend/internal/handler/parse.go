package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/pkg/validator"
)

var errResponseSent = errors.New("response already sent")

func parseAndValidate(c *fiber.Ctx, out interface{}) error {
	if err := c.BodyParser(out); err != nil {
		respondError(c, apperror.BadRequest("invalid request body"))
		return errResponseSent
	}
	if errs := validator.Validate(out); len(errs) > 0 {
		msg := "validation failed"
		c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
			Success: false,
			Error:   &msg,
			Data:    errs,
		})
		return errResponseSent
	}
	return nil
}
