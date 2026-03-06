package model

import "time"

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusReady      Status = "ready"
	StatusFailed     Status = "failed"
)

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeGIF   MediaType = "gif"
)

type SizeVariant string

const (
	SizeSmall    SizeVariant = "small"
	SizeMedium   SizeVariant = "medium"
	SizeLarge    SizeVariant = "large"
	SizeOriginal SizeVariant = "original"
)

var VariantMaxWidth = map[SizeVariant]int{
	SizeSmall:    320,
	SizeMedium:   768,
	SizeLarge:    1440,
	SizeOriginal: 2560,
}

var AllVariants = []SizeVariant{SizeSmall, SizeMedium, SizeLarge, SizeOriginal}

type Media struct {
	ID        string
	Status    Status
	MediaType MediaType
	MimeType  string
	Size      int64
	Width     int
	Height    int
	Error     string
	CreatedAt time.Time
}

type UploadResponse struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
}

type StatusResponse struct {
	ID        string    `json:"id"`
	Status    Status    `json:"status"`
	MediaType MediaType `json:"mediaType,omitempty"`
	MimeType  string    `json:"mimeType,omitempty"`
	Width     int       `json:"width,omitempty"`
	Height    int       `json:"height,omitempty"`
	Size      int64     `json:"size,omitempty"`
	URL       string    `json:"url,omitempty"`
	Error     string    `json:"error,omitempty"`
}
