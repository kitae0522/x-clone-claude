package service

import (
	"time"

	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type PostService interface {
	GetPosts() ([]model.Post, error)
}

type postService struct{}

func NewPostService() PostService {
	return &postService{}
}

func (s *postService) GetPosts() ([]model.Post, error) {
	now := time.Now()

	posts := []model.Post{
		{
			ID:         "550e8400-e29b-41d4-a716-446655440001",
			AuthorID:   "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			Content:    "Hello, world! This is my first public post.",
			Visibility: model.VisibilityPublic,
			CreatedAt:  now.Add(-2 * time.Hour),
			UpdatedAt:  now.Add(-2 * time.Hour),
		},
		{
			ID:         "550e8400-e29b-41d4-a716-446655440002",
			AuthorID:   "b2c3d4e5-f6a7-8901-bcde-f12345678901",
			Content:    "Only my friends can see this post!",
			Visibility: model.VisibilityFriends,
			CreatedAt:  now.Add(-1 * time.Hour),
			UpdatedAt:  now.Add(-1 * time.Hour),
		},
		{
			ID:         "550e8400-e29b-41d4-a716-446655440003",
			AuthorID:   "c3d4e5f6-a7b8-9012-cdef-123456789012",
			Content:    "This is a private note to myself.",
			Visibility: model.VisibilityPrivate,
			CreatedAt:  now.Add(-30 * time.Minute),
			UpdatedAt:  now.Add(-30 * time.Minute),
		},
	}

	return posts, nil
}
