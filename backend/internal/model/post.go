package model

import (
	"time"

	"github.com/google/uuid"
)

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityFriends Visibility = "friends"
	VisibilityPrivate Visibility = "private"
)

type Post struct {
	ID         uuid.UUID
	AuthorID   uuid.UUID
	Content    string
	Visibility Visibility
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PostWithAuthor struct {
	Post
	AuthorUsername        string
	AuthorDisplayName    string
	AuthorProfileImageURL string
}
