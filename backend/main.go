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

	postRepo := repository.NewPostRepository(pool)
	postService := service.NewPostService(postRepo)
	postHandler := handler.NewPostHandler(postService)

	userRepo := repository.NewUserRepository(pool)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	authHandler := handler.NewAuthHandler(authService)

	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	api := app.Group("/api")
	posts := api.Group("/posts")
	posts.Get("/", postHandler.GetPosts)
	posts.Post("/", middleware.AuthRequired(cfg.JWTSecret), postHandler.CreatePost)
	posts.Get("/:id", postHandler.GetPostByID)

	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", authHandler.Logout)
	auth.Get("/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Me)

	users := api.Group("/users")
	users.Put("/profile", middleware.AuthRequired(cfg.JWTSecret), userHandler.UpdateProfile)
	users.Get("/:handle", userHandler.GetProfile)

	log.Fatal(app.Listen(":8080"))
}
