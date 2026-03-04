package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
)

type LikeService interface {
	Like(ctx context.Context, userID, postID uuid.UUID) (*dto.LikeStatusResponse, error)
	Unlike(ctx context.Context, userID, postID uuid.UUID) (*dto.LikeStatusResponse, error)
}

type likeService struct {
	likeRepo repository.LikeRepository
	postRepo repository.PostRepository
}

func NewLikeService(likeRepo repository.LikeRepository, postRepo repository.PostRepository) LikeService {
	return &likeService{likeRepo: likeRepo, postRepo: postRepo}
}

func (s *likeService) Like(ctx context.Context, userID, postID uuid.UUID) (*dto.LikeStatusResponse, error) {
	_, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Internal("failed to find post")
	}

	liked, err := s.likeRepo.IsLiked(ctx, userID, postID)
	if err != nil {
		return nil, apperror.Internal("failed to check like status")
	}
	if liked {
		return nil, apperror.Conflict("already liked")
	}

	if err := s.likeRepo.Like(ctx, userID, postID); err != nil {
		return nil, apperror.Internal("failed to like post")
	}

	return &dto.LikeStatusResponse{Liked: true}, nil
}

func (s *likeService) Unlike(ctx context.Context, userID, postID uuid.UUID) (*dto.LikeStatusResponse, error) {
	_, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Internal("failed to find post")
	}

	liked, err := s.likeRepo.IsLiked(ctx, userID, postID)
	if err != nil {
		return nil, apperror.Internal("failed to check like status")
	}
	if !liked {
		return nil, apperror.Conflict("not liked yet")
	}

	if err := s.likeRepo.Unlike(ctx, userID, postID); err != nil {
		return nil, apperror.Internal("failed to unlike post")
	}

	return &dto.LikeStatusResponse{Liked: false}, nil
}
