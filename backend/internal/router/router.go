package router

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/handler"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/middleware"
	"github.com/kitae0522/twitter-clone-claude/backend/pkg/config"
	"go.uber.org/fx"
)

type Params struct {
	fx.In
	App             *fiber.App
	Config          *config.Config
	Logger          *slog.Logger
	PostHandler     *handler.PostHandler
	AuthHandler     *handler.AuthHandler
	LikeHandler     *handler.LikeHandler
	FollowHandler   *handler.FollowHandler
	BookmarkHandler *handler.BookmarkHandler
	UserHandler     *handler.UserHandler
	PollHandler     *handler.PollHandler
	RepostHandler   *handler.RepostHandler
}

func Setup(p Params) {
	p.App.Use(middleware.RequestLogger(p.Logger))

	p.App.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := p.App.Group("/api")
	jwtSecret := p.Config.JWTSecret

	posts := api.Group("/posts")
	posts.Get("/", middleware.OptionalAuth(jwtSecret), p.PostHandler.GetPosts)
	posts.Post("/", middleware.AuthRequired(jwtSecret), p.PostHandler.CreatePost)
	posts.Get("/:id", middleware.OptionalAuth(jwtSecret), p.PostHandler.GetPostByID)
	posts.Put("/:id", middleware.AuthRequired(jwtSecret), p.PostHandler.UpdatePost)
	posts.Delete("/:id", middleware.AuthRequired(jwtSecret), p.PostHandler.DeletePost)
	posts.Post("/:id/reply", middleware.AuthRequired(jwtSecret), p.PostHandler.CreateReply)
	posts.Get("/:id/replies", middleware.OptionalAuth(jwtSecret), p.PostHandler.ListReplies)
	posts.Post("/:id/like", middleware.AuthRequired(jwtSecret), p.LikeHandler.Like)
	posts.Delete("/:id/like", middleware.AuthRequired(jwtSecret), p.LikeHandler.Unlike)
	posts.Post("/:id/bookmark", middleware.AuthRequired(jwtSecret), p.BookmarkHandler.Bookmark)
	posts.Delete("/:id/bookmark", middleware.AuthRequired(jwtSecret), p.BookmarkHandler.Unbookmark)
	posts.Post("/:id/vote", middleware.AuthRequired(jwtSecret), p.PollHandler.Vote)
	posts.Delete("/:id/vote", middleware.AuthRequired(jwtSecret), p.PollHandler.Unvote)
	posts.Post("/:id/repost", middleware.AuthRequired(jwtSecret), p.RepostHandler.Repost)
	posts.Delete("/:id/repost", middleware.AuthRequired(jwtSecret), p.RepostHandler.Unrepost)

	auth := api.Group("/auth")
	auth.Post("/register", p.AuthHandler.Register)
	auth.Post("/login", p.AuthHandler.Login)
	auth.Post("/logout", p.AuthHandler.Logout)
	auth.Get("/me", middleware.AuthRequired(jwtSecret), p.AuthHandler.Me)

	users := api.Group("/users")
	users.Get("/bookmarks", middleware.AuthRequired(jwtSecret), p.BookmarkHandler.ListBookmarks)
	users.Put("/profile", middleware.AuthRequired(jwtSecret), p.UserHandler.UpdateProfile)
	users.Post("/:handle/follow", middleware.AuthRequired(jwtSecret), p.FollowHandler.Follow)
	users.Delete("/:handle/follow", middleware.AuthRequired(jwtSecret), p.FollowHandler.Unfollow)
	users.Get("/:handle/following", p.FollowHandler.GetFollowing)
	users.Get("/:handle/followers", p.FollowHandler.GetFollowers)
	users.Get("/:handle/posts", middleware.OptionalAuth(jwtSecret), p.UserHandler.GetUserPosts)
	users.Get("/:handle/replies", middleware.OptionalAuth(jwtSecret), p.UserHandler.GetUserReplies)
	users.Get("/:handle/likes", middleware.OptionalAuth(jwtSecret), p.UserHandler.GetUserLikes)
	users.Get("/:handle", middleware.OptionalAuth(jwtSecret), p.UserHandler.GetProfile)
}
