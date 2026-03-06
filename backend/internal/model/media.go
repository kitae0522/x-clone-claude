package model

import (
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeGIF   MediaType = "gif"
)

type Media struct {
	ID              uuid.UUID
	PostID          *uuid.UUID
	UploaderID      uuid.UUID
	URL             string
	MediaType       MediaType
	MimeType        string
	Width           *int
	Height          *int
	SizeBytes       int64
	DurationSeconds *float64
	SortOrder       int16
	CreatedAt       time.Time
}
