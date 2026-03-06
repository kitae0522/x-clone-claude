package service

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type mockPostRepo struct {
	posts map[uuid.UUID]*model.PostWithAuthor
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

func TestCreatePost_Success(t *testing.T) {
	repo := newMockPostRepo()
	svc := NewPostService(repo, newMockPollRepo(), nil, nil)

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
	svc := NewPostService(repo, newMockPollRepo(), nil, nil)

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
	svc := NewPostService(repo, newMockPollRepo(), nil, nil)

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
	svc := NewPostService(repo, newMockPollRepo(), nil, nil)

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
	svc := NewPostService(repo, newMockPollRepo(), nil, nil)

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
			svc := NewPostService(repo, newMockPollRepo(), nil, nil)

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
			svc := NewPostService(repo, newMockPollRepo(), nil, nil)

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
	svc := NewPostService(repo, newMockPollRepo(), nil, nil)

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
			svc := NewPostService(repo, newMockPollRepo(), nil, nil)

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
			svc := NewPostService(repo, newMockPollRepo(), nil, nil)

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

			svc := NewPostService(postRepo, newMockPollRepo(), nil, followRepo)

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
