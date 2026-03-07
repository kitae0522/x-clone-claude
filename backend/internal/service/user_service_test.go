package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

func TestGetProfile_Success(t *testing.T) {
	repo := newMockUserRepo()
	followRepo := newMockFollowRepo()
	svc := NewUserService(repo, followRepo)

	authSvc := NewAuthService(repo, testConfig())
	_, _ = authSvc.Register(context.Background(), dto.RegisterRequest{
		Email: "profile@example.com", Username: "profileuser", Password: "password123",
	})

	profile, err := svc.GetProfile(context.Background(), "profileuser", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if profile.Username != "profileuser" {
		t.Errorf("expected username profileuser, got %s", profile.Username)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	followRepo := newMockFollowRepo()
	svc := NewUserService(repo, followRepo)

	_, err := svc.GetProfile(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != 404 {
		t.Errorf("expected 404, got %d", appErr.Code)
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	repo := newMockUserRepo()
	followRepo := newMockFollowRepo()
	svc := NewUserService(repo, followRepo)

	user := &model.User{
		Email:        "update@example.com",
		PasswordHash: "hash",
		Username:     "updateuser",
		DisplayName:  "updateuser",
	}
	_ = repo.Create(context.Background(), user)

	updated, err := svc.UpdateProfile(context.Background(), user.ID, dto.UpdateProfileRequest{
		DisplayName: "New Name",
		Bio:         "Hello world",
		Username:    "updateuser",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.DisplayName != "New Name" {
		t.Errorf("expected displayName 'New Name', got %s", updated.DisplayName)
	}
}

func TestUpdateProfile_DuplicateUsername(t *testing.T) {
	repo := newMockUserRepo()
	followRepo := newMockFollowRepo()
	svc := NewUserService(repo, followRepo)

	user1 := &model.User{
		Email: "user1@example.com", PasswordHash: "hash", Username: "user1", DisplayName: "User 1",
	}
	user2 := &model.User{
		Email: "user2@example.com", PasswordHash: "hash", Username: "user2", DisplayName: "User 2",
	}
	_ = repo.Create(context.Background(), user1)
	_ = repo.Create(context.Background(), user2)

	_, err := svc.UpdateProfile(context.Background(), user2.ID, dto.UpdateProfileRequest{
		DisplayName: "User 2",
		Username:    "user1",
	})
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != 409 {
		t.Errorf("expected 409, got %d", appErr.Code)
	}
}

func TestUpdateProfile_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	followRepo := newMockFollowRepo()
	svc := NewUserService(repo, followRepo)

	_, err := svc.UpdateProfile(context.Background(), uuid.New(), dto.UpdateProfileRequest{
		DisplayName: "Nobody",
		Username:    "nobody",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != 404 {
		t.Errorf("expected 404, got %d", appErr.Code)
	}
}

func TestUpdateProfile_ImageURL(t *testing.T) {
	tests := []struct {
		name                   string
		existingProfileImage   string
		existingHeaderImage    string
		reqProfileImage        string
		reqHeaderImage         string
		expectedProfileImage   string
		expectedHeaderImage    string
	}{
		{
			name:                 "empty image URLs preserve existing values",
			existingProfileImage: "https://example.com/old-profile.jpg",
			existingHeaderImage:  "https://example.com/old-header.jpg",
			reqProfileImage:      "",
			reqHeaderImage:       "",
			expectedProfileImage: "https://example.com/old-profile.jpg",
			expectedHeaderImage:  "https://example.com/old-header.jpg",
		},
		{
			name:                 "relative path /uploads/ is accepted for profile image",
			existingProfileImage: "https://example.com/old-profile.jpg",
			existingHeaderImage:  "https://example.com/old-header.jpg",
			reqProfileImage:      "/uploads/profile-123.jpg",
			reqHeaderImage:       "/uploads/header-456.jpg",
			expectedProfileImage: "/uploads/profile-123.jpg",
			expectedHeaderImage:  "/uploads/header-456.jpg",
		},
		{
			name:                 "absolute URL https is accepted",
			existingProfileImage: "/uploads/old-profile.jpg",
			existingHeaderImage:  "/uploads/old-header.jpg",
			reqProfileImage:      "https://cdn.example.com/profile.jpg",
			reqHeaderImage:       "https://cdn.example.com/header.jpg",
			expectedProfileImage: "https://cdn.example.com/profile.jpg",
			expectedHeaderImage:  "https://cdn.example.com/header.jpg",
		},
		{
			name:                 "only profile image updated, header preserved",
			existingProfileImage: "/uploads/old-profile.jpg",
			existingHeaderImage:  "/uploads/old-header.jpg",
			reqProfileImage:      "/uploads/new-profile.jpg",
			reqHeaderImage:       "",
			expectedProfileImage: "/uploads/new-profile.jpg",
			expectedHeaderImage:  "/uploads/old-header.jpg",
		},
		{
			name:                 "only header image updated, profile preserved",
			existingProfileImage: "/uploads/old-profile.jpg",
			existingHeaderImage:  "/uploads/old-header.jpg",
			reqProfileImage:      "",
			reqHeaderImage:       "/uploads/new-header.jpg",
			expectedProfileImage: "/uploads/old-profile.jpg",
			expectedHeaderImage:  "/uploads/new-header.jpg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newMockUserRepo()
			followRepo := newMockFollowRepo()
			svc := NewUserService(repo, followRepo)

			user := &model.User{
				Email:           "imgtest@example.com",
				PasswordHash:    "hash",
				Username:        "imgtestuser",
				DisplayName:     "Image Test User",
				ProfileImageURL: tc.existingProfileImage,
				HeaderImageURL:  tc.existingHeaderImage,
			}
			_ = repo.Create(context.Background(), user)

			updated, err := svc.UpdateProfile(context.Background(), user.ID, dto.UpdateProfileRequest{
				DisplayName:     "Image Test User",
				Username:        "imgtestuser",
				ProfileImageURL: tc.reqProfileImage,
				HeaderImageURL:  tc.reqHeaderImage,
			})
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if updated.ProfileImageURL != tc.expectedProfileImage {
				t.Errorf("ProfileImageURL: expected %q, got %q", tc.expectedProfileImage, updated.ProfileImageURL)
			}
			if updated.HeaderImageURL != tc.expectedHeaderImage {
				t.Errorf("HeaderImageURL: expected %q, got %q", tc.expectedHeaderImage, updated.HeaderImageURL)
			}
		})
	}
}
