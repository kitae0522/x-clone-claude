package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetProfile(ctx context.Context, username string, viewerID *uuid.UUID) (*dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateProfileRequest) (*dto.UserResponse, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequest) error
	DeleteAccount(ctx context.Context, userID uuid.UUID, req dto.DeleteAccountRequest) error
}

type userService struct {
	userRepo   repository.UserRepository
	followRepo repository.FollowRepository
	likeRepo   repository.LikeRepository
}

func NewUserService(userRepo repository.UserRepository, followRepo repository.FollowRepository, likeRepo repository.LikeRepository) UserService {
	return &userService{userRepo: userRepo, followRepo: followRepo, likeRepo: likeRepo}
}

func (s *userService) GetProfile(ctx context.Context, username string, viewerID *uuid.UUID) (*dto.ProfileResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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
		if errors.Is(err, pgx.ErrNoRows) {
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
	if req.ProfileImageURL != "" {
		if !isAllowedImageURL(req.ProfileImageURL) {
			return nil, apperror.BadRequest("invalid profile image URL format")
		}
		user.ProfileImageURL = req.ProfileImageURL
	}
	if req.HeaderImageURL != "" {
		if !isAllowedImageURL(req.HeaderImageURL) {
			return nil, apperror.BadRequest("invalid header image URL format")
		}
		user.HeaderImageURL = req.HeaderImageURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Internal("failed to update profile")
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.NotFound("user not found")
		}
		return apperror.Internal("failed to find user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return apperror.Unauthorized("password is incorrect")
	}

	if req.CurrentPassword == req.NewPassword {
		return apperror.BadRequest("new password must be different from current password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperror.Internal("failed to hash password")
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, string(hashed)); err != nil {
		return apperror.Internal("failed to update password")
	}

	return nil
}

func (s *userService) DeleteAccount(ctx context.Context, userID uuid.UUID, req dto.DeleteAccountRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.NotFound("user not found")
		}
		return apperror.Internal("failed to find user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return apperror.Unauthorized("password is incorrect")
	}

	if _, err := s.likeRepo.SoftDeleteByUserID(ctx, userID); err != nil {
		return apperror.Internal("failed to soft delete likes")
	}

	if err := s.userRepo.SoftDelete(ctx, userID); err != nil {
		return apperror.Internal("failed to delete account")
	}

	return nil
}

func isAllowedImageURL(url string) bool {
	return strings.HasPrefix(url, "/media/") || strings.HasPrefix(url, "/uploads/") || strings.HasPrefix(url, "https://")
}
