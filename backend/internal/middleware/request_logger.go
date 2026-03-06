package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := uuid.New().String()
		c.Locals("requestID", requestID)

		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		attrs := []slog.Attr{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("latency", latency),
			slog.String("ip", c.IP()),
		}

		if userID, ok := c.Locals("userID").(string); ok {
			attrs = append(attrs, slog.String("user_id", userID))
		}

		level := slog.LevelInfo
		if c.Response().StatusCode() >= 500 {
			level = slog.LevelError
		} else if c.Response().StatusCode() >= 400 {
			level = slog.LevelWarn
		}

		logger.LogAttrs(c.Context(), level, "HTTP request", attrs...)
		return err
	}
}
