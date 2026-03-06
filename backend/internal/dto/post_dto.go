package dto

import (
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type CreatePostRequest struct {
	Content    string           `json:"content"    validate:"omitempty,max=500"`
	Visibility string           `json:"visibility" validate:"omitempty,oneof=public follower private"`
	MediaIds   []string         `json:"mediaIds"   validate:"omitempty,max=4"`
	Location   *LocationRequest `json:"location"`
	Poll       *PollRequest     `json:"poll"`
}

type LocationRequest struct {
	Latitude  float64 `json:"latitude"  validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	Name      string  `json:"name"      validate:"omitempty,max=100"`
}

type LocationResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name"`
}

type PollRequest struct {
	Options         []string `json:"options"         validate:"required,min=2,max=4,dive,required,max=25"`
	DurationMinutes int      `json:"durationMinutes" validate:"required,min=60,max=10080"`
}

type PollOptionResponse struct {
	Text      string `json:"text"`
	VoteCount int    `json:"voteCount"`
}

type PollResponse struct {
	Options    []PollOptionResponse `json:"options"`
	TotalVotes int                  `json:"totalVotes"`
	VotedIndex int                  `json:"votedIndex"`
	ExpiresAt  string               `json:"expiresAt"`
	IsExpired  bool                 `json:"isExpired"`
}

type VoteRequest struct {
	OptionIndex int `json:"optionIndex" validate:"min=0,max=3"`
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
	ViewCount    int                  `json:"viewCount"`
	IsLiked      bool                 `json:"isLiked"`
	IsBookmarked bool                 `json:"isBookmarked"`
	Location     *LocationResponse    `json:"location,omitempty"`
	Poll         *PollResponse        `json:"poll,omitempty"`
	Media        []MediaResponse      `json:"media,omitempty"`
	TopReplies   []PostDetailResponse `json:"topReplies"`
	CreatedAt    string               `json:"createdAt"`
	UpdatedAt    string               `json:"updatedAt"`
}

type CreateReplyRequest struct {
	Content  string           `json:"content"  validate:"required,min=1,max=500"`
	MediaIds []string         `json:"mediaIds" validate:"omitempty,max=4"`
	Location *LocationRequest `json:"location"`
	Poll     *PollRequest     `json:"poll"`
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
		ViewCount:    p.ViewCount,
		IsLiked:      p.IsLiked,
		IsBookmarked: p.IsBookmarked,
		CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if p.LocationLat != nil && p.LocationLng != nil {
		loc := &LocationResponse{
			Latitude:  *p.LocationLat,
			Longitude: *p.LocationLng,
		}
		if p.LocationName != nil {
			loc.Name = *p.LocationName
		}
		resp.Location = loc
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
