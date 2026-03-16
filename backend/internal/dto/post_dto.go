package dto

import (
	"time"

	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type CreatePostRequest struct {
	Content    string           `json:"content"    validate:"omitempty,max=500"`
	Visibility string           `json:"visibility" validate:"omitempty,oneof=public follower private"`
	MediaIds   []string         `json:"mediaIds"   validate:"omitempty,max=4,dive,required"`
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
	IsDeleted       bool   `json:"isDeleted,omitempty"`
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
	RepostCount  int                  `json:"repostCount"`
	IsLiked      bool                 `json:"isLiked"`
	IsBookmarked bool                 `json:"isBookmarked"`
	IsReposted   bool                 `json:"isReposted"`
	RepostedBy   *RepostedBy          `json:"repostedBy,omitempty"`
	Location     *LocationResponse    `json:"location,omitempty"`
	Poll         *PollResponse        `json:"poll,omitempty"`
	Media        []MediaResponse      `json:"media,omitempty"`
	TopReplies   []PostDetailResponse `json:"topReplies"`
	CreatedAt    string               `json:"createdAt"`
	UpdatedAt    string               `json:"updatedAt"`
}

type RepostedBy struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

type UpdatePostRequest struct {
	Content       *string          `json:"content"        validate:"omitempty,max=500"`
	Visibility    *string          `json:"visibility"     validate:"omitempty,oneof=public follower private"`
	MediaIds      *[]string        `json:"mediaIds"       validate:"omitempty,max=4,dive,required"`
	Location      *LocationRequest `json:"location"`
	ClearLocation bool             `json:"clearLocation"`
	Poll          *PollRequest     `json:"poll"`
	ClearPoll     bool             `json:"clearPoll"`
}

type DeletePostResponse struct {
	Message string `json:"message"`
}

type CreateReplyRequest struct {
	Content  string           `json:"content"  validate:"omitempty,max=500"`
	MediaIds []string         `json:"mediaIds" validate:"omitempty,max=4,dive,required"`
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

	author := PostAuthor{
		Username:        p.AuthorUsername,
		DisplayName:     p.AuthorDisplayName,
		ProfileImageURL: p.AuthorProfileImageURL,
	}
	if p.AuthorDeleted {
		author = PostAuthor{
			Username:    "deleted",
			DisplayName: "탈퇴한 사용자",
			IsDeleted:   true,
		}
	}

	resp := PostDetailResponse{
		ID:           p.ID.String(),
		AuthorID:     p.AuthorID.String(),
		ParentID:     parentID,
		Content:      p.Content,
		Visibility:   string(p.Visibility),
		Author:       author,
		LikeCount:    p.LikeCount,
		ReplyCount:   p.ReplyCount,
		ViewCount:    p.ViewCount,
		RepostCount:  p.RepostCount,
		IsLiked:      p.IsLiked,
		IsBookmarked: p.IsBookmarked,
		IsReposted:   p.IsReposted,
		CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if p.RepostedByUsername != nil {
		resp.RepostedBy = &RepostedBy{
			Username:    *p.RepostedByUsername,
			DisplayName: derefStr(p.RepostedByDisplayName),
		}
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

	if p.ParentPostID != nil && p.ParentContent != nil {
		parentAuthor := PostAuthor{
			Username:        derefStr(p.ParentAuthorUsername),
			DisplayName:     derefStr(p.ParentAuthorDisplayName),
			ProfileImageURL: derefStr(p.ParentAuthorProfileImageURL),
		}
		if p.ParentAuthorUsername == nil || *p.ParentAuthorUsername == "" {
			parentAuthor = PostAuthor{
				Username:    "deleted",
				DisplayName: "탈퇴한 사용자",
				IsDeleted:   true,
			}
		}
		resp.Parent = &ParentPostSummary{
			ID:      p.ParentPostID.String(),
			Content: *p.ParentContent,
			Author:  parentAuthor,
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

// Trash DTOs

const trashRetentionDays = 30

func TrashRetentionDays() int {
	return trashRetentionDays
}

type TrashPostResponse struct {
	ID          string            `json:"id"`
	AuthorID    string            `json:"authorId"`
	ParentID    *string           `json:"parentId"`
	Content     string            `json:"content"`
	Visibility  string            `json:"visibility"`
	Author      PostAuthor        `json:"author"`
	LikeCount   int               `json:"likeCount"`
	ReplyCount  int               `json:"replyCount"`
	ViewCount   int               `json:"viewCount"`
	RepostCount int               `json:"repostCount"`
	Location    *LocationResponse `json:"location,omitempty"`
	Media       []MediaResponse   `json:"media,omitempty"`
	Poll        *PollResponse     `json:"poll,omitempty"`
	CreatedAt   string            `json:"createdAt"`
	DeletedAt   string            `json:"deletedAt"`
	CanRestore  bool              `json:"canRestore"`
}

type TrashListResponse struct {
	Posts      []TrashPostResponse `json:"posts"`
	NextCursor *string             `json:"nextCursor"`
	HasMore    bool                `json:"hasMore"`
}

type RestorePostResponse struct {
	Message string             `json:"message"`
	Post    PostDetailResponse `json:"post"`
}

type PermanentDeleteResponse struct {
	Message string `json:"message"`
}

func ToTrashPostResponse(p model.PostWithAuthor, now time.Time) TrashPostResponse {
	var parentID *string
	if p.ParentID != nil {
		s := p.ParentID.String()
		parentID = &s
	}

	author := PostAuthor{
		Username:        p.AuthorUsername,
		DisplayName:     p.AuthorDisplayName,
		ProfileImageURL: p.AuthorProfileImageURL,
	}
	if p.AuthorDeleted {
		author = PostAuthor{
			Username:    "deleted",
			DisplayName: "탈퇴한 사용자",
			IsDeleted:   true,
		}
	}

	canRestore := true
	deletedAtStr := ""
	if p.DeletedAt != nil {
		deletedAtStr = p.DeletedAt.Format("2006-01-02T15:04:05Z")
		if now.Sub(*p.DeletedAt) > time.Duration(trashRetentionDays)*24*time.Hour {
			canRestore = false
		}
	}

	resp := TrashPostResponse{
		ID:          p.ID.String(),
		AuthorID:    p.AuthorID.String(),
		ParentID:    parentID,
		Content:     p.Content,
		Visibility:  string(p.Visibility),
		Author:      author,
		LikeCount:   p.LikeCount,
		ReplyCount:  p.ReplyCount,
		ViewCount:   p.ViewCount,
		RepostCount: p.RepostCount,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		DeletedAt:   deletedAtStr,
		CanRestore:  canRestore,
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

	return resp
}
