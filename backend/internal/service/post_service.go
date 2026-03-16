package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
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
	UpdatePost(ctx context.Context, postID, requesterID uuid.UUID, req dto.UpdatePostRequest) (*dto.PostDetailResponse, error)
	DeletePost(ctx context.Context, postID, requesterID uuid.UUID) error
	ListTrash(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) (*dto.TrashListResponse, error)
	RestorePost(ctx context.Context, postID, requesterID uuid.UUID) (*dto.PostDetailResponse, error)
	PermanentDeletePost(ctx context.Context, postID, requesterID uuid.UUID) error
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
	content := strings.TrimSpace(req.Content)
	hasMedia := len(req.MediaIds) > 0

	if req.Poll != nil && hasMedia {
		return nil, apperror.BadRequest("poll and media cannot be used together")
	}

	if len(req.MediaIds) > 4 {
		return nil, apperror.BadRequest("media must not exceed 4 items")
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
			exists, isDeleted, checkErr := s.postRepo.ExistsIncludingDeleted(ctx, id)
			if checkErr == nil && exists && isDeleted {
				return nil, apperror.Gone("this post has been deleted")
			}
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

	s.incrementViewCounts(ctx, replies, userID)

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
	content := strings.TrimSpace(req.Content)
	hasMedia := len(req.MediaIds) > 0

	if req.Poll != nil && hasMedia {
		return nil, apperror.BadRequest("poll and media cannot be used together")
	}

	if len(req.MediaIds) > 4 {
		return nil, apperror.BadRequest("media must not exceed 4 items")
	}

	if utf8.RuneCountInString(content) == 0 && !hasMedia {
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

	if req.Location != nil {
		post.LocationLat = &req.Location.Latitude
		post.LocationLng = &req.Location.Longitude
		if req.Location.Name != "" {
			post.LocationName = &req.Location.Name
		}
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
	_ = s.enrichWithPollAndMedia(ctx, &resp, nil)
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

	s.incrementViewCounts(ctx, replies, userID)

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

func (s *postService) incrementViewCounts(ctx context.Context, posts []model.PostWithAuthor, userID *uuid.UUID) {
	var ids []uuid.UUID
	for i := range posts {
		if userID == nil || posts[i].AuthorID != *userID {
			ids = append(ids, posts[i].ID)
		}
	}
	if len(ids) == 0 {
		return
	}
	if err := s.postRepo.IncrementViewCountBatch(ctx, ids); err != nil {
		slog.Error("failed to batch increment view count", "error", err)
		return
	}
	for i := range posts {
		if userID == nil || posts[i].AuthorID != *userID {
			posts[i].ViewCount++
		}
	}
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

func (s *postService) UpdatePost(ctx context.Context, postID, requesterID uuid.UUID, req dto.UpdatePostRequest) (*dto.PostDetailResponse, error) {
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Internal("failed to retrieve post")
	}

	if post.AuthorID != requesterID {
		return nil, apperror.Forbidden("you can only edit your own post")
	}

	content := post.Content
	if req.Content != nil {
		content = strings.TrimSpace(*req.Content)
	}

	hasMedia := false
	if s.mediaRepo != nil {
		existingMedia, _ := s.mediaRepo.FindByPostID(ctx, postID)
		hasMedia = len(existingMedia) > 0
	}
	if req.MediaIds != nil {
		if len(*req.MediaIds) > 4 {
			return nil, apperror.BadRequest("media must not exceed 4 items")
		}
		hasMedia = len(*req.MediaIds) > 0
	}

	if req.Poll != nil && req.MediaIds != nil && len(*req.MediaIds) > 0 {
		return nil, apperror.BadRequest("poll and media cannot be used together")
	}

	if utf8.RuneCountInString(content) == 0 && !hasMedia {
		return nil, apperror.BadRequest("content must not be empty")
	}
	if utf8.RuneCountInString(content) > 500 {
		return nil, apperror.BadRequest("content must not exceed 500 characters")
	}

	visibility := post.Visibility
	if req.Visibility != nil && post.ParentID == nil {
		switch model.Visibility(*req.Visibility) {
		case model.VisibilityPublic, model.VisibilityFollower, model.VisibilityPrivate:
			visibility = model.Visibility(*req.Visibility)
		default:
			return nil, apperror.BadRequest("invalid visibility: %s", *req.Visibility)
		}
	}

	locationLat := post.LocationLat
	locationLng := post.LocationLng
	locationName := post.LocationName

	if req.ClearLocation {
		locationLat = nil
		locationLng = nil
		locationName = nil
	} else if req.Location != nil {
		locationLat = &req.Location.Latitude
		locationLng = &req.Location.Longitude
		if req.Location.Name != "" {
			locationName = &req.Location.Name
		} else {
			locationName = nil
		}
	}

	if err := s.postRepo.Update(ctx, postID, content, visibility, locationLat, locationLng, locationName); err != nil {
		return nil, apperror.Internal("failed to update post")
	}

	if req.ClearPoll {
		if s.pollRepo != nil {
			if err := s.pollRepo.DeleteByPostID(ctx, postID); err != nil {
				slog.Error("failed to delete poll", "error", err)
			}
		}
	} else if req.Poll != nil {
		if s.pollRepo != nil {
			_ = s.pollRepo.DeleteByPostID(ctx, postID)

			expiresAt := time.Now().Add(time.Duration(req.Poll.DurationMinutes) * time.Minute)
			poll := &model.Poll{
				PostID:    postID,
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
				slog.Error("failed to create poll on update", "error", err)
			}
		}
	}

	if req.MediaIds != nil {
		if s.mediaRepo != nil {
			_ = s.mediaRepo.UnlinkByPostID(ctx, postID)
		}
		if len(*req.MediaIds) > 0 {
			if err := s.linkMediaFromService(ctx, *req.MediaIds, postID, requesterID); err != nil {
				slog.Error("failed to link media on update", "error", err)
			}
		}
	}

	result, err := s.postRepo.FindByIDWithUser(ctx, postID, requesterID)
	if err != nil {
		return nil, apperror.Internal("failed to retrieve updated post")
	}

	resp := dto.ToPostDetailResponse(*result)
	_ = s.enrichWithPollAndMedia(ctx, &resp, &requesterID)
	return &resp, nil
}

func (s *postService) DeletePost(ctx context.Context, postID, requesterID uuid.UUID) error {
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return apperror.NotFound("post not found")
		}
		return apperror.Internal("failed to retrieve post")
	}

	isReply := post.ParentID != nil
	isPostAuthor := post.AuthorID == requesterID

	if isReply {
		var isParentAuthor bool
		parent, parentErr := s.postRepo.FindByID(ctx, *post.ParentID)
		if parentErr == nil && parent != nil {
			isParentAuthor = parent.AuthorID == requesterID
		}
		if !isPostAuthor && !isParentAuthor {
			return apperror.Forbidden("you can only delete your own reply or replies on your post")
		}
	} else {
		if !isPostAuthor {
			return apperror.Forbidden("you can only delete your own post")
		}
	}

	// Clean up associated poll and media records
	if s.pollRepo != nil {
		if err := s.pollRepo.DeleteByPostID(ctx, postID); err != nil {
			slog.Error("failed to delete poll on post delete", "error", err)
		}
	}
	if s.mediaRepo != nil {
		if err := s.mediaRepo.UnlinkByPostID(ctx, postID); err != nil {
			slog.Error("failed to unlink media on post delete", "error", err)
		}
	}

	if isReply {
		if err := s.postRepo.SoftDeleteReply(ctx, postID, *post.ParentID); err != nil {
			return apperror.Internal("failed to delete reply")
		}
		return nil
	}
	if err := s.postRepo.SoftDelete(ctx, postID); err != nil {
		return apperror.Internal("failed to delete post")
	}
	return nil
}

func (s *postService) ListTrash(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) (*dto.TrashListResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	posts, err := s.postRepo.FindDeletedByAuthor(ctx, userID, limit+1, cursor)
	if err != nil {
		return nil, apperror.Internal("failed to retrieve trash")
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	now := time.Now()
	items := make([]dto.TrashPostResponse, len(posts))
	for i, p := range posts {
		items[i] = dto.ToTrashPostResponse(p, now)
	}

	var nextCursor *string
	if hasMore && len(posts) > 0 {
		last := posts[len(posts)-1]
		if last.DeletedAt != nil {
			c := last.DeletedAt.Format(time.RFC3339)
			nextCursor = &c
		}
	}

	return &dto.TrashListResponse{
		Posts:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *postService) RestorePost(ctx context.Context, postID, requesterID uuid.UUID) (*dto.PostDetailResponse, error) {
	post, err := s.postRepo.FindByIDIncludingDeleted(ctx, postID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Internal("failed to retrieve post")
	}

	if post.DeletedAt == nil {
		return nil, apperror.BadRequest("post is not deleted")
	}

	if post.AuthorID != requesterID {
		return nil, apperror.Forbidden("you can only restore your own post")
	}

	if time.Since(*post.DeletedAt) > time.Duration(dto.TrashRetentionDays())*24*time.Hour {
		return nil, apperror.BadRequest("post cannot be restored after 30 days")
	}

	if post.ParentID != nil {
		parentExists, parentDeleted, checkErr := s.postRepo.ExistsIncludingDeleted(ctx, *post.ParentID)
		if checkErr != nil {
			return nil, apperror.Internal("failed to check parent post")
		}
		if !parentExists || parentDeleted {
			return nil, apperror.BadRequest("cannot restore reply: parent post is deleted")
		}

		if err := s.postRepo.RestoreReply(ctx, postID, *post.ParentID); err != nil {
			return nil, apperror.Internal("failed to restore reply")
		}
	} else {
		if err := s.postRepo.Restore(ctx, postID); err != nil {
			return nil, apperror.Internal("failed to restore post")
		}
	}

	result, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, apperror.Internal("failed to retrieve restored post")
	}

	resp := dto.ToPostDetailResponse(*result)
	_ = s.enrichWithPollAndMedia(ctx, &resp, &requesterID)
	return &resp, nil
}

func (s *postService) PermanentDeletePost(ctx context.Context, postID, requesterID uuid.UUID) error {
	post, err := s.postRepo.FindByIDIncludingDeleted(ctx, postID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return apperror.NotFound("post not found")
		}
		return apperror.Internal("failed to retrieve post")
	}

	if post.DeletedAt == nil {
		return apperror.BadRequest("post is not in trash")
	}

	if post.AuthorID != requesterID {
		return apperror.Forbidden("you can only permanently delete your own post")
	}

	if err := s.postRepo.HardDelete(ctx, postID); err != nil {
		return apperror.Internal("failed to permanently delete post")
	}

	return nil
}
