package main

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/handler"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/router"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/service"
	"github.com/kitae0522/twitter-clone-claude/backend/pkg/config"
	"github.com/kitae0522/twitter-clone-claude/backend/pkg/database"
	"github.com/kitae0522/twitter-clone-claude/backend/pkg/logger"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			config.Load,
			provideLogger,
			providePool,
			provideFiberApp,
		),
		repository.Module,
		service.Module,
		handler.Module,
		fx.Invoke(router.Setup),
		fx.Invoke(startServer),
	).Run()
}

func provideLogger(cfg *config.Config) *slog.Logger {
	l := logger.New(cfg.Env)
	slog.SetDefault(l)
	return l
}

func providePool(lc fx.Lifecycle, cfg *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	pool, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := database.Migrate(pool, "migrations"); err != nil {
		pool.Close()
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing database pool")
			pool.Close()
			return nil
		},
	})

	log.Info("database connected")
	return pool, nil
}

func provideFiberApp() *fiber.App {
	return fiber.New()
}

func startServer(lc fx.Lifecycle, app *fiber.App, log *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := app.Listen(":8080"); err != nil {
					log.Error("server failed", slog.String("error", err.Error()))
				}
			}()
			log.Info("server started", slog.String("addr", ":8080"))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("shutting down server")
			return app.Shutdown()
		},
	})
}
