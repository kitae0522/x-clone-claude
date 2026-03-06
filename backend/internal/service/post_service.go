package service

import (
	"context"
	"sort"
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
	GetPostByID(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*dto.PostDetailResponse, error)
	GetPosts(ctx context.Context, userID *uuid.UUID) ([]dto.PostDetailResponse, error)
	CreateReply(ctx context.Context, parentID, authorID uuid.UUID, req dto.CreateReplyRequest) (*dto.PostDetailResponse, error)
	ListReplies(ctx context.Context, parentID uuid.UUID, userID *uuid.UUID) ([]dto.PostDetailResponse, error)
	ListPostsByHandle(ctx context.Context, handle string, viewerID *uuid.UUID) ([]dto.PostDetailResponse, error)
	ListRepliesByHandle(ctx context.Context, handle string, viewerID *uuid.UUID) ([]dto.PostDetailResponse, error)
	ListLikedPostsByHandle(ctx context.Context, handle string, viewerID *uuid.UUID) ([]dto.PostDetailResponse, error)
}

const maxAuthorThreadDepth = 10

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

func (s *postService) GetPostByID(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*dto.PostDetailResponse, error) {
	var result *model.PostWithAuthor
	var err error

	if userID != nil {
		result, err = s.postRepo.FindByIDWithUser(ctx, id, *userID)
	} else {
		result, err = s.postRepo.FindByID(ctx, id)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Internal("failed to retrieve post")
	}

	resp := dto.ToPostDetailResponse(*result)

	replies, err := s.fetchReplies(ctx, id, userID)
	if err != nil {
		return nil, apperror.Internal("failed to retrieve replies")
	}

	postAuthorID := result.AuthorID

	sort.SliceStable(replies, func(i, j int) bool {
		iIsOP := replies[i].AuthorID == postAuthorID.String()
		jIsOP := replies[j].AuthorID == postAuthorID.String()
		if iIsOP != jIsOP {
			return iIsOP
		}
		return false
	})

	for i, reply := range replies {
		replyAuthorID, _ := uuid.Parse(reply.AuthorID)
		replyID, _ := uuid.Parse(reply.ID)
		thread, err := s.buildAuthorThread(ctx, replyID, replyAuthorID, postAuthorID, userID, 0)
		if err != nil {
			return nil, apperror.Internal("failed to build author thread")
		}
		replies[i].TopReplies = thread
	}

	resp.TopReplies = replies
	return &resp, nil
}

func (s *postService) fetchReplies(ctx context.Context, postID uuid.UUID, userID *uuid.UUID) ([]dto.PostDetailResponse, error) {
	var replies []model.PostWithAuthor
	var err error

	if userID != nil {
		replies, err = s.postRepo.FindRepliesByPostIDWithUser(ctx, postID, *userID, 50, 0)
	} else {
		replies, err = s.postRepo.FindRepliesByPostID(ctx, postID, 50, 0)
	}

	if err != nil {
		return nil, err
	}

	responses := make([]dto.PostDetailResponse, len(replies))
	for i, r := range replies {
		responses[i] = dto.ToPostDetailResponse(r)
	}
	return responses, nil
}

func (s *postService) buildAuthorThread(ctx context.Context, postID, replyAuthorID, postAuthorID uuid.UUID, userID *uuid.UUID, depth int) ([]dto.PostDetailResponse, error) {
	if depth >= maxAuthorThreadDepth {
		return nil, nil
	}

	var results []dto.PostDetailResponse

	selfReply, err := s.findAuthorReply(ctx, postID, replyAuthorID, userID)
	if err != nil {
		return nil, err
	}
	if selfReply != nil {
		resp := dto.ToPostDetailResponse(*selfReply)
		nested, err := s.buildAuthorThread(ctx, selfReply.ID, selfReply.AuthorID, postAuthorID, userID, depth+1)
		if err != nil {
			return nil, err
		}
		resp.TopReplies = nested
		results = append(results, resp)
	}

	if postAuthorID != replyAuthorID {
		opReply, err := s.findAuthorReply(ctx, postID, postAuthorID, userID)
		if err != nil {
			return nil, err
		}
		if opReply != nil {
			resp := dto.ToPostDetailResponse(*opReply)
			nested, err := s.buildAuthorThread(ctx, opReply.ID, opReply.AuthorID, postAuthorID, userID, depth+1)
			if err != nil {
				return nil, err
			}
			resp.TopReplies = nested
			results = append(results, resp)
		}
	}

	return results, nil
}

func (s *postService) findAuthorReply(ctx context.Context, postID, authorID uuid.UUID, userID *uuid.UUID) (*model.PostWithAuthor, error) {
	var result *model.PostWithAuthor
	var err error

	if userID != nil {
		result, err = s.postRepo.FindAuthorReplyByPostIDWithUser(ctx, postID, authorID, *userID)
	} else {
		result, err = s.postRepo.FindAuthorReplyByPostID(ctx, postID, authorID)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (s *postService) GetPosts(ctx context.Context, userID *uuid.UUID) ([]dto.PostDetailResponse, error) {
	var posts []model.PostWithAuthor
	var err error

	if userID != nil {
		posts, err = s.postRepo.FindAllWithUser(ctx, 50, 0, *userID)
	} else {
		posts, err = s.postRepo.FindAll(ctx, 50, 0)
	}

	if err != nil {
		return nil, apperror.Internal("failed to retrieve posts")
	}

	responses := make([]dto.PostDetailResponse, len(posts))
	for i, p := range posts {
		responses[i] = dto.ToPostDetailResponse(p)
	}
	return responses, nil
}

func (s *postService) CreateReply(ctx context.Context, parentID, authorID uuid.UUID, req dto.CreateReplyRequest) (*dto.PostDetailResponse, error) {
	content := req.Content
	if utf8.RuneCountInString(content) == 0 {
		return nil, apperror.BadRequest("content must not be empty")
	}
	if utf8.RuneCountInString(content) > 280 {
		return nil, apperror.BadRequest("content must not exceed 280 characters")
	}

	_, err := s.postRepo.FindByID(ctx, parentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("parent post not found")
		}
		return nil, apperror.Internal("failed to verify parent post")
	}

	post := &model.Post{
		AuthorID:   authorID,
		ParentID:   &parentID,
		Content:    content,
		Visibility: model.VisibilityPublic,
	}

	if err := s.postRepo.CreateReply(ctx, post); err != nil {
		return nil, apperror.Internal("failed to create reply")
	}

	result, err := s.postRepo.FindByID(ctx, post.ID)
	if err != nil {
		return nil, apperror.Internal("failed to retrieve created reply")
	}

	resp := dto.ToPostDetailResponse(*result)
	return &resp, nil
}

func (s *postService) ListReplies(ctx context.Context, parentID uuid.UUID, userID *uuid.UUID) ([]dto.PostDetailResponse, error) {
	_, err := s.postRepo.FindByID(ctx, parentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Internal("failed to verify post")
	}

	var replies []model.PostWithAuthor

	if userID != nil {
		replies, err = s.postRepo.FindRepliesByPostIDWithUser(ctx, parentID, *userID, 50, 0)
	} else {
		replies, err = s.postRepo.FindRepliesByPostID(ctx, parentID, 50, 0)
	}

	if err != nil {
		return nil, apperror.Internal("failed to retrieve replies")
	}

	responses := make([]dto.PostDetailResponse, len(replies))
	for i, r := range replies {
		responses[i] = dto.ToPostDetailResponse(r)
	}
	return responses, nil
}

func (s *postService) toPostDetailResponses(posts []model.PostWithAuthor) []dto.PostDetailResponse {
	responses := make([]dto.PostDetailResponse, len(posts))
	for i, p := range posts {
		responses[i] = dto.ToPostDetailResponse(p)
	}
	return responses
}

func (s *postService) ListPostsByHandle(ctx context.Context, handle string, viewerID *uuid.UUID) ([]dto.PostDetailResponse, error) {
	var posts []model.PostWithAuthor
	var err error

	if viewerID != nil {
		posts, err = s.postRepo.FindByAuthorHandleWithUser(ctx, handle, 50, 0, *viewerID)
	} else {
		posts, err = s.postRepo.FindByAuthorHandle(ctx, handle, 50, 0)
	}

	if err != nil {
		return nil, apperror.Internal("failed to retrieve user posts")
	}

	return s.toPostDetailResponses(posts), nil
}

func (s *postService) ListRepliesByHandle(ctx context.Context, handle string, viewerID *uuid.UUID) ([]dto.PostDetailResponse, error) {
	var posts []model.PostWithAuthor
	var err error

	if viewerID != nil {
		posts, err = s.postRepo.FindRepliesByAuthorHandleWithUser(ctx, handle, 50, 0, *viewerID)
	} else {
		posts, err = s.postRepo.FindRepliesByAuthorHandle(ctx, handle, 50, 0)
	}

	if err != nil {
		return nil, apperror.Internal("failed to retrieve user replies")
	}

	return s.toPostDetailResponses(posts), nil
}

func (s *postService) ListLikedPostsByHandle(ctx context.Context, handle string, viewerID *uuid.UUID) ([]dto.PostDetailResponse, error) {
	var posts []model.PostWithAuthor
	var err error

	if viewerID != nil {
		posts, err = s.postRepo.FindLikedByUserHandleWithViewer(ctx, handle, 50, 0, *viewerID)
	} else {
		posts, err = s.postRepo.FindLikedByUserHandle(ctx, handle, 50, 0)
	}

	if err != nil {
		return nil, apperror.Internal("failed to retrieve liked posts")
	}

	return s.toPostDetailResponses(posts), nil
}
