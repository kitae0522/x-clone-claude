package dto

import "github.com/kitae0522/twitter-clone-claude/backend/internal/model"

type RegisterRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

type UserResponse struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	Bio             string `json:"bio"`
	ProfileImageURL string `json:"profileImageUrl"`
	HeaderImageURL  string `json:"headerImageUrl"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

func ToUserResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:              u.ID.String(),
		Email:           u.Email,
		Username:        u.Username,
		DisplayName:     u.DisplayName,
		Bio:             u.Bio,
		ProfileImageURL: u.ProfileImageURL,
		HeaderImageURL:  u.HeaderImageURL,
		CreatedAt:       u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
