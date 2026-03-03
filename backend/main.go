package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/handler"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/middleware"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/service"
	"github.com/kitae0522/twitter-clone-claude/backend/pkg/config"
	"github.com/kitae0522/twitter-clone-claude/backend/pkg/database"
)

func main() {
	cfg := config.Load()

	pool, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := database.Migrate(pool, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	postService := service.NewPostService()
	postHandler := handler.NewPostHandler(postService)

	userRepo := repository.NewUserRepository(pool)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	authHandler := handler.NewAuthHandler(authService)

	api := app.Group("/api")
	api.Get("/posts", postHandler.GetPosts)

	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", authHandler.Logout)
	auth.Get("/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Me)

	log.Fatal(app.Listen(":8080"))
}
