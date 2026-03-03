package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
)

func AuthRequired(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := c.Cookies("token")
		if tokenStr == "" {
			msg := "authentication required"
			return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
				Success: false,
				Error:   &msg,
			})
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			msg := "invalid or expired token"
			return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
				Success: false,
				Error:   &msg,
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			msg := "invalid token claims"
			return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
				Success: false,
				Error:   &msg,
			})
		}

		sub, ok := claims["sub"].(string)
		if !ok {
			msg := "invalid token subject"
			return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse{
				Success: false,
				Error:   &msg,
			})
		}

		c.Locals("userID", sub)
		return c.Next()
	}
}
