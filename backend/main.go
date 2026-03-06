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

	likeRepo := repository.NewLikeRepository(pool)
	likeService := service.NewLikeService(likeRepo, postRepo)
	likeHandler := handler.NewLikeHandler(likeService)

	userRepo := repository.NewUserRepository(pool)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours)
	authHandler := handler.NewAuthHandler(authService)

	followRepo := repository.NewFollowRepository(pool)
	followService := service.NewFollowService(followRepo, userRepo)
	followHandler := handler.NewFollowHandler(followService)

	bookmarkRepo := repository.NewBookmarkRepository(pool)
	bookmarkService := service.NewBookmarkService(bookmarkRepo, postRepo)
	bookmarkHandler := handler.NewBookmarkHandler(bookmarkService)

	userService := service.NewUserService(userRepo, followRepo)
	userHandler := handler.NewUserHandler(userService, postService)

	api := app.Group("/api")
	posts := api.Group("/posts")
	posts.Get("/", middleware.OptionalAuth(cfg.JWTSecret), postHandler.GetPosts)
	posts.Post("/", middleware.AuthRequired(cfg.JWTSecret), postHandler.CreatePost)
	posts.Get("/:id", middleware.OptionalAuth(cfg.JWTSecret), postHandler.GetPostByID)
	posts.Post("/:id/reply", middleware.AuthRequired(cfg.JWTSecret), postHandler.CreateReply)
	posts.Get("/:id/replies", middleware.OptionalAuth(cfg.JWTSecret), postHandler.ListReplies)
	posts.Post("/:id/like", middleware.AuthRequired(cfg.JWTSecret), likeHandler.Like)
	posts.Delete("/:id/like", middleware.AuthRequired(cfg.JWTSecret), likeHandler.Unlike)
	posts.Post("/:id/bookmark", middleware.AuthRequired(cfg.JWTSecret), bookmarkHandler.Bookmark)
	posts.Delete("/:id/bookmark", middleware.AuthRequired(cfg.JWTSecret), bookmarkHandler.Unbookmark)

	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", authHandler.Logout)
	auth.Get("/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Me)

	users := api.Group("/users")
	users.Get("/bookmarks", middleware.AuthRequired(cfg.JWTSecret), bookmarkHandler.ListBookmarks)
	users.Put("/profile", middleware.AuthRequired(cfg.JWTSecret), userHandler.UpdateProfile)
	users.Post("/:handle/follow", middleware.AuthRequired(cfg.JWTSecret), followHandler.Follow)
	users.Delete("/:handle/follow", middleware.AuthRequired(cfg.JWTSecret), followHandler.Unfollow)
	users.Get("/:handle/following", followHandler.GetFollowing)
	users.Get("/:handle/followers", followHandler.GetFollowers)
	users.Get("/:handle/posts", middleware.OptionalAuth(cfg.JWTSecret), userHandler.GetUserPosts)
	users.Get("/:handle/replies", middleware.OptionalAuth(cfg.JWTSecret), userHandler.GetUserReplies)
	users.Get("/:handle/likes", middleware.OptionalAuth(cfg.JWTSecret), userHandler.GetUserLikes)
	users.Get("/:handle", middleware.OptionalAuth(cfg.JWTSecret), userHandler.GetProfile)

	log.Fatal(app.Listen(":8080"))
}
