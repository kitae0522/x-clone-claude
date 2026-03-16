package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type mockPostRepo struct {
	posts map[uuid.UUID]*model.PostWithAuthor

	existsIncludingDeletedFn    func(ctx context.Context, id uuid.UUID) (bool, bool, error)
	findByIDIncludingDeletedFn  func(ctx context.Context, id uuid.UUID) (*model.PostWithAuthor, error)
	findDeletedByAuthorFn       func(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) ([]model.PostWithAuthor, error)
	restoreFn                   func(ctx context.Context, id uuid.UUID) error
	restoreReplyFn              func(ctx context.Context, id, parentID uuid.UUID) error
	hardDeleteFn                func(ctx context.Context, id uuid.UUID) error
}

func newMockPostRepo() *mockPostRepo {
	return &mockPostRepo{
		posts: make(map[uuid.UUID]*model.PostWithAuthor),
	}
}

func (m *mockPostRepo) Create(_ context.Context, post *model.Post) error {
	post.ID = uuid.New()
	m.posts[post.ID] = &model.PostWithAuthor{
		Post:                  *post,
		AuthorUsername:        "testuser",
		AuthorDisplayName:     "Test User",
		AuthorProfileImageURL: "",
	}
	return nil
}

func (m *mockPostRepo) FindByID(_ context.Context, id uuid.UUID) (*model.PostWithAuthor, error) {
	if p, ok := m.posts[id]; ok {
		return p, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockPostRepo) FindAll(_ context.Context, _, _ int) ([]model.PostWithAuthor, error) {
	var posts []model.PostWithAuthor
	for _, p := range m.posts {
		posts = append(posts, *p)
	}
	return posts, nil
}

func (m *mockPostRepo) FindByIDWithUser(_ context.Context, id, _ uuid.UUID) (*model.PostWithAuthor, error) {
	return m.FindByID(context.Background(), id)
}

func (m *mockPostRepo) FindAllWithUser(_ context.Context, limit, offset int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return m.FindAll(context.Background(), limit, offset)
}

func (m *mockPostRepo) CreateReply(_ context.Context, post *model.Post) error {
	post.ID = uuid.New()
	m.posts[post.ID] = &model.PostWithAuthor{
		Post:                  *post,
		AuthorUsername:        "replyuser",
		AuthorDisplayName:     "Reply User",
		AuthorProfileImageURL: "",
	}
	if post.ParentID != nil {
		if parent, ok := m.posts[*post.ParentID]; ok {
			parent.ReplyCount++
		}
	}
	return nil
}

func (m *mockPostRepo) FindRepliesByPostID(_ context.Context, postID uuid.UUID, _, _ int) ([]model.PostWithAuthor, error) {
	var replies []model.PostWithAuthor
	for _, p := range m.posts {
		if p.ParentID != nil && *p.ParentID == postID {
			replies = append(replies, *p)
		}
	}
	return replies, nil
}

func (m *mockPostRepo) FindRepliesByPostIDWithUser(_ context.Context, postID, _ uuid.UUID, limit, offset int) ([]model.PostWithAuthor, error) {
	return m.FindRepliesByPostID(context.Background(), postID, limit, offset)
}

func (m *mockPostRepo) FindAuthorReplyByPostID(_ context.Context, postID, authorID uuid.UUID) (*model.PostWithAuthor, error) {
	for _, p := range m.posts {
		if p.ParentID != nil && *p.ParentID == postID && p.AuthorID == authorID {
			return p, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *mockPostRepo) FindAuthorReplyByPostIDWithUser(_ context.Context, postID, authorID, _ uuid.UUID) (*model.PostWithAuthor, error) {
	return m.FindAuthorReplyByPostID(context.Background(), postID, authorID)
}

func (m *mockPostRepo) FindByAuthorHandle(_ context.Context, _ string, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}

func (m *mockPostRepo) FindByAuthorHandleWithUser(_ context.Context, _ string, _, _ int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return nil, nil
}

func (m *mockPostRepo) FindRepliesByAuthorHandle(_ context.Context, _ string, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}

func (m *mockPostRepo) FindRepliesByAuthorHandleWithUser(_ context.Context, _ string, _, _ int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return nil, nil
}

func (m *mockPostRepo) FindLikedByUserHandle(_ context.Context, _ string, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}

func (m *mockPostRepo) FindLikedByUserHandleWithViewer(_ context.Context, _ string, _, _ int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return nil, nil
}

func (m *mockPostRepo) IncrementViewCount(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockPostRepo) IncrementViewCountBatch(_ context.Context, _ []uuid.UUID) error {
	return nil
}

func (m *mockPostRepo) Update(_ context.Context, id uuid.UUID, content string, visibility model.Visibility, locationLat *float64, locationLng *float64, locationName *string) error {
	if p, ok := m.posts[id]; ok {
		p.Content = content
		p.Visibility = visibility
		p.LocationLat = locationLat
		p.LocationLng = locationLng
		p.LocationName = locationName
		return nil
	}
	return pgx.ErrNoRows
}

func (m *mockPostRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.posts[id]; ok {
		delete(m.posts, id)
		return nil
	}
	return pgx.ErrNoRows
}

func (m *mockPostRepo) SoftDeleteReply(_ context.Context, id uuid.UUID, parentID uuid.UUID) error {
	if _, ok := m.posts[id]; ok {
		delete(m.posts, id)
		if parent, ok := m.posts[parentID]; ok {
			parent.ReplyCount--
		}
		return nil
	}
	return pgx.ErrNoRows
}

func (m *mockPostRepo) ExistsIncludingDeleted(ctx context.Context, id uuid.UUID) (bool, bool, error) {
	if m.existsIncludingDeletedFn != nil {
		return m.existsIncludingDeletedFn(ctx, id)
	}
	return false, false, nil
}

func (m *mockPostRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*model.PostWithAuthor, error) {
	if m.findByIDIncludingDeletedFn != nil {
		return m.findByIDIncludingDeletedFn(ctx, id)
	}
	return nil, pgx.ErrNoRows
}

func (m *mockPostRepo) FindDeletedByAuthor(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) ([]model.PostWithAuthor, error) {
	if m.findDeletedByAuthorFn != nil {
		return m.findDeletedByAuthorFn(ctx, userID, limit, cursor)
	}
	return nil, nil
}

func (m *mockPostRepo) Restore(ctx context.Context, id uuid.UUID) error {
	if m.restoreFn != nil {
		return m.restoreFn(ctx, id)
	}
	return nil
}

func (m *mockPostRepo) RestoreReply(ctx context.Context, id, parentID uuid.UUID) error {
	if m.restoreReplyFn != nil {
		return m.restoreReplyFn(ctx, id, parentID)
	}
	return nil
}

func (m *mockPostRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	if m.hardDeleteFn != nil {
		return m.hardDeleteFn(ctx, id)
	}
	return nil
}

type mockPollRepo struct{}

func newMockPollRepo() *mockPollRepo {
	return &mockPollRepo{}
}

func (m *mockPollRepo) CreatePoll(_ context.Context, _ *model.Poll, _ []model.PollOption) error {
	return nil
}

func (m *mockPollRepo) FindByPostID(_ context.Context, _ uuid.UUID) (*model.Poll, []model.PollOption, error) {
	return nil, nil, nil
}

func (m *mockPollRepo) Vote(_ context.Context, _, _ uuid.UUID, _ int16) error {
	return nil
}

func (m *mockPollRepo) Unvote(_ context.Context, _, _ uuid.UUID, _ int16) error {
	return nil
}

func (m *mockPollRepo) GetUserVote(_ context.Context, _, _ uuid.UUID) (*int16, error) {
	return nil, nil
}

func (m *mockPollRepo) FindByPostIDs(_ context.Context, _ []uuid.UUID) (map[uuid.UUID]model.Poll, map[uuid.UUID][]model.PollOption, error) {
	return nil, nil, nil
}

func (m *mockPollRepo) DeleteByPostID(_ context.Context, _ uuid.UUID) error {
	return nil
}

func TestCreatePost_Success(t *testing.T) {
	repo := newMockPostRepo()
	svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

	resp, err := svc.CreatePost(context.Background(), uuid.New(), dto.CreatePostRequest{
		Content:    "Hello, world!",
		Visibility: "public",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "Hello, world!" {
		t.Errorf("expected content 'Hello, world!', got %s", resp.Content)
	}
	if resp.Author.Username != "testuser" {
		t.Errorf("expected author username 'testuser', got %s", resp.Author.Username)
	}
}

func TestCreatePost_EmptyContent(t *testing.T) {
	repo := newMockPostRepo()
	svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

	_, err := svc.CreatePost(context.Background(), uuid.New(), dto.CreatePostRequest{
		Content: "",
	})

	if err == nil {
		t.Fatal("expected error for empty content")
	}
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != 400 {
		t.Errorf("expected 400, got %d", appErr.Code)
	}
}

func TestCreatePost_ExceedsMaxLength(t *testing.T) {
	repo := newMockPostRepo()
	svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

	longContent := strings.Repeat("a", 501)
	_, err := svc.CreatePost(context.Background(), uuid.New(), dto.CreatePostRequest{
		Content: longContent,
	})

	if err == nil {
		t.Fatal("expected error for content exceeding 500 characters")
	}
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != 400 {
		t.Errorf("expected 400, got %d", appErr.Code)
	}
}

func TestGetPostByID_Success(t *testing.T) {
	repo := newMockPostRepo()
	svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

	created, _ := svc.CreatePost(context.Background(), uuid.New(), dto.CreatePostRequest{
		Content: "Test post",
	})

	postID, _ := uuid.Parse(created.ID)
	resp, err := svc.GetPostByID(context.Background(), postID, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "Test post" {
		t.Errorf("expected content 'Test post', got %s", resp.Content)
	}
}

func TestGetPostByID_NotFound(t *testing.T) {
	repo := newMockPostRepo()
	svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

	_, err := svc.GetPostByID(context.Background(), uuid.New(), nil)
	if err == nil {
		t.Fatal("expected error for nonexistent post")
	}
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != 404 {
		t.Errorf("expected 404, got %d", appErr.Code)
	}
}

func TestCreateReply(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		parentExist bool
		wantErr     bool
		wantCode    int
	}{
		{
			name:        "success",
			content:     "This is a reply",
			parentExist: true,
			wantErr:     false,
		},
		{
			name:        "empty content",
			content:     "",
			parentExist: true,
			wantErr:     true,
			wantCode:    400,
		},
		{
			name:        "content exceeds 500 chars",
			content:     strings.Repeat("a", 501),
			parentExist: true,
			wantErr:     true,
			wantCode:    400,
		},
		{
			name:        "parent post not found",
			content:     "Reply to nothing",
			parentExist: false,
			wantErr:     true,
			wantCode:    404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockPostRepo()
			svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

			var parentID uuid.UUID
			if tt.parentExist {
				parent, _ := svc.CreatePost(context.Background(), uuid.New(), dto.CreatePostRequest{
					Content: "Parent post",
				})
				parentID, _ = uuid.Parse(parent.ID)
			} else {
				parentID = uuid.New()
			}

			resp, err := svc.CreateReply(context.Background(), parentID, uuid.New(), dto.CreateReplyRequest{
				Content: tt.content,
			})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T", err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("expected code %d, got %d", tt.wantCode, appErr.Code)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp.Content != tt.content {
					t.Errorf("expected content %q, got %q", tt.content, resp.Content)
				}
				if resp.ParentID == nil || *resp.ParentID != parentID.String() {
					t.Errorf("expected parentID %s, got %v", parentID, resp.ParentID)
				}
			}
		})
	}
}

func TestListReplies(t *testing.T) {
	tests := []struct {
		name        string
		replyCount  int
		parentExist bool
		wantErr     bool
		wantCode    int
	}{
		{
			name:        "list replies for post with 2 replies",
			replyCount:  2,
			parentExist: true,
			wantErr:     false,
		},
		{
			name:        "list replies for post with no replies",
			replyCount:  0,
			parentExist: true,
			wantErr:     false,
		},
		{
			name:        "parent post not found",
			replyCount:  0,
			parentExist: false,
			wantErr:     true,
			wantCode:    404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockPostRepo()
			svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

			var parentID uuid.UUID
			if tt.parentExist {
				parent, _ := svc.CreatePost(context.Background(), uuid.New(), dto.CreatePostRequest{
					Content: "Parent post",
				})
				parentID, _ = uuid.Parse(parent.ID)

				for i := 0; i < tt.replyCount; i++ {
					_, _ = svc.CreateReply(context.Background(), parentID, uuid.New(), dto.CreateReplyRequest{
						Content: "Reply content",
					})
				}
			} else {
				parentID = uuid.New()
			}

			replies, err := svc.ListReplies(context.Background(), parentID, nil)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T", err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("expected code %d, got %d", tt.wantCode, appErr.Code)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if len(replies) != tt.replyCount {
					t.Errorf("expected %d replies, got %d", tt.replyCount, len(replies))
				}
			}
		})
	}
}

func TestCreateReply_IncrementsParentReplyCount(t *testing.T) {
	repo := newMockPostRepo()
	svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

	parent, _ := svc.CreatePost(context.Background(), uuid.New(), dto.CreatePostRequest{
		Content: "Parent post",
	})
	parentID, _ := uuid.Parse(parent.ID)

	_, _ = svc.CreateReply(context.Background(), parentID, uuid.New(), dto.CreateReplyRequest{
		Content: "First reply",
	})
	_, _ = svc.CreateReply(context.Background(), parentID, uuid.New(), dto.CreateReplyRequest{
		Content: "Second reply",
	})

	updated, err := svc.GetPostByID(context.Background(), parentID, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.ReplyCount != 2 {
		t.Errorf("expected reply_count 2, got %d", updated.ReplyCount)
	}
}

func TestGetPostByID_IncrementsViewCount(t *testing.T) {
	tests := []struct {
		name             string
		initialViewCount int
		wantViewCount    int
	}{
		{
			name:             "view count increments from 0 to 1",
			initialViewCount: 0,
			wantViewCount:    1,
		},
		{
			name:             "view count increments from 5 to 6",
			initialViewCount: 5,
			wantViewCount:    6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockPostRepo()
			svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

			// Insert a post directly into the mock repo with a known view count
			postID := uuid.New()
			authorID := uuid.New()
			repo.posts[postID] = &model.PostWithAuthor{
				Post: model.Post{
					ID:         postID,
					AuthorID:   authorID,
					Content:    "test post",
					Visibility: model.VisibilityPublic,
					ViewCount:  tt.initialViewCount,
				},
				AuthorUsername:    "testuser",
				AuthorDisplayName: "Test User",
			}

			_, err := svc.GetPostByID(context.Background(), postID, nil)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			// Verify the view count was incremented in the mock repo
			got := repo.posts[postID].ViewCount
			if got != tt.wantViewCount {
				t.Errorf("expected view count %d, got %d", tt.wantViewCount, got)
			}
		})
	}
}

func TestGetPosts_DoesNotIncrementViewCount(t *testing.T) {
	tests := []struct {
		name             string
		initialViewCount int
	}{
		{
			name:             "view count stays 0 after GetPosts",
			initialViewCount: 0,
		},
		{
			name:             "view count stays 10 after GetPosts",
			initialViewCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockPostRepo()
			svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

			// Insert a post directly into the mock repo
			postID := uuid.New()
			authorID := uuid.New()
			repo.posts[postID] = &model.PostWithAuthor{
				Post: model.Post{
					ID:         postID,
					AuthorID:   authorID,
					Content:    "test post",
					Visibility: model.VisibilityPublic,
					ViewCount:  tt.initialViewCount,
				},
				AuthorUsername:    "testuser",
				AuthorDisplayName: "Test User",
			}

			_, err := svc.GetPosts(context.Background(), nil)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			// Verify the view count was NOT incremented
			got := repo.posts[postID].ViewCount
			if got != tt.initialViewCount {
				t.Errorf("expected view count to remain %d, got %d", tt.initialViewCount, got)
			}
		})
	}
}

func TestGetPostByID_VisibilityAccess(t *testing.T) {
	authorID := uuid.New()
	followerID := uuid.New()
	nonFollowerID := uuid.New()

	tests := []struct {
		name       string
		visibility model.Visibility
		viewerID   *uuid.UUID
		isFollower bool
		wantErr    bool
		wantCode   int
	}{
		{
			name:       "public post - unauthenticated viewer",
			visibility: model.VisibilityPublic,
			viewerID:   nil,
			wantErr:    false,
		},
		{
			name:       "public post - authenticated viewer",
			visibility: model.VisibilityPublic,
			viewerID:   &nonFollowerID,
			wantErr:    false,
		},
		{
			name:       "follower post - unauthenticated viewer",
			visibility: model.VisibilityFollower,
			viewerID:   nil,
			wantErr:    true,
			wantCode:   404,
		},
		{
			name:       "follower post - follower viewer",
			visibility: model.VisibilityFollower,
			viewerID:   &followerID,
			isFollower: true,
			wantErr:    false,
		},
		{
			name:       "follower post - non-follower viewer",
			visibility: model.VisibilityFollower,
			viewerID:   &nonFollowerID,
			isFollower: false,
			wantErr:    true,
			wantCode:   404,
		},
		{
			name:       "follower post - author self-view",
			visibility: model.VisibilityFollower,
			viewerID:   &authorID,
			wantErr:    false,
		},
		{
			name:       "private post - unauthenticated viewer",
			visibility: model.VisibilityPrivate,
			viewerID:   nil,
			wantErr:    true,
			wantCode:   404,
		},
		{
			name:       "private post - other user viewer",
			visibility: model.VisibilityPrivate,
			viewerID:   &followerID,
			wantErr:    true,
			wantCode:   404,
		},
		{
			name:       "private post - author self-view",
			visibility: model.VisibilityPrivate,
			viewerID:   &authorID,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postRepo := newMockPostRepo()
			followRepo := newMockFollowRepo()

			if tt.isFollower {
				followRepo.follows[followKey(followerID, authorID)] = true
			}

			svc := NewPostService(postRepo, newMockPollRepo(), nil, followRepo, nil)

			// Create a post with the specified visibility directly in the mock repo
			postID := uuid.New()
			postRepo.posts[postID] = &model.PostWithAuthor{
				Post: model.Post{
					ID:         postID,
					AuthorID:   authorID,
					Content:    "test post",
					Visibility: tt.visibility,
				},
				AuthorUsername:    "testauthor",
				AuthorDisplayName: "Test Author",
			}

			resp, err := svc.GetPostByID(context.Background(), postID, tt.viewerID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T", err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("expected code %d, got %d", tt.wantCode, appErr.Code)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp.Content != "test post" {
					t.Errorf("expected content 'test post', got %s", resp.Content)
				}
			}
		})
	}
}

func TestGetPostByID_Deleted(t *testing.T) {
	tests := []struct {
		name             string
		existsInDeleted  bool
		isDeleted        bool
		wantCode         int
		wantMessage      string
	}{
		{
			name:            "deleted post returns 410 Gone",
			existsInDeleted: true,
			isDeleted:       true,
			wantCode:        410,
			wantMessage:     "this post has been deleted",
		},
		{
			name:            "non-existent post returns 404",
			existsInDeleted: false,
			isDeleted:       false,
			wantCode:        404,
			wantMessage:     "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockPostRepo()
			repo.existsIncludingDeletedFn = func(_ context.Context, _ uuid.UUID) (bool, bool, error) {
				return tt.existsInDeleted, tt.isDeleted, nil
			}
			svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

			_, err := svc.GetPostByID(context.Background(), uuid.New(), nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			appErr, ok := err.(*apperror.AppError)
			if !ok {
				t.Fatalf("expected AppError, got %T", err)
			}
			if appErr.Code != tt.wantCode {
				t.Errorf("expected code %d, got %d", tt.wantCode, appErr.Code)
			}
			if appErr.Message != tt.wantMessage {
				t.Errorf("expected message %q, got %q", tt.wantMessage, appErr.Message)
			}
		})
	}
}

func TestListTrash(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	deletedAt := now.Add(-2 * 24 * time.Hour)

	makePost := func(id uuid.UUID) model.PostWithAuthor {
		return model.PostWithAuthor{
			Post: model.Post{
				ID:         id,
				AuthorID:   userID,
				Content:    "deleted post",
				Visibility: model.VisibilityPublic,
				DeletedAt:  &deletedAt,
				CreatedAt:  now.Add(-5 * 24 * time.Hour),
			},
			AuthorUsername:    "testuser",
			AuthorDisplayName: "Test User",
		}
	}

	tests := []struct {
		name       string
		limit      int
		posts      []model.PostWithAuthor
		wantCount  int
		wantMore   bool
		wantCursor bool
	}{
		{
			name:       "empty trash",
			limit:      20,
			posts:      nil,
			wantCount:  0,
			wantMore:   false,
			wantCursor: false,
		},
		{
			name:  "trash with posts within limit",
			limit: 20,
			posts: []model.PostWithAuthor{
				makePost(uuid.New()),
				makePost(uuid.New()),
			},
			wantCount:  2,
			wantMore:   false,
			wantCursor: false,
		},
		{
			name:  "trash with pagination (hasMore)",
			limit: 2,
			posts: func() []model.PostWithAuthor {
				// Return limit+1 posts to trigger hasMore
				return []model.PostWithAuthor{
					makePost(uuid.New()),
					makePost(uuid.New()),
					makePost(uuid.New()),
				}
			}(),
			wantCount:  2,
			wantMore:   true,
			wantCursor: true,
		},
		{
			name:  "invalid limit defaults to 20",
			limit: 0,
			posts: nil,
			wantCount:  0,
			wantMore:   false,
			wantCursor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockPostRepo()
			repo.findDeletedByAuthorFn = func(_ context.Context, _ uuid.UUID, _ int, _ *time.Time) ([]model.PostWithAuthor, error) {
				return tt.posts, nil
			}
			svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

			resp, err := svc.ListTrash(context.Background(), userID, tt.limit, nil)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if len(resp.Posts) != tt.wantCount {
				t.Errorf("expected %d posts, got %d", tt.wantCount, len(resp.Posts))
			}
			if resp.HasMore != tt.wantMore {
				t.Errorf("expected hasMore=%v, got %v", tt.wantMore, resp.HasMore)
			}
			if tt.wantCursor && resp.NextCursor == nil {
				t.Error("expected nextCursor to be set, got nil")
			}
			if !tt.wantCursor && resp.NextCursor != nil {
				t.Errorf("expected nextCursor to be nil, got %v", *resp.NextCursor)
			}
		})
	}
}

func TestRestorePost(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()
	postID := uuid.New()
	parentID := uuid.New()
	now := time.Now()
	recentDelete := now.Add(-5 * 24 * time.Hour)
	expiredDelete := now.Add(-31 * 24 * time.Hour)

	tests := []struct {
		name              string
		findIncludingDel  func(context.Context, uuid.UUID) (*model.PostWithAuthor, error)
		existsIncludingDel func(context.Context, uuid.UUID) (bool, bool, error)
		findByID          *model.PostWithAuthor
		requesterID       uuid.UUID
		wantErr           bool
		wantCode          int
		wantMessage       string
	}{
		{
			name: "success - restore top-level post",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return &model.PostWithAuthor{
					Post: model.Post{
						ID:         postID,
						AuthorID:   ownerID,
						Content:    "restored post",
						Visibility: model.VisibilityPublic,
						DeletedAt:  &recentDelete,
					},
					AuthorUsername:    "testuser",
					AuthorDisplayName: "Test User",
				}, nil
			},
			findByID: &model.PostWithAuthor{
				Post: model.Post{
					ID:         postID,
					AuthorID:   ownerID,
					Content:    "restored post",
					Visibility: model.VisibilityPublic,
				},
				AuthorUsername:    "testuser",
				AuthorDisplayName: "Test User",
			},
			requesterID: ownerID,
			wantErr:     false,
		},
		{
			name: "post not found",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return nil, pgx.ErrNoRows
			},
			requesterID: ownerID,
			wantErr:     true,
			wantCode:    404,
			wantMessage: "post not found",
		},
		{
			name: "post is not deleted",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return &model.PostWithAuthor{
					Post: model.Post{
						ID:        postID,
						AuthorID:  ownerID,
						DeletedAt: nil,
					},
				}, nil
			},
			requesterID: ownerID,
			wantErr:     true,
			wantCode:    400,
			wantMessage: "post is not deleted",
		},
		{
			name: "not owner",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return &model.PostWithAuthor{
					Post: model.Post{
						ID:        postID,
						AuthorID:  ownerID,
						DeletedAt: &recentDelete,
					},
				}, nil
			},
			requesterID: otherID,
			wantErr:     true,
			wantCode:    403,
			wantMessage: "you can only restore your own post",
		},
		{
			name: "expired - past 30 days",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return &model.PostWithAuthor{
					Post: model.Post{
						ID:        postID,
						AuthorID:  ownerID,
						DeletedAt: &expiredDelete,
					},
				}, nil
			},
			requesterID: ownerID,
			wantErr:     true,
			wantCode:    400,
			wantMessage: "post cannot be restored after 30 days",
		},
		{
			name: "reply with deleted parent",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return &model.PostWithAuthor{
					Post: model.Post{
						ID:        postID,
						AuthorID:  ownerID,
						ParentID:  &parentID,
						DeletedAt: &recentDelete,
					},
				}, nil
			},
			existsIncludingDel: func(_ context.Context, _ uuid.UUID) (bool, bool, error) {
				return true, true, nil // parent exists but is deleted
			},
			requesterID: ownerID,
			wantErr:     true,
			wantCode:    400,
			wantMessage: "cannot restore reply: parent post is deleted",
		},
		{
			name: "reply with missing parent",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return &model.PostWithAuthor{
					Post: model.Post{
						ID:        postID,
						AuthorID:  ownerID,
						ParentID:  &parentID,
						DeletedAt: &recentDelete,
					},
				}, nil
			},
			existsIncludingDel: func(_ context.Context, _ uuid.UUID) (bool, bool, error) {
				return false, false, nil // parent does not exist
			},
			requesterID: ownerID,
			wantErr:     true,
			wantCode:    400,
			wantMessage: "cannot restore reply: parent post is deleted",
		},
		{
			name: "success - restore reply with live parent",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return &model.PostWithAuthor{
					Post: model.Post{
						ID:        postID,
						AuthorID:  ownerID,
						ParentID:  &parentID,
						DeletedAt: &recentDelete,
					},
					AuthorUsername:    "testuser",
					AuthorDisplayName: "Test User",
				}, nil
			},
			existsIncludingDel: func(_ context.Context, _ uuid.UUID) (bool, bool, error) {
				return true, false, nil // parent exists and is NOT deleted
			},
			findByID: &model.PostWithAuthor{
				Post: model.Post{
					ID:       postID,
					AuthorID: ownerID,
					ParentID: &parentID,
					Content:  "restored reply",
				},
				AuthorUsername:    "testuser",
				AuthorDisplayName: "Test User",
			},
			requesterID: ownerID,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockPostRepo()
			repo.findByIDIncludingDeletedFn = tt.findIncludingDel
			if tt.existsIncludingDel != nil {
				repo.existsIncludingDeletedFn = tt.existsIncludingDel
			}
			// If we expect success, set up FindByID to return the restored post
			if tt.findByID != nil {
				repo.posts[postID] = tt.findByID
			}

			svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

			resp, err := svc.RestorePost(context.Background(), postID, tt.requesterID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T", err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("expected code %d, got %d", tt.wantCode, appErr.Code)
				}
				if appErr.Message != tt.wantMessage {
					t.Errorf("expected message %q, got %q", tt.wantMessage, appErr.Message)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp == nil {
					t.Fatal("expected response, got nil")
				}
			}
		})
	}
}

func TestPermanentDeletePost(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()
	postID := uuid.New()
	now := time.Now()
	deletedAt := now.Add(-2 * 24 * time.Hour)

	tests := []struct {
		name             string
		findIncludingDel func(context.Context, uuid.UUID) (*model.PostWithAuthor, error)
		requesterID      uuid.UUID
		wantErr          bool
		wantCode         int
		wantMessage      string
	}{
		{
			name: "success",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return &model.PostWithAuthor{
					Post: model.Post{
						ID:        postID,
						AuthorID:  ownerID,
						DeletedAt: &deletedAt,
					},
				}, nil
			},
			requesterID: ownerID,
			wantErr:     false,
		},
		{
			name: "post not found",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return nil, pgx.ErrNoRows
			},
			requesterID: ownerID,
			wantErr:     true,
			wantCode:    404,
			wantMessage: "post not found",
		},
		{
			name: "post not in trash",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return &model.PostWithAuthor{
					Post: model.Post{
						ID:        postID,
						AuthorID:  ownerID,
						DeletedAt: nil,
					},
				}, nil
			},
			requesterID: ownerID,
			wantErr:     true,
			wantCode:    400,
			wantMessage: "post is not in trash",
		},
		{
			name: "not owner",
			findIncludingDel: func(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
				return &model.PostWithAuthor{
					Post: model.Post{
						ID:        postID,
						AuthorID:  ownerID,
						DeletedAt: &deletedAt,
					},
				}, nil
			},
			requesterID: otherID,
			wantErr:     true,
			wantCode:    403,
			wantMessage: "you can only permanently delete your own post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockPostRepo()
			repo.findByIDIncludingDeletedFn = tt.findIncludingDel

			svc := NewPostService(repo, newMockPollRepo(), nil, nil, nil)

			err := svc.PermanentDeletePost(context.Background(), postID, tt.requesterID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T", err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("expected code %d, got %d", tt.wantCode, appErr.Code)
				}
				if appErr.Message != tt.wantMessage {
					t.Errorf("expected message %q, got %q", tt.wantMessage, appErr.Message)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}
