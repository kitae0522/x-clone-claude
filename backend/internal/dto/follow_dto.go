package dto

import "github.com/kitae0522/twitter-clone-claude/backend/internal/model"

type FollowStatusResponse struct {
	Following bool `json:"following"`
}

type FollowUserResponse struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	Bio             string `json:"bio"`
	ProfileImageURL string `json:"profileImageUrl"`
}

type FollowListResponse struct {
	Users []FollowUserResponse `json:"users"`
	Total int                  `json:"total"`
}

func ToFollowUserResponse(u *model.User) FollowUserResponse {
	return FollowUserResponse{
		ID:              u.ID.String(),
		Username:        u.Username,
		DisplayName:     u.DisplayName,
		Bio:             u.Bio,
		ProfileImageURL: u.ProfileImageURL,
	}
}
