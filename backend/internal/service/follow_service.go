package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
)

type FollowService interface {
	Follow(ctx context.Context, followerID uuid.UUID, targetHandle string) (*dto.FollowStatusResponse, error)
	Unfollow(ctx context.Context, followerID uuid.UUID, targetHandle string) (*dto.FollowStatusResponse, error)
	GetFollowing(ctx context.Context, handle string) (*dto.FollowListResponse, error)
	GetFollowers(ctx context.Context, handle string) (*dto.FollowListResponse, error)
}

type followService struct {
	followRepo repository.FollowRepository
	userRepo   repository.UserRepository
}

func NewFollowService(followRepo repository.FollowRepository, userRepo repository.UserRepository) FollowService {
	return &followService{followRepo: followRepo, userRepo: userRepo}
}

func (s *followService) Follow(ctx context.Context, followerID uuid.UUID, targetHandle string) (*dto.FollowStatusResponse, error) {
	targetUser, err := s.userRepo.FindByUsername(ctx, targetHandle)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Internal("failed to find user")
	}

	if followerID == targetUser.ID {
		return nil, apperror.BadRequest("cannot follow yourself")
	}

	already, err := s.followRepo.IsFollowing(ctx, followerID, targetUser.ID)
	if err != nil {
		return nil, apperror.Internal("failed to check follow status")
	}
	if already {
		return nil, apperror.Conflict("already following this user")
	}

	if err := s.followRepo.Follow(ctx, followerID, targetUser.ID); err != nil {
		return nil, apperror.Internal("failed to follow user")
	}

	return &dto.FollowStatusResponse{Following: true}, nil
}

func (s *followService) Unfollow(ctx context.Context, followerID uuid.UUID, targetHandle string) (*dto.FollowStatusResponse, error) {
	targetUser, err := s.userRepo.FindByUsername(ctx, targetHandle)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Internal("failed to find user")
	}

	if followerID == targetUser.ID {
		return nil, apperror.BadRequest("cannot unfollow yourself")
	}

	removed, err := s.followRepo.Unfollow(ctx, followerID, targetUser.ID)
	if err != nil {
		return nil, apperror.Internal("failed to unfollow user")
	}
	if !removed {
		return nil, apperror.NotFound("not following this user")
	}

	return &dto.FollowStatusResponse{Following: false}, nil
}

func (s *followService) GetFollowing(ctx context.Context, handle string) (*dto.FollowListResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, handle)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Internal("failed to find user")
	}

	users, err := s.followRepo.GetFollowing(ctx, user.ID)
	if err != nil {
		return nil, apperror.Internal("failed to get following list")
	}

	followUsers := make([]dto.FollowUserResponse, len(users))
	for i, u := range users {
		followUsers[i] = dto.ToFollowUserResponse(u)
	}

	return &dto.FollowListResponse{Users: followUsers, Total: len(followUsers)}, nil
}

func (s *followService) GetFollowers(ctx context.Context, handle string) (*dto.FollowListResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, handle)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Internal("failed to find user")
	}

	users, err := s.followRepo.GetFollowers(ctx, user.ID)
	if err != nil {
		return nil, apperror.Internal("failed to get followers list")
	}

	followUsers := make([]dto.FollowUserResponse, len(users))
	for i, u := range users {
		followUsers[i] = dto.ToFollowUserResponse(u)
	}

	return &dto.FollowListResponse{Users: followUsers, Total: len(followUsers)}, nil
}
