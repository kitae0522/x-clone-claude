package dto

import (
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type CreatePostRequest struct {
	Content    string `json:"content"`
	Visibility string `json:"visibility"`
}

type PostAuthor struct {
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	ProfileImageURL string `json:"profileImageUrl"`
}

type PostResponse struct {
	ID         string `json:"id"`
	AuthorID   string `json:"authorId"`
	Content    string `json:"content"`
	Visibility string `json:"visibility"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type PostDetailResponse struct {
	ID         string               `json:"id"`
	AuthorID   string               `json:"authorId"`
	ParentID   *string              `json:"parentId"`
	Content    string               `json:"content"`
	Visibility string               `json:"visibility"`
	Author     PostAuthor           `json:"author"`
	LikeCount  int                  `json:"likeCount"`
	ReplyCount int                  `json:"replyCount"`
	IsLiked    bool                 `json:"isLiked"`
	TopReplies []PostDetailResponse `json:"topReplies"`
	CreatedAt  string               `json:"createdAt"`
	UpdatedAt  string               `json:"updatedAt"`
}

type CreateReplyRequest struct {
	Content string `json:"content"`
}

func ToPostResponse(p model.Post) PostResponse {
	return PostResponse{
		ID:         p.ID.String(),
		AuthorID:   p.AuthorID.String(),
		Content:    p.Content,
		Visibility: string(p.Visibility),
		CreatedAt:  p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToPostDetailResponse(p model.PostWithAuthor) PostDetailResponse {
	var parentID *string
	if p.ParentID != nil {
		s := p.ParentID.String()
		parentID = &s
	}

	return PostDetailResponse{
		ID:         p.ID.String(),
		AuthorID:   p.AuthorID.String(),
		ParentID:   parentID,
		Content:    p.Content,
		Visibility: string(p.Visibility),
		Author: PostAuthor{
			Username:        p.AuthorUsername,
			DisplayName:     p.AuthorDisplayName,
			ProfileImageURL: p.AuthorProfileImageURL,
		},
		LikeCount:  p.LikeCount,
		ReplyCount: p.ReplyCount,
		IsLiked:    p.IsLiked,
		CreatedAt:  p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
