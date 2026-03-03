package dto

import "github.com/kitae0522/twitter-clone-claude/backend/internal/model"

type UpdateProfileRequest struct {
	DisplayName     string `json:"displayName"`
	Bio             string `json:"bio"`
	Username        string `json:"username"`
	ProfileImageURL string `json:"profileImageUrl"`
	HeaderImageURL  string `json:"headerImageUrl"`
}

type ProfileResponse struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	Bio             string `json:"bio"`
	ProfileImageURL string `json:"profileImageUrl"`
	HeaderImageURL  string `json:"headerImageUrl"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

func ToProfileResponse(u *model.User) ProfileResponse {
	return ProfileResponse{
		ID:              u.ID.String(),
		Username:        u.Username,
		DisplayName:     u.DisplayName,
		Bio:             u.Bio,
		ProfileImageURL: u.ProfileImageURL,
		HeaderImageURL:  u.HeaderImageURL,
		CreatedAt:       u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
