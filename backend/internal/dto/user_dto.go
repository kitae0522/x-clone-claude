package dto

import "github.com/kitae0522/twitter-clone-claude/backend/internal/model"

type UpdateProfileRequest struct {
	DisplayName     string `json:"displayName"     validate:"omitempty,max=50"`
	Bio             string `json:"bio"             validate:"omitempty,max=160"`
	Username        string `json:"username"        validate:"omitempty,min=3,max=30,alphanum"`
	ProfileImageURL string `json:"profileImageUrl" validate:"omitempty,url"`
	HeaderImageURL  string `json:"headerImageUrl"  validate:"omitempty,url"`
}

type ProfileResponse struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	Bio             string `json:"bio"`
	ProfileImageURL string `json:"profileImageUrl"`
	HeaderImageURL  string `json:"headerImageUrl"`
	FollowersCount  int    `json:"followersCount"`
	FollowingCount  int    `json:"followingCount"`
	IsFollowing     bool   `json:"isFollowing"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

func ToProfileResponse(u *model.User, followersCount, followingCount int, isFollowing bool) ProfileResponse {
	return ProfileResponse{
		ID:              u.ID.String(),
		Username:        u.Username,
		DisplayName:     u.DisplayName,
		Bio:             u.Bio,
		ProfileImageURL: u.ProfileImageURL,
		HeaderImageURL:  u.HeaderImageURL,
		FollowersCount:  followersCount,
		FollowingCount:  followingCount,
		IsFollowing:     isFollowing,
		CreatedAt:       u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
