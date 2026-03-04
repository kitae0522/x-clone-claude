package model

import (
	"time"

	"github.com/google/uuid"
)

type Like struct {
	UserID    uuid.UUID
	PostID    uuid.UUID
	CreatedAt time.Time
}
