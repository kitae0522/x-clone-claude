package model

import (
	"time"

	"github.com/google/uuid"
)

type Follow struct {
	FollowerID  uuid.UUID
	FollowingID uuid.UUID
	CreatedAt   time.Time
}
