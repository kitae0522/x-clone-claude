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
	ParentID   *uuid.UUID
	Content    string
	Visibility Visibility
	LikeCount  int
	ReplyCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PostWithAuthor struct {
	Post
	AuthorUsername        string
	AuthorDisplayName     string
	AuthorProfileImageURL string
	IsLiked               bool
	// Parent post info (optional, populated for replies in profile context)
	ParentPostID               *uuid.UUID
	ParentContent              *string
	ParentAuthorUsername        *string
	ParentAuthorDisplayName     *string
	ParentAuthorProfileImageURL *string
}
