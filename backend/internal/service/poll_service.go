package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/repository"
)

type PollService interface {
	Vote(ctx context.Context, postID, userID uuid.UUID, optionIndex int16) (*dto.PollResponse, error)
}

type pollService struct {
	pollRepo repository.PollRepository
	postRepo repository.PostRepository
}

func NewPollService(pollRepo repository.PollRepository, postRepo repository.PostRepository) PollService {
	return &pollService{
		pollRepo: pollRepo,
		postRepo: postRepo,
	}
}

func (s *pollService) Vote(ctx context.Context, postID, userID uuid.UUID, optionIndex int16) (*dto.PollResponse, error) {
	post, err := s.postRepo.FindByID(ctx, postID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Internal("failed to find post")
	}

	if post.AuthorID == userID {
		return nil, apperror.BadRequest("cannot vote on your own poll")
	}

	poll, options, err := s.pollRepo.FindByPostID(ctx, postID)
	if err != nil {
		return nil, apperror.Internal("failed to find poll")
	}
	if poll == nil {
		return nil, apperror.NotFound("poll not found")
	}

	now := time.Now()
	if now.After(poll.ExpiresAt) {
		return nil, apperror.BadRequest("poll has expired")
	}

	if int(optionIndex) >= len(options) {
		return nil, apperror.BadRequest("invalid option index: %d", optionIndex)
	}

	existingVote, err := s.pollRepo.GetUserVote(ctx, poll.ID, userID)
	if err != nil {
		return nil, apperror.Internal("failed to check existing vote")
	}
	if existingVote != nil {
		return nil, apperror.Conflict("already voted on this poll")
	}

	if err := s.pollRepo.Vote(ctx, poll.ID, userID, optionIndex); err != nil {
		return nil, apperror.Internal("failed to vote")
	}

	poll, options, err = s.pollRepo.FindByPostID(ctx, postID)
	if err != nil {
		return nil, apperror.Internal("failed to retrieve updated poll")
	}

	resp := buildPollResponse(poll, options, &optionIndex)
	return resp, nil
}

func buildPollResponse(poll *model.Poll, options []model.PollOption, votedIndex *int16) *dto.PollResponse {
	optionResponses := make([]dto.PollOptionResponse, len(options))
	for i, opt := range options {
		optionResponses[i] = dto.PollOptionResponse{
			Text:      opt.Text,
			VoteCount: opt.VoteCount,
		}
	}

	voted := -1
	if votedIndex != nil {
		voted = int(*votedIndex)
	}

	now := time.Now()
	return &dto.PollResponse{
		Options:    optionResponses,
		TotalVotes: poll.TotalVotes,
		VotedIndex: voted,
		ExpiresAt:  poll.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		IsExpired:  now.After(poll.ExpiresAt),
	}
}
