package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth validates JWT tokens from cookie (primary) or Authorization header (fallback).
func JWTAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := c.Cookies("token")

		// Fallback to Authorization header
		if tokenStr == "" {
			auth := c.Get("Authorization")
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
			if tokenStr == auth {
				tokenStr = ""
			}
		}

		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "not authenticated",
			})
		}

		claims := jwt.MapClaims{}
		parsed, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secret), nil
		})

		if err != nil || !parsed.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "invalid or expired token",
			})
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			// Fallback to "sub" claim
			userID, ok = claims["sub"].(string)
		}
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "invalid token claims",
			})
		}

		c.Locals("userID", userID)
		return c.Next()
	}
}

// InternalAPIKey validates the X-API-Key header for service-to-service calls.
func InternalAPIKey(apiKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("X-API-Key")
		if key == "" || key != apiKey {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "invalid api key",
			})
		}
		return c.Next()
	}
}
