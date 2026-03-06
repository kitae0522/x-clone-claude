package dto

type MediaResponse struct {
	ID       string   `json:"id"`
	URL      string   `json:"url"`
	Type     string   `json:"type"`
	MimeType string   `json:"mimeType"`
	Width    *int     `json:"width"`
	Height   *int     `json:"height"`
	Size     int64    `json:"size"`
	Duration *float64 `json:"duration"`
}
