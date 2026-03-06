package repository

import (
	"context"
	"fmt"

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
	CreateReply(ctx context.Context, post *model.Post) error
	FindRepliesByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]model.PostWithAuthor, error)
	FindRepliesByPostIDWithUser(ctx context.Context, postID, userID uuid.UUID, limit, offset int) ([]model.PostWithAuthor, error)
	FindAuthorReplyByPostID(ctx context.Context, postID, authorID uuid.UUID) (*model.PostWithAuthor, error)
	FindAuthorReplyByPostIDWithUser(ctx context.Context, postID, authorID, userID uuid.UUID) (*model.PostWithAuthor, error)
	FindByAuthorHandle(ctx context.Context, handle string, limit, offset int) ([]model.PostWithAuthor, error)
	FindByAuthorHandleWithUser(ctx context.Context, handle string, limit, offset int, userID uuid.UUID) ([]model.PostWithAuthor, error)
	FindRepliesByAuthorHandle(ctx context.Context, handle string, limit, offset int) ([]model.PostWithAuthor, error)
	FindRepliesByAuthorHandleWithUser(ctx context.Context, handle string, limit, offset int, userID uuid.UUID) ([]model.PostWithAuthor, error)
	FindLikedByUserHandle(ctx context.Context, handle string, limit, offset int) ([]model.PostWithAuthor, error)
	FindLikedByUserHandleWithViewer(ctx context.Context, handle string, limit, offset int, viewerID uuid.UUID) ([]model.PostWithAuthor, error)
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.id = $1`

	var visibility string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.CreatedAt, &p.UpdatedAt,
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.parent_id IS NULL
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
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.CreatedAt, &p.UpdatedAt,
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url,
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $2 AND l.post_id = p.id) AS is_liked
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.id = $1`

	var visibility string
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.CreatedAt, &p.UpdatedAt,
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url,
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $3 AND l.post_id = p.id) AS is_liked
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.parent_id IS NULL
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
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.CreatedAt, &p.UpdatedAt,
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

func (r *postRepository) CreateReply(ctx context.Context, post *model.Post) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO posts (author_id, parent_id, content, visibility)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(ctx, query,
		post.AuthorID, post.ParentID, post.Content, string(post.Visibility),
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert reply: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE posts SET reply_count = reply_count + 1 WHERE id = $1`,
		post.ParentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update reply_count: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *postRepository) FindRepliesByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.parent_id = $1
		ORDER BY p.created_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []model.PostWithAuthor
	for rows.Next() {
		var p model.PostWithAuthor
		var visibility string
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
		); err != nil {
			return nil, err
		}
		p.Visibility = model.Visibility(visibility)
		replies = append(replies, p)
	}
	return replies, rows.Err()
}

func (r *postRepository) FindAuthorReplyByPostID(ctx context.Context, postID, authorID uuid.UUID) (*model.PostWithAuthor, error) {
	p := &model.PostWithAuthor{}
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.parent_id = $1 AND p.author_id = $2
		ORDER BY p.created_at ASC
		LIMIT 1`

	var visibility string
	err := r.pool.QueryRow(ctx, query, postID, authorID).Scan(
		&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
	)
	if err != nil {
		return nil, err
	}
	p.Visibility = model.Visibility(visibility)
	return p, nil
}

func (r *postRepository) FindAuthorReplyByPostIDWithUser(ctx context.Context, postID, authorID, userID uuid.UUID) (*model.PostWithAuthor, error) {
	p := &model.PostWithAuthor{}
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url,
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $3 AND l.post_id = p.id) AS is_liked
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.parent_id = $1 AND p.author_id = $2
		ORDER BY p.created_at ASC
		LIMIT 1`

	var visibility string
	err := r.pool.QueryRow(ctx, query, postID, authorID, userID).Scan(
		&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
		&p.IsLiked,
	)
	if err != nil {
		return nil, err
	}
	p.Visibility = model.Visibility(visibility)
	return p, nil
}

func (r *postRepository) FindRepliesByPostIDWithUser(ctx context.Context, postID, userID uuid.UUID, limit, offset int) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url,
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $3 AND l.post_id = p.id) AS is_liked
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.parent_id = $1
		ORDER BY p.created_at ASC
		LIMIT $2 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, postID, limit, userID, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []model.PostWithAuthor
	for rows.Next() {
		var p model.PostWithAuthor
		var visibility string
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
			&p.IsLiked,
		); err != nil {
			return nil, err
		}
		p.Visibility = model.Visibility(visibility)
		replies = append(replies, p)
	}
	return replies, rows.Err()
}

type scannable interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
}

func (r *postRepository) scanPostRows(rows scannable, withIsLiked bool) ([]model.PostWithAuthor, error) {
	defer rows.Close()
	var posts []model.PostWithAuthor
	for rows.Next() {
		var p model.PostWithAuthor
		var visibility string
		var scanArgs []any
		scanArgs = append(scanArgs,
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
		)
		if withIsLiked {
			scanArgs = append(scanArgs, &p.IsLiked)
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		p.Visibility = model.Visibility(visibility)
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) FindByAuthorHandle(ctx context.Context, handle string, limit, offset int) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE u.username = $1 AND p.parent_id IS NULL
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, handle, limit, offset)
	if err != nil {
		return nil, err
	}
	return r.scanPostRows(rows, false)
}

func (r *postRepository) FindByAuthorHandleWithUser(ctx context.Context, handle string, limit, offset int, userID uuid.UUID) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url,
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $4 AND l.post_id = p.id) AS is_liked
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE u.username = $1 AND p.parent_id IS NULL
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, handle, limit, offset, userID)
	if err != nil {
		return nil, err
	}
	return r.scanPostRows(rows, true)
}

func (r *postRepository) scanReplyWithParentRows(rows scannable, withIsLiked bool) ([]model.PostWithAuthor, error) {
	defer rows.Close()
	var posts []model.PostWithAuthor
	for rows.Next() {
		var p model.PostWithAuthor
		var visibility string
		var scanArgs []any
		scanArgs = append(scanArgs,
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
		)
		if withIsLiked {
			scanArgs = append(scanArgs, &p.IsLiked)
		}
		scanArgs = append(scanArgs,
			&p.ParentPostID, &p.ParentContent,
			&p.ParentAuthorUsername, &p.ParentAuthorDisplayName, &p.ParentAuthorProfileImageURL,
		)
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		p.Visibility = model.Visibility(visibility)
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) FindRepliesByAuthorHandle(ctx context.Context, handle string, limit, offset int) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url,
		       pp.id, pp.content,
		       pu.username, pu.display_name, pu.profile_image_url
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN posts pp ON pp.id = p.parent_id
		LEFT JOIN users pu ON pu.id = pp.author_id
		WHERE u.username = $1 AND p.parent_id IS NOT NULL
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, handle, limit, offset)
	if err != nil {
		return nil, err
	}
	return r.scanReplyWithParentRows(rows, false)
}

func (r *postRepository) FindRepliesByAuthorHandleWithUser(ctx context.Context, handle string, limit, offset int, userID uuid.UUID) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url,
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $4 AND l.post_id = p.id) AS is_liked,
		       pp.id, pp.content,
		       pu.username, pu.display_name, pu.profile_image_url
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN posts pp ON pp.id = p.parent_id
		LEFT JOIN users pu ON pu.id = pp.author_id
		WHERE u.username = $1 AND p.parent_id IS NOT NULL
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, handle, limit, offset, userID)
	if err != nil {
		return nil, err
	}
	return r.scanReplyWithParentRows(rows, true)
}

func (r *postRepository) FindLikedByUserHandle(ctx context.Context, handle string, limit, offset int) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url
		FROM likes lk
		JOIN users target ON target.username = $1
		JOIN posts p ON p.id = lk.post_id
		JOIN users u ON p.author_id = u.id
		WHERE lk.user_id = target.id
		ORDER BY lk.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, handle, limit, offset)
	if err != nil {
		return nil, err
	}
	return r.scanPostRows(rows, false)
}

func (r *postRepository) FindLikedByUserHandleWithViewer(ctx context.Context, handle string, limit, offset int, viewerID uuid.UUID) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       u.username, u.display_name, u.profile_image_url,
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $4 AND l.post_id = p.id) AS is_liked
		FROM likes lk
		JOIN users target ON target.username = $1
		JOIN posts p ON p.id = lk.post_id
		JOIN users u ON p.author_id = u.id
		WHERE lk.user_id = target.id
		ORDER BY lk.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, handle, limit, offset, viewerID)
	if err != nil {
		return nil, err
	}
	return r.scanPostRows(rows, true)
}
