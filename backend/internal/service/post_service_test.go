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

func TestCreatePost_Success(t *testing.T) {
	repo := newMockPostRepo()
	svc := NewPostService(repo)

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
	svc := NewPostService(repo)

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
	svc := NewPostService(repo)

	longContent := strings.Repeat("a", 281)
	_, err := svc.CreatePost(context.Background(), uuid.New(), dto.CreatePostRequest{
		Content: longContent,
	})

	if err == nil {
		t.Fatal("expected error for content exceeding 280 characters")
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
	svc := NewPostService(repo)

	created, _ := svc.CreatePost(context.Background(), uuid.New(), dto.CreatePostRequest{
		Content: "Test post",
	})

	postID, _ := uuid.Parse(created.ID)
	resp, err := svc.GetPostByID(context.Background(), postID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "Test post" {
		t.Errorf("expected content 'Test post', got %s", resp.Content)
	}
}

func TestGetPostByID_NotFound(t *testing.T) {
	repo := newMockPostRepo()
	svc := NewPostService(repo)

	_, err := svc.GetPostByID(context.Background(), uuid.New())
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
