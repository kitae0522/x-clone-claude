package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/handler"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/service"
)

func main() {
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	postService := service.NewPostService()
	postHandler := handler.NewPostHandler(postService)

	api := app.Group("/api")
	api.Get("/posts", postHandler.GetPosts)

	log.Fatal(app.Listen(":8080"))
}
