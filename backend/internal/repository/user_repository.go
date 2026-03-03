package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, password_hash, username, display_name, bio, profile_image_url, header_image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		user.Email, user.PasswordHash, user.Username, user.DisplayName, user.Bio, user.ProfileImageURL, user.HeaderImageURL,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, email, password_hash, username, display_name, bio, profile_image_url, header_image_url, created_at, updated_at
		FROM users WHERE email = $1`

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Username,
		&user.DisplayName, &user.Bio, &user.ProfileImageURL, &user.HeaderImageURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, email, password_hash, username, display_name, bio, profile_image_url, header_image_url, created_at, updated_at
		FROM users WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Username,
		&user.DisplayName, &user.Bio, &user.ProfileImageURL, &user.HeaderImageURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, email, password_hash, username, display_name, bio, profile_image_url, header_image_url, created_at, updated_at
		FROM users WHERE username = $1`

	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Username,
		&user.DisplayName, &user.Bio, &user.ProfileImageURL, &user.HeaderImageURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET username = $1, display_name = $2, bio = $3, profile_image_url = $4, header_image_url = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		user.Username, user.DisplayName, user.Bio, user.ProfileImageURL, user.HeaderImageURL, user.ID,
	).Scan(&user.UpdatedAt)
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	return exists, err
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	return exists, err
}
