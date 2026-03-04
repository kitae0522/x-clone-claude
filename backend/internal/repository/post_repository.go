package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.PostWithAuthor, error)
	FindAll(ctx context.Context, limit, offset int) ([]model.PostWithAuthor, error)
	FindByIDWithUser(ctx context.Context, id, userID uuid.UUID) (*model.PostWithAuthor, error)
	FindAllWithUser(ctx context.Context, limit, offset int, userID uuid.UUID) ([]model.PostWithAuthor, error)
}

type postRepository struct {
	pool *pgxpool.Pool
}

func NewPostRepository(pool *pgxpool.Pool) PostRepository {
	return &postRepository{pool: pool}
}

func (r *postRepository) Create(ctx context.Context, post *model.Post) error {
	query := `
		INSERT INTO posts (author_id, content, visibility)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		post.AuthorID, post.Content, string(post.Visibility),
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (r *postRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.PostWithAuthor, error) {
	p := &model.PostWithAuthor{}
	query := `
		SELECT p.id, p.author_id, p.content, p.visibility, p.like_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.id = $1`

	var visibility string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.AuthorID, &p.Content, &visibility, &p.LikeCount, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
	)
	if err != nil {
		return nil, err
	}
	p.Visibility = model.Visibility(visibility)
	return p, nil
}

func (r *postRepository) FindAll(ctx context.Context, limit, offset int) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.content, p.visibility, p.like_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url
		FROM posts p
		JOIN users u ON p.author_id = u.id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.PostWithAuthor
	for rows.Next() {
		var p model.PostWithAuthor
		var visibility string
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Content, &visibility, &p.LikeCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
		); err != nil {
			return nil, err
		}
		p.Visibility = model.Visibility(visibility)
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) FindByIDWithUser(ctx context.Context, id, userID uuid.UUID) (*model.PostWithAuthor, error) {
	p := &model.PostWithAuthor{}
	query := `
		SELECT p.id, p.author_id, p.content, p.visibility, p.like_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url,
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $2 AND l.post_id = p.id) AS is_liked
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.id = $1`

	var visibility string
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&p.ID, &p.AuthorID, &p.Content, &visibility, &p.LikeCount, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
		&p.IsLiked,
	)
	if err != nil {
		return nil, err
	}
	p.Visibility = model.Visibility(visibility)
	return p, nil
}

func (r *postRepository) FindAllWithUser(ctx context.Context, limit, offset int, userID uuid.UUID) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.content, p.visibility, p.like_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url,
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $3 AND l.post_id = p.id) AS is_liked
		FROM posts p
		JOIN users u ON p.author_id = u.id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.PostWithAuthor
	for rows.Next() {
		var p model.PostWithAuthor
		var visibility string
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Content, &visibility, &p.LikeCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
			&p.IsLiked,
		); err != nil {
			return nil, err
		}
		p.Visibility = model.Visibility(visibility)
		posts = append(posts, p)
	}
	return posts, rows.Err()
}
