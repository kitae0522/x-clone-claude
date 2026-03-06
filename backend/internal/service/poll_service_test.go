package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

// --- Mock implementations for PollService ---

type mockPollRepoForVote struct {
	poll      *model.Poll
	options   []model.PollOption
	userVote  *int16
	voteCalled bool
	voteErr   error
}

func (m *mockPollRepoForVote) CreatePoll(_ context.Context, _ *model.Poll, _ []model.PollOption) error {
	return nil
}

func (m *mockPollRepoForVote) FindByPostID(_ context.Context, _ uuid.UUID) (*model.Poll, []model.PollOption, error) {
	return m.poll, m.options, nil
}

func (m *mockPollRepoForVote) Vote(_ context.Context, _, _ uuid.UUID, _ int16) error {
	m.voteCalled = true
	return m.voteErr
}

func (m *mockPollRepoForVote) GetUserVote(_ context.Context, _, _ uuid.UUID) (*int16, error) {
	return m.userVote, nil
}

func (m *mockPollRepoForVote) FindByPostIDs(_ context.Context, _ []uuid.UUID) (map[uuid.UUID]model.Poll, map[uuid.UUID][]model.PollOption, error) {
	return nil, nil, nil
}

type mockPostRepoForPoll struct {
	post *model.PostWithAuthor
	err  error
}

func (m *mockPostRepoForPoll) Create(_ context.Context, _ *model.Post) error { return nil }
func (m *mockPostRepoForPoll) FindByID(_ context.Context, _ uuid.UUID) (*model.PostWithAuthor, error) {
	if m.post == nil {
		return nil, pgx.ErrNoRows
	}
	return m.post, m.err
}
func (m *mockPostRepoForPoll) FindAll(_ context.Context, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForPoll) FindByIDWithUser(_ context.Context, _, _ uuid.UUID) (*model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForPoll) FindAllWithUser(_ context.Context, _, _ int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForPoll) CreateReply(_ context.Context, _ *model.Post) error { return nil }
func (m *mockPostRepoForPoll) FindRepliesByPostID(_ context.Context, _ uuid.UUID, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForPoll) FindRepliesByPostIDWithUser(_ context.Context, _, _ uuid.UUID, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForPoll) FindAuthorReplyByPostID(_ context.Context, _, _ uuid.UUID) (*model.PostWithAuthor, error) {
	return nil, pgx.ErrNoRows
}
func (m *mockPostRepoForPoll) FindAuthorReplyByPostIDWithUser(_ context.Context, _, _, _ uuid.UUID) (*model.PostWithAuthor, error) {
	return nil, pgx.ErrNoRows
}
func (m *mockPostRepoForPoll) FindByAuthorHandle(_ context.Context, _ string, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForPoll) FindByAuthorHandleWithUser(_ context.Context, _ string, _, _ int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForPoll) FindRepliesByAuthorHandle(_ context.Context, _ string, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForPoll) FindRepliesByAuthorHandleWithUser(_ context.Context, _ string, _, _ int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForPoll) FindLikedByUserHandle(_ context.Context, _ string, _, _ int) ([]model.PostWithAuthor, error) {
	return nil, nil
}
func (m *mockPostRepoForPoll) FindLikedByUserHandleWithViewer(_ context.Context, _ string, _, _ int, _ uuid.UUID) ([]model.PostWithAuthor, error) {
	return nil, nil
}

// --- Tests ---

func TestPollService_Vote(t *testing.T) {
	postID := uuid.New()
	pollID := uuid.New()
	authorID := uuid.New()
	voterID := uuid.New()

	basePoll := &model.Poll{
		ID:         pollID,
		PostID:     postID,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		TotalVotes: 0,
	}

	baseOptions := []model.PollOption{
		{ID: uuid.New(), PollID: pollID, OptionIndex: 0, Text: "Option A", VoteCount: 0},
		{ID: uuid.New(), PollID: pollID, OptionIndex: 1, Text: "Option B", VoteCount: 0},
	}

	basePost := &model.PostWithAuthor{
		Post: model.Post{
			ID:       postID,
			AuthorID: authorID,
		},
		AuthorUsername: "pollauthor",
	}

	existingVote := int16(0)

	tests := []struct {
		name        string
		voterID     uuid.UUID
		optionIndex int16
		poll        *model.Poll
		options     []model.PollOption
		post        *model.PostWithAuthor
		userVote    *int16
		wantErr     bool
		wantCode    int
		wantMsg     string
	}{
		{
			name:        "successful vote",
			voterID:     voterID,
			optionIndex: 0,
			poll:        basePoll,
			options:     baseOptions,
			post:        basePost,
			userVote:    nil,
			wantErr:     false,
		},
		{
			name:        "post not found",
			voterID:     voterID,
			optionIndex: 0,
			poll:        basePoll,
			options:     baseOptions,
			post:        nil,
			userVote:    nil,
			wantErr:     true,
			wantCode:    404,
			wantMsg:     "post not found",
		},
		{
			name:        "expired poll",
			voterID:     voterID,
			optionIndex: 0,
			poll: &model.Poll{
				ID:         pollID,
				PostID:     postID,
				ExpiresAt:  time.Now().Add(-1 * time.Hour),
				TotalVotes: 0,
			},
			options:  baseOptions,
			post:     basePost,
			userVote: nil,
			wantErr:  true,
			wantCode: 400,
			wantMsg:  "poll has expired",
		},
		{
			name:        "duplicate vote",
			voterID:     voterID,
			optionIndex: 0,
			poll:        basePoll,
			options:     baseOptions,
			post:        basePost,
			userVote:    &existingVote,
			wantErr:     true,
			wantCode:    409,
			wantMsg:     "already voted on this poll",
		},
		{
			name:        "invalid option index too high",
			voterID:     voterID,
			optionIndex: 5,
			poll:        basePoll,
			options:     baseOptions,
			post:        basePost,
			userVote:    nil,
			wantErr:     true,
			wantCode:    400,
			wantMsg:     "invalid option index",
		},
		{
			name:        "cannot vote on own poll",
			voterID:     authorID,
			optionIndex: 0,
			poll:        basePoll,
			options:     baseOptions,
			post:        basePost,
			userVote:    nil,
			wantErr:     true,
			wantCode:    400,
			wantMsg:     "cannot vote on your own poll",
		},
		{
			name:        "poll not found for post",
			voterID:     voterID,
			optionIndex: 0,
			poll:        nil,
			options:     nil,
			post:        basePost,
			userVote:    nil,
			wantErr:     true,
			wantCode:    404,
			wantMsg:     "poll not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pollRepo := &mockPollRepoForVote{
				poll:     tt.poll,
				options:  tt.options,
				userVote: tt.userVote,
			}
			postRepo := &mockPostRepoForPoll{
				post: tt.post,
			}

			svc := NewPollService(pollRepo, postRepo)

			_, err := svc.Vote(context.Background(), postID, tt.voterID, tt.optionIndex)

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
				if tt.wantMsg != "" {
					if got := appErr.Message; got != tt.wantMsg {
						// Allow partial match for messages with format args
						if len(got) < len(tt.wantMsg) || got[:len(tt.wantMsg)] != tt.wantMsg {
							t.Errorf("expected message containing %q, got %q", tt.wantMsg, got)
						}
					}
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestPollService_Vote_SuccessReturnsResponse(t *testing.T) {
	postID := uuid.New()
	pollID := uuid.New()
	authorID := uuid.New()
	voterID := uuid.New()

	poll := &model.Poll{
		ID:         pollID,
		PostID:     postID,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		TotalVotes: 1,
	}

	options := []model.PollOption{
		{ID: uuid.New(), PollID: pollID, OptionIndex: 0, Text: "Yes", VoteCount: 1},
		{ID: uuid.New(), PollID: pollID, OptionIndex: 1, Text: "No", VoteCount: 0},
	}

	post := &model.PostWithAuthor{
		Post: model.Post{
			ID:       postID,
			AuthorID: authorID,
		},
	}

	pollRepo := &mockPollRepoForVote{
		poll:     poll,
		options:  options,
		userVote: nil,
	}
	postRepo := &mockPostRepoForPoll{post: post}

	svc := NewPollService(pollRepo, postRepo)

	resp, err := svc.Vote(context.Background(), postID, voterID, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.Options) != 2 {
		t.Errorf("expected 2 options, got %d", len(resp.Options))
	}
	if resp.Options[0].Text != "Yes" {
		t.Errorf("expected first option text 'Yes', got %q", resp.Options[0].Text)
	}
	if resp.VotedIndex != 0 {
		t.Errorf("expected votedIndex 0, got %d", resp.VotedIndex)
	}
	if resp.TotalVotes != 1 {
		t.Errorf("expected totalVotes 1, got %d", resp.TotalVotes)
	}
	if resp.IsExpired {
		t.Error("expected IsExpired to be false")
	}

	if !pollRepo.voteCalled {
		t.Error("expected Vote to be called on poll repo")
	}
}
