package service

import (
	"context"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
)

type PostService interface {
	CreatePost(ctx context.Context, authorID uuid.UUID, req dto.CreatePostRequest) (*dto.PostDetailResponse, error)
	GetPostByID(ctx context.Context, id uuid.UUID) (*dto.PostDetailResponse, error)
	GetPosts(ctx context.Context) ([]dto.PostDetailResponse, error)
}

type postService struct {
	postRepo repository.PostRepository
}

func NewPostService(postRepo repository.PostRepository) PostService {
	return &postService{postRepo: postRepo}
}

func (s *postService) CreatePost(ctx context.Context, authorID uuid.UUID, req dto.CreatePostRequest) (*dto.PostDetailResponse, error) {
	content := req.Content
	if utf8.RuneCountInString(content) == 0 {
		return nil, apperror.BadRequest("content must not be empty")
	}
	if utf8.RuneCountInString(content) > 280 {
		return nil, apperror.BadRequest("content must not exceed 280 characters")
	}

	visibility := model.VisibilityPublic
	if req.Visibility != "" {
		switch model.Visibility(req.Visibility) {
		case model.VisibilityPublic, model.VisibilityFriends, model.VisibilityPrivate:
			visibility = model.Visibility(req.Visibility)
		default:
			return nil, apperror.BadRequest("invalid visibility: %s", req.Visibility)
		}
	}

	post := &model.Post{
		AuthorID:   authorID,
		Content:    content,
		Visibility: visibility,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, apperror.Internal("failed to create post")
	}

	result, err := s.postRepo.FindByID(ctx, post.ID)
	if err != nil {
		return nil, apperror.Internal("failed to retrieve created post")
	}

	resp := dto.ToPostDetailResponse(*result)
	return &resp, nil
}

func (s *postService) GetPostByID(ctx context.Context, id uuid.UUID) (*dto.PostDetailResponse, error) {
	result, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Internal("failed to retrieve post")
	}

	resp := dto.ToPostDetailResponse(*result)
	return &resp, nil
}

func (s *postService) GetPosts(ctx context.Context) ([]dto.PostDetailResponse, error) {
	posts, err := s.postRepo.FindAll(ctx, 50, 0)
	if err != nil {
		return nil, apperror.Internal("failed to retrieve posts")
	}

	responses := make([]dto.PostDetailResponse, len(posts))
	for i, p := range posts {
		responses[i] = dto.ToPostDetailResponse(p)
	}
	return responses, nil
}
