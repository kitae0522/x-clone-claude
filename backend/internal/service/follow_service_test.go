package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/apperror"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

// mockFollowRepo implements repository.FollowRepository for testing
type mockFollowRepo struct {
	follows map[string]bool // "followerID:followingID" -> true
	users   map[uuid.UUID]*model.User
}

func newMockFollowRepo() *mockFollowRepo {
	return &mockFollowRepo{
		follows: make(map[string]bool),
		users:   make(map[uuid.UUID]*model.User),
	}
}

func followKey(a, b uuid.UUID) string {
	return a.String() + ":" + b.String()
}

func (m *mockFollowRepo) Follow(_ context.Context, followerID, followingID uuid.UUID) error {
	m.follows[followKey(followerID, followingID)] = true
	return nil
}

func (m *mockFollowRepo) Unfollow(_ context.Context, followerID, followingID uuid.UUID) (bool, error) {
	key := followKey(followerID, followingID)
	if m.follows[key] {
		delete(m.follows, key)
		return true, nil
	}
	return false, nil
}

func (m *mockFollowRepo) IsFollowing(_ context.Context, followerID, followingID uuid.UUID) (bool, error) {
	return m.follows[followKey(followerID, followingID)], nil
}

func (m *mockFollowRepo) GetFollowing(_ context.Context, userID uuid.UUID) ([]*model.User, error) {
	var result []*model.User
	prefix := userID.String() + ":"
	for key := range m.follows {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			followingIDStr := key[len(prefix):]
			followingID, _ := uuid.Parse(followingIDStr)
			if u, ok := m.users[followingID]; ok {
				result = append(result, u)
			}
		}
	}
	return result, nil
}

func (m *mockFollowRepo) GetFollowers(_ context.Context, userID uuid.UUID) ([]*model.User, error) {
	var result []*model.User
	suffix := ":" + userID.String()
	for key := range m.follows {
		if len(key) > len(suffix) && key[len(key)-len(suffix):] == suffix {
			followerIDStr := key[:len(key)-len(suffix)]
			followerID, _ := uuid.Parse(followerIDStr)
			if u, ok := m.users[followerID]; ok {
				result = append(result, u)
			}
		}
	}
	return result, nil
}

func (m *mockFollowRepo) CountFollowing(_ context.Context, userID uuid.UUID) (int, error) {
	count := 0
	prefix := userID.String() + ":"
	for key := range m.follows {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			count++
		}
	}
	return count, nil
}

func (m *mockFollowRepo) CountFollowers(_ context.Context, userID uuid.UUID) (int, error) {
	count := 0
	suffix := ":" + userID.String()
	for key := range m.follows {
		if len(key) > len(suffix) && key[len(key)-len(suffix):] == suffix {
			count++
		}
	}
	return count, nil
}

// mockUserRepoForFollow implements repository.UserRepository for follow tests
type mockUserRepoForFollow struct {
	users       map[string]*model.User // username -> user
	usersByID   map[uuid.UUID]*model.User
	emailExists map[string]bool
	nameExists  map[string]bool
}

func newMockUserRepoForFollow() *mockUserRepoForFollow {
	return &mockUserRepoForFollow{
		users:       make(map[string]*model.User),
		usersByID:   make(map[uuid.UUID]*model.User),
		emailExists: make(map[string]bool),
		nameExists:  make(map[string]bool),
	}
}

func (m *mockUserRepoForFollow) addUser(u *model.User) {
	m.users[u.Username] = u
	m.usersByID[u.ID] = u
	m.emailExists[u.Email] = true
	m.nameExists[u.Username] = true
}

func (m *mockUserRepoForFollow) Create(_ context.Context, user *model.User) error {
	user.ID = uuid.New()
	m.addUser(user)
	return nil
}

func (m *mockUserRepoForFollow) FindByEmail(_ context.Context, email string) (*model.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepoForFollow) FindByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	if u, ok := m.usersByID[id]; ok {
		return u, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepoForFollow) FindByUsername(_ context.Context, username string) (*model.User, error) {
	if u, ok := m.users[username]; ok {
		return u, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepoForFollow) Update(_ context.Context, user *model.User) error {
	m.users[user.Username] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *mockUserRepoForFollow) ExistsByEmail(_ context.Context, email string) (bool, error) {
	return m.emailExists[email], nil
}

func (m *mockUserRepoForFollow) ExistsByUsername(_ context.Context, username string) (bool, error) {
	return m.nameExists[username], nil
}

func (m *mockUserRepoForFollow) UpdatePassword(_ context.Context, id uuid.UUID, passwordHash string) error {
	if u, ok := m.usersByID[id]; ok {
		u.PasswordHash = passwordHash
		return nil
	}
	return pgx.ErrNoRows
}

func (m *mockUserRepoForFollow) SoftDelete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.usersByID[id]; ok {
		delete(m.usersByID, id)
		return nil
	}
	return pgx.ErrNoRows
}

func setupFollowTest() (*followService, *mockFollowRepo, *mockUserRepoForFollow, *model.User, *model.User) {
	userRepo := newMockUserRepoForFollow()
	followRepo := newMockFollowRepo()

	user1 := &model.User{ID: uuid.New(), Username: "alice", Email: "alice@test.com", DisplayName: "Alice"}
	user2 := &model.User{ID: uuid.New(), Username: "bob", Email: "bob@test.com", DisplayName: "Bob"}
	userRepo.addUser(user1)
	userRepo.addUser(user2)
	followRepo.users[user1.ID] = user1
	followRepo.users[user2.ID] = user2

	svc := &followService{followRepo: followRepo, userRepo: userRepo}
	return svc, followRepo, userRepo, user1, user2
}

func TestFollow_Success(t *testing.T) {
	svc, _, _, user1, _ := setupFollowTest()

	resp, err := svc.Follow(context.Background(), user1.ID, "bob")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !resp.Following {
		t.Error("expected following to be true")
	}
}

func TestFollow_SelfFollow(t *testing.T) {
	svc, _, _, user1, _ := setupFollowTest()

	_, err := svc.Follow(context.Background(), user1.ID, "alice")
	if err == nil {
		t.Fatal("expected error for self follow")
	}
	appErr, ok := err.(*apperror.AppError)
	if !ok || appErr.Code != 400 {
		t.Errorf("expected 400 error, got %v", err)
	}
}

func TestFollow_AlreadyFollowing(t *testing.T) {
	svc, _, _, user1, _ := setupFollowTest()

	_, _ = svc.Follow(context.Background(), user1.ID, "bob")
	_, err := svc.Follow(context.Background(), user1.ID, "bob")
	if err == nil {
		t.Fatal("expected error for already following")
	}
	appErr, ok := err.(*apperror.AppError)
	if !ok || appErr.Code != 409 {
		t.Errorf("expected 409 error, got %v", err)
	}
}

func TestFollow_UserNotFound(t *testing.T) {
	svc, _, _, user1, _ := setupFollowTest()

	_, err := svc.Follow(context.Background(), user1.ID, "nonexistent")
	if err == nil {
		t.Fatal("expected error for user not found")
	}
	appErr, ok := err.(*apperror.AppError)
	if !ok || appErr.Code != 404 {
		t.Errorf("expected 404 error, got %v", err)
	}
}

func TestUnfollow_Success(t *testing.T) {
	svc, _, _, user1, _ := setupFollowTest()

	_, _ = svc.Follow(context.Background(), user1.ID, "bob")
	resp, err := svc.Unfollow(context.Background(), user1.ID, "bob")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Following {
		t.Error("expected following to be false")
	}
}

func TestUnfollow_NotFollowing(t *testing.T) {
	svc, _, _, user1, _ := setupFollowTest()

	_, err := svc.Unfollow(context.Background(), user1.ID, "bob")
	if err == nil {
		t.Fatal("expected error for not following")
	}
	appErr, ok := err.(*apperror.AppError)
	if !ok || appErr.Code != 404 {
		t.Errorf("expected 404 error, got %v", err)
	}
}

func TestUnfollow_SelfUnfollow(t *testing.T) {
	svc, _, _, user1, _ := setupFollowTest()

	_, err := svc.Unfollow(context.Background(), user1.ID, "alice")
	if err == nil {
		t.Fatal("expected error for self unfollow")
	}
	appErr, ok := err.(*apperror.AppError)
	if !ok || appErr.Code != 400 {
		t.Errorf("expected 400 error, got %v", err)
	}
}

func TestGetFollowing_Success(t *testing.T) {
	svc, _, _, user1, _ := setupFollowTest()

	_, _ = svc.Follow(context.Background(), user1.ID, "bob")
	resp, err := svc.GetFollowing(context.Background(), "alice")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected 1 following, got %d", resp.Total)
	}
	if resp.Users[0].Username != "bob" {
		t.Errorf("expected bob, got %s", resp.Users[0].Username)
	}
}

func TestGetFollowers_Success(t *testing.T) {
	svc, _, _, user1, _ := setupFollowTest()

	_, _ = svc.Follow(context.Background(), user1.ID, "bob")
	resp, err := svc.GetFollowers(context.Background(), "bob")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected 1 follower, got %d", resp.Total)
	}
	if resp.Users[0].Username != "alice" {
		t.Errorf("expected alice, got %s", resp.Users[0].Username)
	}
}
