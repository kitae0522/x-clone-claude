package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type mockLikeRepoForUser struct {
	softDeleteErr   error
	softDeleteCount int64
	softDeleteCalls int
}

func newMockLikeRepoForUser() *mockLikeRepoForUser {
	return &mockLikeRepoForUser{}
}

func (m *mockLikeRepoForUser) Like(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockLikeRepoForUser) Unlike(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockLikeRepoForUser) IsLiked(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockLikeRepoForUser) SoftDeleteByUserID(_ context.Context, _ uuid.UUID) (int64, error) {
	m.softDeleteCalls++
	return m.softDeleteCount, m.softDeleteErr
}

func TestGetProfile_Success(t *testing.T) {
	repo := newMockUserRepo()
	followRepo := newMockFollowRepo()
	svc := NewUserService(repo, followRepo, newMockLikeRepoForUser())

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
	svc := NewUserService(repo, followRepo, newMockLikeRepoForUser())

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
	svc := NewUserService(repo, followRepo, newMockLikeRepoForUser())

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
	svc := NewUserService(repo, followRepo, newMockLikeRepoForUser())

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
	svc := NewUserService(repo, followRepo, newMockLikeRepoForUser())

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

func TestDeleteAccount(t *testing.T) {
	password := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name            string
		setupUser       bool
		password        string
		likeDeleteErr   error
		likeDeleteCount int64
		wantCode        int
		wantLikeCalls   int
	}{
		{
			name:            "success - soft deletes likes then user",
			setupUser:       true,
			password:        password,
			likeDeleteCount: 5,
			wantCode:        0,
			wantLikeCalls:   1,
		},
		{
			name:      "wrong password",
			setupUser: true,
			password:  "wrongpassword",
			wantCode:  401,
		},
		{
			name:     "user not found",
			password: password,
			wantCode: 404,
		},
		{
			name:          "like soft delete fails",
			setupUser:     true,
			password:      password,
			likeDeleteErr: fmt.Errorf("db error"),
			wantCode:      500,
			wantLikeCalls: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newMockUserRepo()
			followRepo := newMockFollowRepo()
			likeRepo := newMockLikeRepoForUser()
			likeRepo.softDeleteErr = tc.likeDeleteErr
			likeRepo.softDeleteCount = tc.likeDeleteCount

			svc := NewUserService(repo, followRepo, likeRepo)

			var userID uuid.UUID
			if tc.setupUser {
				user := &model.User{
					Email:        "delete@example.com",
					PasswordHash: string(hashed),
					Username:     "deleteuser",
					DisplayName:  "Delete User",
				}
				_ = repo.Create(context.Background(), user)
				userID = user.ID
			} else {
				userID = uuid.New()
			}

			err := svc.DeleteAccount(context.Background(), userID, dto.DeleteAccountRequest{
				Password: tc.password,
			})

			if tc.wantCode == 0 {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if likeRepo.softDeleteCalls != tc.wantLikeCalls {
					t.Errorf("expected %d SoftDeleteByUserID calls, got %d", tc.wantLikeCalls, likeRepo.softDeleteCalls)
				}
			} else {
				if err == nil {
					t.Fatal("expected error")
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T", err)
				}
				if appErr.Code != tc.wantCode {
					t.Errorf("expected code %d, got %d", tc.wantCode, appErr.Code)
				}
			}
		})
	}
}
