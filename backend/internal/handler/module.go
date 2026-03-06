package handler

import "go.uber.org/fx"

var Module = fx.Module("handler",
	fx.Provide(
		NewPostHandler,
		NewAuthHandler,
		NewLikeHandler,
		NewFollowHandler,
		NewBookmarkHandler,
		NewUserHandler,
	),
)
