package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return respondError(c, apperror.BadRequest("invalid request body"))
	}

	resp, err := h.authService.Register(c.Context(), req)
	if err != nil {
		return respondError(c, err)
	}

	setTokenCookie(c, resp.Token)
	return c.Status(fiber.StatusCreated).JSON(dto.APIResponse{
		Success: true,
		Data:    resp.User,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return respondError(c, apperror.BadRequest("invalid request body"))
	}

	resp, err := h.authService.Login(c.Context(), req)
	if err != nil {
		return respondError(c, err)
	}

	setTokenCookie(c, resp.Token)
	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    resp.User,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    nil,
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return respondError(c, apperror.Unauthorized("not authenticated"))
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return respondError(c, apperror.Unauthorized("invalid user ID"))
	}

	user, err := h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(dto.APIResponse{
		Success: true,
		Data:    user,
	})
}

func setTokenCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
	})
}
