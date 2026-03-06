package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
)

type BookmarkService interface {
	Bookmark(ctx context.Context, userID, postID uuid.UUID) (*dto.BookmarkStatusResponse, error)
	Unbookmark(ctx context.Context, userID, postID uuid.UUID) (*dto.BookmarkStatusResponse, error)
	ListBookmarks(ctx context.Context, userID uuid.UUID, cursor string, limit int) (*dto.BookmarkListResponse, error)
}

type bookmarkService struct {
	bookmarkRepo repository.BookmarkRepository
	postRepo     repository.PostRepository
	nowFunc      func() time.Time
}

func NewBookmarkService(bookmarkRepo repository.BookmarkRepository, postRepo repository.PostRepository) BookmarkService {
	return &bookmarkService{bookmarkRepo: bookmarkRepo, postRepo: postRepo, nowFunc: time.Now}
}

func (s *bookmarkService) Bookmark(ctx context.Context, userID, postID uuid.UUID) (*dto.BookmarkStatusResponse, error) {
	_, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Internal("failed to find post")
	}

	if err := s.bookmarkRepo.Bookmark(ctx, userID, postID); err != nil {
		if errors.Is(err, repository.ErrAlreadyBookmarked) {
			return nil, apperror.Conflict("already bookmarked")
		}
		return nil, apperror.Internal("failed to bookmark post")
	}

	return &dto.BookmarkStatusResponse{Bookmarked: true}, nil
}

func (s *bookmarkService) Unbookmark(ctx context.Context, userID, postID uuid.UUID) (*dto.BookmarkStatusResponse, error) {
	if err := s.bookmarkRepo.Unbookmark(ctx, userID, postID); err != nil {
		if errors.Is(err, repository.ErrNotBookmarked) {
			return nil, apperror.Conflict("not bookmarked yet")
		}
		return nil, apperror.Internal("failed to unbookmark post")
	}

	return &dto.BookmarkStatusResponse{Bookmarked: false}, nil
}

func (s *bookmarkService) ListBookmarks(ctx context.Context, userID uuid.UUID, cursor string, limit int) (*dto.BookmarkListResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	cursorTime := s.nowFunc().Add(time.Second)
	if cursor != "" {
		parsed, err := time.Parse(time.RFC3339Nano, cursor)
		if err != nil {
			return nil, apperror.BadRequest("invalid cursor format")
		}
		cursorTime = parsed
	}

	posts, lastCursor, hasMore, err := s.bookmarkRepo.ListByUserID(ctx, userID, cursorTime, limit)
	if err != nil {
		return nil, apperror.Internal("failed to retrieve bookmarks")
	}

	responses := make([]dto.PostDetailResponse, len(posts))
	for i, p := range posts {
		responses[i] = dto.ToPostDetailResponse(p)
	}

	var nextCursor string
	if hasMore && lastCursor != nil {
		nextCursor = lastCursor.Format(time.RFC3339Nano)
	}

	return &dto.BookmarkListResponse{
		Posts:      responses,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
