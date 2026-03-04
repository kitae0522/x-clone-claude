package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/dto"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	users       map[string]*model.User
	usersByID   map[uuid.UUID]*model.User
	emailExists map[string]bool
	nameExists  map[string]bool
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:       make(map[string]*model.User),
		usersByID:   make(map[uuid.UUID]*model.User),
		emailExists: make(map[string]bool),
		nameExists:  make(map[string]bool),
	}
}

func (m *mockUserRepo) Create(_ context.Context, user *model.User) error {
	user.ID = uuid.New()
	m.users[user.Email] = user
	m.usersByID[user.ID] = user
	m.emailExists[user.Email] = true
	m.nameExists[user.Username] = true
	return nil
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*model.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepo) FindByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	if u, ok := m.usersByID[id]; ok {
		return u, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepo) ExistsByEmail(_ context.Context, email string) (bool, error) {
	return m.emailExists[email], nil
}

func (m *mockUserRepo) FindByUsername(_ context.Context, username string) (*model.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepo) Update(_ context.Context, user *model.User) error {
	m.users[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *mockUserRepo) ExistsByUsername(_ context.Context, username string) (bool, error) {
	return m.nameExists[username], nil
}

func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 24)

	resp, err := svc.Register(context.Background(), dto.RegisterRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.User.Email)
	}
	if resp.User.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", resp.User.Username)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 24)

	_, _ = svc.Register(context.Background(), dto.RegisterRequest{
		Email: "dup@example.com", Username: "user1", Password: "password123",
	})

	_, err := svc.Register(context.Background(), dto.RegisterRequest{
		Email: "dup@example.com", Username: "user2", Password: "password123",
	})

	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 24)

	_, _ = svc.Register(context.Background(), dto.RegisterRequest{
		Email: "a@example.com", Username: "dupuser", Password: "password123",
	})

	_, err := svc.Register(context.Background(), dto.RegisterRequest{
		Email: "b@example.com", Username: "dupuser", Password: "password123",
	})

	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 24)

	_, _ = svc.Register(context.Background(), dto.RegisterRequest{
		Email: "login@example.com", Username: "loginuser", Password: "password123",
	})

	resp, err := svc.Login(context.Background(), dto.LoginRequest{
		Email: "login@example.com", Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_WrongEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 24)

	_, err := svc.Login(context.Background(), dto.LoginRequest{
		Email: "nonexistent@example.com", Password: "password123",
	})

	if err == nil {
		t.Fatal("expected error for wrong email")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 24)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), 10)
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: string(hash),
		Username:     "testuser",
	}
	_ = repo.Create(context.Background(), user)

	_, err := svc.Login(context.Background(), dto.LoginRequest{
		Email: "test@example.com", Password: "wrongpassword",
	})

	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}
