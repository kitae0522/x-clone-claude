package model

import (
	"time"

	"github.com/google/uuid"
)

type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityFollower Visibility = "follower"
	VisibilityPrivate  Visibility = "private"
)

type Post struct {
	ID           uuid.UUID
	AuthorID     uuid.UUID
	ParentID     *uuid.UUID
	Content      string
	Visibility   Visibility
	LikeCount    int
	ReplyCount   int
	ViewCount    int
	RepostCount  int
	LocationLat  *float64
	LocationLng  *float64
	LocationName *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type PostWithAuthor struct {
	Post
	AuthorUsername        string
	AuthorDisplayName     string
	AuthorProfileImageURL string
	IsLiked               bool
	IsBookmarked          bool
	IsReposted            bool
	RepostedByUsername    *string
	RepostedByDisplayName *string
	RepostedAt            *time.Time
	LocationLat           *float64
	LocationLng           *float64
	LocationName          *string
	// Parent post info (optional, populated for replies in profile context)
	ParentPostID                *uuid.UUID
	ParentContent               *string
	ParentAuthorUsername        *string
	ParentAuthorDisplayName     *string
	ParentAuthorProfileImageURL *string
}
