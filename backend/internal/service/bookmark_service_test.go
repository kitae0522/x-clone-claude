package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
)

// mockBookmarkRepo implements repository.BookmarkRepository for testing.
type mockBookmarkRepo struct {
	bookmarkFn     func(ctx context.Context, userID, postID uuid.UUID) error
	unbookmarkFn   func(ctx context.Context, userID, postID uuid.UUID) error
	isBookmarkedFn func(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	listByUserIDFn func(ctx context.Context, userID uuid.UUID, cursor time.Time, limit int) ([]model.PostWithAuthor, *time.Time, bool, error)
}

func (m *mockBookmarkRepo) Bookmark(ctx context.Context, userID, postID uuid.UUID) error {
	return m.bookmarkFn(ctx, userID, postID)
}

func (m *mockBookmarkRepo) Unbookmark(ctx context.Context, userID, postID uuid.UUID) error {
	return m.unbookmarkFn(ctx, userID, postID)
}

func (m *mockBookmarkRepo) IsBookmarked(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	if m.isBookmarkedFn != nil {
		return m.isBookmarkedFn(ctx, userID, postID)
	}
	return false, nil
}

func (m *mockBookmarkRepo) ListByUserID(ctx context.Context, userID uuid.UUID, cursor time.Time, limit int) ([]model.PostWithAuthor, *time.Time, bool, error) {
	return m.listByUserIDFn(ctx, userID, cursor, limit)
}

// mockPostRepoForBookmark implements repository.PostRepository for bookmark tests.
type mockPostRepoForBookmark struct {
	findByIDFn func(ctx context.Context, id uuid.UUID) (*model.PostWithAuthor, error)
}

func (m *mockPostRepoForBookmark) Create(_ context.Context, _ *model.Post) error { return nil }
func (m *mockPostRepoForBookmark) FindByID(ctx context.Context, id uuid.UUID) (*model.PostWithAuthor, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockPostRepoForBookmark) FindAll(_ context.Context, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) FindByIDWithUser(_ context.Context, _, _ uuid.UUID) (*model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) FindAllWithUser(_ context.Context, _, _ int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) CreateReply(_ context.Context, _ *model.Post) error { return nil }
func (m *mockPostRepoForBookmark) FindRepliesByPostID(_ context.Context, _ uuid.UUID, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) FindRepliesByPostIDWithUser(_ context.Context, _, _ uuid.UUID, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) FindAuthorReplyByPostID(_ context.Context, _, _ uuid.UUID) (*model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) FindAuthorReplyByPostIDWithUser(_ context.Context, _, _, _ uuid.UUID) (*model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) FindByAuthorHandle(_ context.Context, _ string, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) FindByAuthorHandleWithUser(_ context.Context, _ string, _, _ int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) FindRepliesByAuthorHandle(_ context.Context, _ string, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) FindRepliesByAuthorHandleWithUser(_ context.Context, _ string, _, _ int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) FindLikedByUserHandle(_ context.Context, _ string, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) FindLikedByUserHandleWithViewer(_ context.Context, _ string, _, _ int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForBookmark) IncrementViewCount(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockPostRepoForBookmark) IncrementViewCountBatch(_ context.Context, _ []uuid.UUID) error {
	return nil
}

func (m *mockPostRepoForBookmark) Update(_ context.Context, _ uuid.UUID, _ string, _ model.Visibility, _ *float64, _ *float64, _ *string) error {
	return nil
}

func (m *mockPostRepoForBookmark) SoftDelete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockPostRepoForBookmark) SoftDeleteReply(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockPostRepoForBookmark) ExistsIncludingDeleted(_ context.Context, _ uuid.UUID) (bool, bool, error) {
	return false, false, nil
}

func (m *mockPostRepoForBookmark) FindByIDIncludingDeleted(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
	return nil, nil
}

func (m *mockPostRepoForBookmark) FindDeletedByAuthor(_ context.Context, _ uuid.UUID, _ int, _ *time.Time) ([]model.PostWithAuthor, error) {
	return nil, nil
}

func (m *mockPostRepoForBookmark) Restore(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockPostRepoForBookmark) RestoreReply(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockPostRepoForBookmark) HardDelete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func TestBookmark(t *testing.T) {
	existingPostID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name         string
		postID       uuid.UUID
		postExists   bool
		postRepoErr  error
		bookmarkErr  error
		wantErr      bool
		wantCode     int
		wantBookmark bool
	}{
		{
			name:         "success - bookmark a post",
			postID:       existingPostID,
			postExists:   true,
			bookmarkErr:  nil,
			wantErr:      false,
			wantBookmark: true,
		},
		{
			name:       "error - post not found",
			postID:     uuid.New(),
			postExists: false,
			wantErr:    true,
			wantCode:   404,
		},
		{
			name:        "error - post repo returns internal error",
			postID:      existingPostID,
			postExists:  false,
			postRepoErr: errors.New("db connection error"),
			wantErr:     true,
			wantCode:    500,
		},
		{
			name:        "error - already bookmarked",
			postID:      existingPostID,
			postExists:  true,
			bookmarkErr: repository.ErrAlreadyBookmarked,
			wantErr:     true,
			wantCode:    409,
		},
		{
			name:        "error - bookmark repo internal error",
			postID:      existingPostID,
			postExists:  true,
			bookmarkErr: errors.New("db write error"),
			wantErr:     true,
			wantCode:    500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postRepo := &mockPostRepoForBookmark{
				findByIDFn: func(_ context.Context, id uuid.UUID) (*model.PostWithAuthor, error) {
					if tt.postRepoErr != nil {
						return nil, tt.postRepoErr
					}
					if !tt.postExists {
						return nil, pgx.ErrNoRows
					}
					return &model.PostWithAuthor{
						Post: model.Post{ID: id, Content: "test post"},
					}, nil
				},
			}
			bookmarkRepo := &mockBookmarkRepo{
				bookmarkFn: func(_ context.Context, _, _ uuid.UUID) error {
					return tt.bookmarkErr
				},
			}

			svc := NewBookmarkService(bookmarkRepo, postRepo)
			resp, err := svc.Bookmark(context.Background(), userID, tt.postID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T: %v", err, err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("expected code %d, got %d", tt.wantCode, appErr.Code)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp.Bookmarked != tt.wantBookmark {
					t.Errorf("expected bookmarked=%v, got %v", tt.wantBookmark, resp.Bookmarked)
				}
			}
		})
	}
}

func TestUnbookmark(t *testing.T) {
	existingPostID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name          string
		postID        uuid.UUID
		postExists    bool
		postRepoErr   error
		unbookmarkErr error
		wantErr       bool
		wantCode      int
		wantBookmark  bool
	}{
		{
			name:          "success - unbookmark a post",
			postID:        existingPostID,
			postExists:    true,
			unbookmarkErr: nil,
			wantErr:       false,
			wantBookmark:  false,
		},
		{
			name:          "error - not bookmarked yet",
			postID:        existingPostID,
			postExists:    true,
			unbookmarkErr: repository.ErrNotBookmarked,
			wantErr:       true,
			wantCode:      409,
		},
		{
			name:          "error - unbookmark repo internal error",
			postID:        existingPostID,
			postExists:    true,
			unbookmarkErr: errors.New("db delete error"),
			wantErr:       true,
			wantCode:      500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postRepo := &mockPostRepoForBookmark{
				findByIDFn: func(_ context.Context, id uuid.UUID) (*model.PostWithAuthor, error) {
					if tt.postRepoErr != nil {
						return nil, tt.postRepoErr
					}
					if !tt.postExists {
						return nil, pgx.ErrNoRows
					}
					return &model.PostWithAuthor{
						Post: model.Post{ID: id, Content: "test post"},
					}, nil
				},
			}
			bookmarkRepo := &mockBookmarkRepo{
				unbookmarkFn: func(_ context.Context, _, _ uuid.UUID) error {
					return tt.unbookmarkErr
				},
			}

			svc := NewBookmarkService(bookmarkRepo, postRepo)
			resp, err := svc.Unbookmark(context.Background(), userID, tt.postID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T: %v", err, err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("expected code %d, got %d", tt.wantCode, appErr.Code)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp.Bookmarked != tt.wantBookmark {
					t.Errorf("expected bookmarked=%v, got %v", tt.wantBookmark, resp.Bookmarked)
				}
			}
		})
	}
}

func TestListBookmarks(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	lastCursor := now.Add(-time.Hour)

	samplePosts := []model.PostWithAuthor{
		{
			Post: model.Post{
				ID:         uuid.New(),
				AuthorID:   uuid.New(),
				Content:    "bookmarked post 1",
				Visibility: model.VisibilityPublic,
				CreatedAt:  now.Add(-time.Hour),
				UpdatedAt:  now.Add(-time.Hour),
			},
			AuthorUsername:    "user1",
			AuthorDisplayName: "User One",
			IsBookmarked:      true,
		},
		{
			Post: model.Post{
				ID:         uuid.New(),
				AuthorID:   uuid.New(),
				Content:    "bookmarked post 2",
				Visibility: model.VisibilityPublic,
				CreatedAt:  now.Add(-2 * time.Hour),
				UpdatedAt:  now.Add(-2 * time.Hour),
			},
			AuthorUsername:    "user2",
			AuthorDisplayName: "User Two",
			IsBookmarked:      true,
		},
	}

	tests := []struct {
		name        string
		cursor      string
		limit       int
		posts       []model.PostWithAuthor
		lastCursor  *time.Time
		hasMore     bool
		repoErr     error
		wantErr     bool
		wantCode    int
		wantCount   int
		wantHasMore bool
		wantCursor  bool
	}{
		{
			name:        "success - list bookmarks with results",
			cursor:      "",
			limit:       20,
			posts:       samplePosts,
			lastCursor:  &lastCursor,
			hasMore:     true,
			wantErr:     false,
			wantCount:   2,
			wantHasMore: true,
			wantCursor:  true,
		},
		{
			name:        "success - empty bookmark list",
			cursor:      "",
			limit:       20,
			posts:       nil,
			lastCursor:  nil,
			hasMore:     false,
			wantErr:     false,
			wantCount:   0,
			wantHasMore: false,
			wantCursor:  false,
		},
		{
			name:        "success - with valid cursor",
			cursor:      now.Format(time.RFC3339),
			limit:       10,
			posts:       samplePosts[:1],
			lastCursor:  nil,
			hasMore:     false,
			wantErr:     false,
			wantCount:   1,
			wantHasMore: false,
			wantCursor:  false,
		},
		{
			name:     "error - invalid cursor format",
			cursor:   "not-a-valid-time",
			limit:    20,
			wantErr:  true,
			wantCode: 400,
		},
		{
			name:     "error - repo returns error",
			cursor:   "",
			limit:    20,
			repoErr:  errors.New("db query error"),
			wantErr:  true,
			wantCode: 500,
		},
		{
			name:        "success - limit defaults to 20 when 0",
			cursor:      "",
			limit:       0,
			posts:       samplePosts,
			lastCursor:  &lastCursor,
			hasMore:     false,
			wantErr:     false,
			wantCount:   2,
			wantHasMore: false,
			wantCursor:  false,
		},
		{
			name:        "success - limit defaults to 20 when negative",
			cursor:      "",
			limit:       -5,
			posts:       samplePosts,
			lastCursor:  nil,
			hasMore:     false,
			wantErr:     false,
			wantCount:   2,
			wantHasMore: false,
			wantCursor:  false,
		},
		{
			name:        "success - limit defaults to 20 when exceeds 50",
			cursor:      "",
			limit:       100,
			posts:       samplePosts,
			lastCursor:  nil,
			hasMore:     false,
			wantErr:     false,
			wantCount:   2,
			wantHasMore: false,
			wantCursor:  false,
		},
		{
			name:        "success - hasMore true but lastCursor nil results in no nextCursor",
			cursor:      "",
			limit:       20,
			posts:       samplePosts,
			lastCursor:  nil,
			hasMore:     true,
			wantErr:     false,
			wantCount:   2,
			wantHasMore: true,
			wantCursor:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedLimit int
			bookmarkRepo := &mockBookmarkRepo{
				listByUserIDFn: func(_ context.Context, _ uuid.UUID, _ time.Time, limit int) ([]model.PostWithAuthor, *time.Time, bool, error) {
					capturedLimit = limit
					if tt.repoErr != nil {
						return nil, nil, false, tt.repoErr
					}
					return tt.posts, tt.lastCursor, tt.hasMore, nil
				},
			}
			postRepo := &mockPostRepoForBookmark{
				findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
					return nil, pgx.ErrNoRows
				},
			}

			svc := NewBookmarkService(bookmarkRepo, postRepo)
			resp, err := svc.ListBookmarks(context.Background(), userID, tt.cursor, tt.limit)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T: %v", err, err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("expected code %d, got %d", tt.wantCode, appErr.Code)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if len(resp.Posts) != tt.wantCount {
				t.Errorf("expected %d posts, got %d", tt.wantCount, len(resp.Posts))
			}
			if resp.HasMore != tt.wantHasMore {
				t.Errorf("expected hasMore=%v, got %v", tt.wantHasMore, resp.HasMore)
			}
			if tt.wantCursor && resp.NextCursor == "" {
				t.Error("expected nextCursor to be set, got empty string")
			}
			if !tt.wantCursor && resp.NextCursor != "" {
				t.Errorf("expected no nextCursor, got %s", resp.NextCursor)
			}

			// Verify limit normalization: invalid limits should be normalized to 20
			if tt.limit <= 0 || tt.limit > 50 {
				if capturedLimit != 20 {
					t.Errorf("expected normalized limit 20, got %d", capturedLimit)
				}
			}
		})
	}
}
