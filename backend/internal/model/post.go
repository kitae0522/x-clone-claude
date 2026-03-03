package model

import "time"

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityFriends Visibility = "friends"
	VisibilityPrivate Visibility = "private"
)

type Post struct {
	ID         string
	AuthorID   string
	Content    string
	Visibility Visibility
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
