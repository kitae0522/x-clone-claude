package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
)

type UserService interface {
	GetProfile(ctx context.Context, username string, viewerID *uuid.UUID) (*dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateProfileRequest) (*dto.UserResponse, error)
}

type userService struct {
	userRepo   repository.UserRepository
	followRepo repository.FollowRepository
}

func NewUserService(userRepo repository.UserRepository, followRepo repository.FollowRepository) UserService {
	return &userService{userRepo: userRepo, followRepo: followRepo}
}

func (s *userService) GetProfile(ctx context.Context, username string, viewerID *uuid.UUID) (*dto.ProfileResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Internal("failed to find user")
	}

	followersCount, err := s.followRepo.CountFollowers(ctx, user.ID)
	if err != nil {
		return nil, apperror.Internal("failed to count followers")
	}

	followingCount, err := s.followRepo.CountFollowing(ctx, user.ID)
	if err != nil {
		return nil, apperror.Internal("failed to count following")
	}

	isFollowing := false
	if viewerID != nil && *viewerID != user.ID {
		isFollowing, err = s.followRepo.IsFollowing(ctx, *viewerID, user.ID)
		if err != nil {
			return nil, apperror.Internal("failed to check follow status")
		}
	}

	resp := dto.ToProfileResponse(user, followersCount, followingCount, isFollowing)
	return &resp, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Internal("failed to find user")
	}

	if req.Username != "" && req.Username != user.Username {
		exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
		if err != nil {
			return nil, apperror.Internal("failed to check username")
		}
		if exists {
			return nil, apperror.Conflict("username already taken")
		}
		user.Username = req.Username
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	user.Bio = req.Bio
	user.ProfileImageURL = req.ProfileImageURL
	user.HeaderImageURL = req.HeaderImageURL

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Internal("failed to update profile")
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}
