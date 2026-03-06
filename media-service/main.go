package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/kitae0522/twitter-clone-claude/media-service/internal/config"
	"github.com/kitae0522/twitter-clone-claude/media-service/internal/handler"
	"github.com/kitae0522/twitter-clone-claude/media-service/internal/storage"
	"github.com/kitae0522/twitter-clone-claude/media-service/internal/worker"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	store, err := storage.NewS3Storage(cfg)
	if err != nil {
		slog.Error("failed to init s3 storage", "error", err)
		os.Exit(1)
	}

	registry := worker.NewRegistry()
	processor := worker.NewProcessor(store, registry, cfg.MaxWorkers, cfg.TempDir)

	mediaHandler := handler.NewMediaHandler(store, processor, registry)

	app := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024, // 100MB
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, http://frontend:5173",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-API-Key",
		AllowMethods:     "GET, POST, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	app.Get("/health", mediaHandler.Health)

	// Public routes - media serving (no auth, UUID is unguessable)
	// Must be registered BEFORE the JWT group to avoid auth on serve
	app.Get("/media/:id", mediaHandler.Serve)

	// Authenticated routes (JWT required)
	app.Post("/media/upload", handler.JWTAuth(cfg.JWTSecret), mediaHandler.Upload)
	app.Get("/media/:id/status", handler.JWTAuth(cfg.JWTSecret), mediaHandler.GetStatus)

	// Internal routes - service-to-service (API key auth)
	internalRoutes := app.Group("/internal", handler.InternalAPIKey(cfg.APIKey))
	internalRoutes.Delete("/media/:id", mediaHandler.Delete)
	internalRoutes.Get("/media/:id/status", mediaHandler.GetStatus)

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("media-service starting", "port", cfg.Port)
	if err := app.Listen(addr); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
