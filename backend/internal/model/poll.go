package model

import (
	"time"

	"github.com/google/uuid"
)

type Poll struct {
	ID         uuid.UUID
	PostID     uuid.UUID
	ExpiresAt  time.Time
	TotalVotes int
	CreatedAt  time.Time
}

type PollOption struct {
	ID          uuid.UUID
	PollID      uuid.UUID
	OptionIndex int16
	Text        string
	VoteCount   int
}

type PollVote struct {
	ID          uuid.UUID
	PollID      uuid.UUID
	UserID      uuid.UUID
	OptionIndex int16
	CreatedAt   time.Time
}
