package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/mediaclient"
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
	postRepo    repository.PostRepository
	pollRepo    repository.PollRepository
	mediaRepo   repository.MediaRepository
	followRepo  repository.FollowRepository
	mediaClient mediaclient.Client
}

func NewPostService(postRepo repository.PostRepository, pollRepo repository.PollRepository, mediaRepo repository.MediaRepository, followRepo repository.FollowRepository, mc mediaclient.Client) PostService {
	return &postService{
		postRepo:    postRepo,
		pollRepo:    pollRepo,
		mediaRepo:   mediaRepo,
		followRepo:  followRepo,
		mediaClient: mc,
	}
}

func (s *postService) CreatePost(ctx context.Context, authorID uuid.UUID, req dto.CreatePostRequest) (*dto.PostDetailResponse, error) {
	content := req.Content
	hasMedia := len(req.MediaIds) > 0

	if req.Poll != nil && hasMedia {
		return nil, apperror.BadRequest("poll and media cannot be used together")
	}

	if utf8.RuneCountInString(content) == 0 && !hasMedia {
		return nil, apperror.BadRequest("content must not be empty")
	}
	if utf8.RuneCountInString(content) > 500 {
		return nil, apperror.BadRequest("content must not exceed 500 characters")
	}

	visibility := model.VisibilityPublic
	if req.Visibility != "" {
		switch model.Visibility(req.Visibility) {
		case model.VisibilityPublic, model.VisibilityFollower, model.VisibilityPrivate:
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

	if req.Location != nil {
		post.LocationLat = &req.Location.Latitude
		post.LocationLng = &req.Location.Longitude
		if req.Location.Name != "" {
			post.LocationName = &req.Location.Name
		}
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, apperror.Internal("failed to create post")
	}

	if len(req.MediaIds) > 0 {
		if err := s.linkMediaFromService(ctx, req.MediaIds, post.ID, authorID); err != nil {
			slog.Error("failed to link media", "error", err)
		}
	}

	if req.Poll != nil {
		expiresAt := time.Now().Add(time.Duration(req.Poll.DurationMinutes) * time.Minute)
		poll := &model.Poll{
			PostID:    post.ID,
			ExpiresAt: expiresAt,
		}

		var options []model.PollOption
		for i, text := range req.Poll.Options {
			options = append(options, model.PollOption{
				OptionIndex: int16(i),
				Text:        text,
			})
		}

		if err := s.pollRepo.CreatePoll(ctx, poll, options); err != nil {
			return nil, apperror.Internal("failed to create poll")
		}
	}

	result, err := s.postRepo.FindByID(ctx, post.ID)
	if err != nil {
		return nil, apperror.Internal("failed to retrieve created post")
	}

	resp := dto.ToPostDetailResponse(*result)
	_ = s.enrichWithPollAndMedia(ctx, &resp, nil)
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

	if err := s.checkVisibilityAccess(ctx, result, userID); err != nil {
		return nil, err
	}

	if userID == nil || result.AuthorID != *userID {
		if err := s.postRepo.IncrementViewCount(ctx, id); err != nil {
			slog.Error("failed to increment view count", "postID", id, "error", err)
		} else {
			result.ViewCount++
		}
	}

	resp := dto.ToPostDetailResponse(*result)
	_ = s.enrichWithPollAndMedia(ctx, &resp, userID)

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
	s.enrichSlice(ctx, responses, userID)
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
		_ = s.enrichWithPollAndMedia(ctx, &resp, userID)
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
			_ = s.enrichWithPollAndMedia(ctx, &resp, userID)
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
	s.enrichSlice(ctx, responses, userID)
	return responses, nil
}

func (s *postService) CreateReply(ctx context.Context, parentID, authorID uuid.UUID, req dto.CreateReplyRequest) (*dto.PostDetailResponse, error) {
	content := req.Content
	if utf8.RuneCountInString(content) == 0 {
		return nil, apperror.BadRequest("content must not be empty")
	}
	if utf8.RuneCountInString(content) > 500 {
		return nil, apperror.BadRequest("content must not exceed 500 characters")
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

	if len(req.MediaIds) > 0 {
		if err := s.linkMediaFromService(ctx, req.MediaIds, post.ID, authorID); err != nil {
			slog.Error("failed to link media to reply", "error", err)
		}
	}

	if req.Poll != nil {
		expiresAt := time.Now().Add(time.Duration(req.Poll.DurationMinutes) * time.Minute)
		poll := &model.Poll{
			PostID:    post.ID,
			ExpiresAt: expiresAt,
		}

		var options []model.PollOption
		for i, text := range req.Poll.Options {
			options = append(options, model.PollOption{
				OptionIndex: int16(i),
				Text:        text,
			})
		}

		if err := s.pollRepo.CreatePoll(ctx, poll, options); err != nil {
			return nil, apperror.Internal("failed to create poll")
		}
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

func (s *postService) enrichWithPollAndMedia(ctx context.Context, resp *dto.PostDetailResponse, userID *uuid.UUID) error {
	postID, err := uuid.Parse(resp.ID)
	if err != nil {
		return err
	}

	if s.pollRepo != nil {
		poll, options, err := s.pollRepo.FindByPostID(ctx, postID)
		if err == nil && poll != nil {
			pollResp := &dto.PollResponse{
				TotalVotes: poll.TotalVotes,
				ExpiresAt:  poll.ExpiresAt.Format("2006-01-02T15:04:05Z"),
				IsExpired:  time.Now().After(poll.ExpiresAt),
				VotedIndex: -1,
			}
			for _, o := range options {
				pollResp.Options = append(pollResp.Options, dto.PollOptionResponse{
					Text:      o.Text,
					VoteCount: o.VoteCount,
				})
			}
			if userID != nil {
				votedIdx, err := s.pollRepo.GetUserVote(ctx, poll.ID, *userID)
				if err == nil && votedIdx != nil {
					pollResp.VotedIndex = int(*votedIdx)
				}
			}
			resp.Poll = pollResp
		}
	}

	if s.mediaRepo != nil {
		mediaList, err := s.mediaRepo.FindByPostID(ctx, postID)
		if err == nil && len(mediaList) > 0 {
			var mediaResponses []dto.MediaResponse
			for _, m := range mediaList {
				mediaResponses = append(mediaResponses, dto.MediaResponse{
					ID:       m.ID.String(),
					URL:      "/media/" + m.ID.String() + "?size=medium",
					Type:     string(m.MediaType),
					MimeType: m.MimeType,
					Width:    m.Width,
					Height:   m.Height,
					Size:     m.SizeBytes,
					Duration: m.DurationSeconds,
				})
			}
			resp.Media = mediaResponses
		}
	}

	return nil
}

func (s *postService) linkMediaFromService(ctx context.Context, mediaIdStrs []string, postID, uploaderID uuid.UUID) error {
	for i, idStr := range mediaIdStrs {
		mediaID, err := uuid.Parse(idStr)
		if err != nil {
			return fmt.Errorf("invalid media ID %s: %w", idStr, err)
		}

		// Check if already exists in DB (legacy local upload)
		existing, _ := s.mediaRepo.FindByID(ctx, mediaID)
		if existing != nil {
			// Legacy record exists, just link it
			if err := s.mediaRepo.LinkToPost(ctx, []uuid.UUID{mediaID}, postID); err != nil {
				return fmt.Errorf("link existing media: %w", err)
			}
			continue
		}

		// Fetch metadata from media-service
		status, err := s.mediaClient.GetStatus(ctx, idStr)
		if err != nil {
			return fmt.Errorf("get media status %s: %w", idStr, err)
		}

		var w, h *int
		if status.Width > 0 {
			w = &status.Width
		}
		if status.Height > 0 {
			h = &status.Height
		}

		media := &model.Media{
			ID:         mediaID,
			PostID:     &postID,
			UploaderID: uploaderID,
			URL:        "/media/" + idStr,
			MediaType:  model.MediaType(status.MediaType),
			MimeType:   status.MimeType,
			SizeBytes:  status.Size,
			Width:      w,
			Height:     h,
			SortOrder:  int16(i),
		}

		if err := s.mediaRepo.Create(ctx, media); err != nil {
			return fmt.Errorf("create media record %s: %w", idStr, err)
		}
	}
	return nil
}

func (s *postService) enrichSlice(ctx context.Context, responses []dto.PostDetailResponse, userID *uuid.UUID) {
	for i := range responses {
		_ = s.enrichWithPollAndMedia(ctx, &responses[i], userID)
	}
}

func (s *postService) checkVisibilityAccess(ctx context.Context, post *model.PostWithAuthor, viewerID *uuid.UUID) error {
	switch post.Visibility {
	case model.VisibilityPublic:
		return nil
	case model.VisibilityFollower:
		if viewerID == nil {
			return apperror.NotFound("post not found")
		}
		if post.AuthorID == *viewerID {
			return nil
		}
		isFollowing, err := s.followRepo.IsFollowing(ctx, *viewerID, post.AuthorID)
		if err != nil {
			return apperror.Internal("failed to check follow relationship")
		}
		if !isFollowing {
			return apperror.NotFound("post not found")
		}
		return nil
	case model.VisibilityPrivate:
		if viewerID == nil || post.AuthorID != *viewerID {
			return apperror.NotFound("post not found")
		}
		return nil
	default:
		return nil
	}
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

	responses := s.toPostDetailResponses(posts)
	s.enrichSlice(ctx, responses, viewerID)
	return responses, nil
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

	responses := s.toPostDetailResponses(posts)
	s.enrichSlice(ctx, responses, viewerID)
	return responses, nil
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

	responses := s.toPostDetailResponses(posts)
	s.enrichSlice(ctx, responses, viewerID)
	return responses, nil
}
