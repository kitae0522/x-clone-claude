package service

import "go.uber.org/fx"

var Module = fx.Module("service",
	fx.Provide(
		NewPostService,
		NewAuthService,
		NewLikeService,
		NewFollowService,
		NewBookmarkService,
		NewUserService,
		NewPollService,
		NewRepostService,
	),
)
