package dto

import (
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type PostResponse struct {
	ID         string `json:"id"`
	AuthorID   string `json:"authorId"`
	Content    string `json:"content"`
	Visibility string `json:"visibility"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

func ToPostResponse(p model.Post) PostResponse {
	return PostResponse{
		ID:         p.ID,
		AuthorID:   p.AuthorID,
		Content:    p.Content,
		Visibility: string(p.Visibility),
		CreatedAt:  p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
