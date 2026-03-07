package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
)

type RepostService interface {
	Repost(ctx context.Context, userID, postID uuid.UUID) (*dto.RepostStatusResponse, error)
	Unrepost(ctx context.Context, userID, postID uuid.UUID) (*dto.RepostStatusResponse, error)
}

type repostService struct {
	repostRepo repository.RepostRepository
	postRepo   repository.PostRepository
}

func NewRepostService(repostRepo repository.RepostRepository, postRepo repository.PostRepository) RepostService {
	return &repostService{repostRepo: repostRepo, postRepo: postRepo}
}

func (s *repostService) Repost(ctx context.Context, userID, postID uuid.UUID) (*dto.RepostStatusResponse, error) {
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Internal("failed to find post")
	}

	if post.AuthorID == userID {
		return nil, apperror.BadRequest("cannot repost your own post")
	}

	reposted, err := s.repostRepo.IsReposted(ctx, userID, postID)
	if err != nil {
		return nil, apperror.Internal("failed to check repost status")
	}
	if reposted {
		return nil, apperror.Conflict("already reposted")
	}

	if err := s.repostRepo.Repost(ctx, userID, postID); err != nil {
		return nil, apperror.Internal("failed to repost")
	}

	return &dto.RepostStatusResponse{Reposted: true}, nil
}

func (s *repostService) Unrepost(ctx context.Context, userID, postID uuid.UUID) (*dto.RepostStatusResponse, error) {
	_, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Internal("failed to find post")
	}

	reposted, err := s.repostRepo.IsReposted(ctx, userID, postID)
	if err != nil {
		return nil, apperror.Internal("failed to check repost status")
	}
	if !reposted {
		return nil, apperror.Conflict("not reposted yet")
	}

	if err := s.repostRepo.Unrepost(ctx, userID, postID); err != nil {
		return nil, apperror.Internal("failed to unrepost")
	}

	return &dto.RepostStatusResponse{Reposted: false}, nil
}
