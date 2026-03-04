package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type FollowRepository interface {
	Follow(ctx context.Context, followerID, followingID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	GetFollowing(ctx context.Context, userID uuid.UUID) ([]*model.User, error)
	GetFollowers(ctx context.Context, userID uuid.UUID) ([]*model.User, error)
	CountFollowing(ctx context.Context, userID uuid.UUID) (int, error)
	CountFollowers(ctx context.Context, userID uuid.UUID) (int, error)
}

type followRepository struct {
	pool *pgxpool.Pool
}

func NewFollowRepository(pool *pgxpool.Pool) FollowRepository {
	return &followRepository{pool: pool}
}

func (r *followRepository) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	query := `INSERT INTO follows (follower_id, following_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, followerID, followingID)
	return err
}

func (r *followRepository) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	query := `DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`
	tag, err := r.pool.Exec(ctx, query, followerID, followingID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *followRepository) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)`
	err := r.pool.QueryRow(ctx, query, followerID, followingID).Scan(&exists)
	return exists, err
}

func (r *followRepository) GetFollowing(ctx context.Context, userID uuid.UUID) ([]*model.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.username, u.display_name, u.bio, u.profile_image_url, u.header_image_url, u.created_at, u.updated_at
		FROM follows f
		JOIN users u ON u.id = f.following_id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Username, &u.DisplayName, &u.Bio, &u.ProfileImageURL, &u.HeaderImageURL, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *followRepository) GetFollowers(ctx context.Context, userID uuid.UUID) ([]*model.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.username, u.display_name, u.bio, u.profile_image_url, u.header_image_url, u.created_at, u.updated_at
		FROM follows f
		JOIN users u ON u.id = f.follower_id
		WHERE f.following_id = $1
		ORDER BY f.created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Username, &u.DisplayName, &u.Bio, &u.ProfileImageURL, &u.HeaderImageURL, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *followRepository) CountFollowing(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM follows WHERE follower_id = $1`, userID).Scan(&count)
	return count, err
}

func (r *followRepository) CountFollowers(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM follows WHERE following_id = $1`, userID).Scan(&count)
	return count, err
}
