package dto

import (
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type CreatePostRequest struct {
	Content    string `json:"content"    validate:"required,min=1,max=280"`
	Visibility string `json:"visibility" validate:"omitempty,oneof=public friends private"`
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

type ParentPostSummary struct {
	ID      string     `json:"id"`
	Content string     `json:"content"`
	Author  PostAuthor `json:"author"`
}

type PostDetailResponse struct {
	ID           string               `json:"id"`
	AuthorID     string               `json:"authorId"`
	ParentID     *string              `json:"parentId"`
	Parent       *ParentPostSummary   `json:"parent,omitempty"`
	Content      string               `json:"content"`
	Visibility   string               `json:"visibility"`
	Author       PostAuthor           `json:"author"`
	LikeCount    int                  `json:"likeCount"`
	ReplyCount   int                  `json:"replyCount"`
	IsLiked      bool                 `json:"isLiked"`
	IsBookmarked bool                 `json:"isBookmarked"`
	TopReplies   []PostDetailResponse `json:"topReplies"`
	CreatedAt    string               `json:"createdAt"`
	UpdatedAt    string               `json:"updatedAt"`
}

type CreateReplyRequest struct {
	Content string `json:"content" validate:"required,min=1,max=280"`
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

	resp := PostDetailResponse{
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
		LikeCount:    p.LikeCount,
		ReplyCount:   p.ReplyCount,
		IsLiked:      p.IsLiked,
		IsBookmarked: p.IsBookmarked,
		CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if p.ParentPostID != nil && p.ParentContent != nil && p.ParentAuthorUsername != nil {
		resp.Parent = &ParentPostSummary{
			ID:      p.ParentPostID.String(),
			Content: *p.ParentContent,
			Author: PostAuthor{
				Username:        *p.ParentAuthorUsername,
				DisplayName:     derefStr(p.ParentAuthorDisplayName),
				ProfileImageURL: derefStr(p.ParentAuthorProfileImageURL),
			},
		}
	}

	return resp
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
