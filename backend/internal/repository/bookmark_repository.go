package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

var (
	ErrAlreadyBookmarked = errors.New("already bookmarked")
	ErrNotBookmarked     = errors.New("not bookmarked")
)

type BookmarkRepository interface {
	Bookmark(ctx context.Context, userID, postID uuid.UUID) error
	Unbookmark(ctx context.Context, userID, postID uuid.UUID) error
	IsBookmarked(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, cursor time.Time, limit int) ([]model.PostWithAuthor, *time.Time, bool, error)
}

type bookmarkRepository struct {
	pool *pgxpool.Pool
}

func NewBookmarkRepository(pool *pgxpool.Pool) BookmarkRepository {
	return &bookmarkRepository{pool: pool}
}

func (r *bookmarkRepository) Bookmark(ctx context.Context, userID, postID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`INSERT INTO bookmarks (user_id, post_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, postID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert bookmark: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAlreadyBookmarked
	}
	return nil
}

func (r *bookmarkRepository) Unbookmark(ctx context.Context, userID, postID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM bookmarks WHERE user_id = $1 AND post_id = $2`,
		userID, postID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotBookmarked
	}
	return nil
}

func (r *bookmarkRepository) IsBookmarked(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM bookmarks WHERE user_id = $1 AND post_id = $2)`,
		userID, postID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check bookmark status: %w", err)
	}
	return exists, nil
}

func (r *bookmarkRepository) ListByUserID(ctx context.Context, userID uuid.UUID, cursor time.Time, limit int) ([]model.PostWithAuthor, *time.Time, bool, error) {
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.created_at, p.updated_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $1 AND l.post_id = p.id) AS is_liked,
		       b.created_at AS bookmark_created_at
		FROM bookmarks b
		JOIN posts p ON p.id = b.post_id
		LEFT JOIN users u ON p.author_id = u.id
		WHERE b.user_id = $1 AND b.created_at < $2
		ORDER BY b.created_at DESC
		LIMIT $3`

	rows, err := r.pool.Query(ctx, query, userID, cursor, limit+1)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to query bookmarks: %w", err)
	}
	defer rows.Close()

	type postWithBookmarkTime struct {
		post              model.PostWithAuthor
		bookmarkCreatedAt time.Time
	}
	var results []postWithBookmarkTime
	for rows.Next() {
		var item postWithBookmarkTime
		var visibility string
		if err := rows.Scan(
			&item.post.ID, &item.post.AuthorID, &item.post.ParentID, &item.post.Content, &visibility,
			&item.post.LikeCount, &item.post.ReplyCount, &item.post.CreatedAt, &item.post.UpdatedAt,
			&item.post.AuthorUsername, &item.post.AuthorDisplayName, &item.post.AuthorProfileImageURL,
			&item.post.AuthorDeleted,
			&item.post.IsLiked,
			&item.bookmarkCreatedAt,
		); err != nil {
			return nil, nil, false, fmt.Errorf("failed to scan bookmark row: %w", err)
		}
		item.post.Visibility = model.Visibility(visibility)
		item.post.IsBookmarked = true
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, false, fmt.Errorf("failed to iterate bookmark rows: %w", err)
	}

	hasMore := len(results) > limit
	if hasMore {
		results = results[:limit]
	}

	posts := make([]model.PostWithAuthor, len(results))
	var lastCursor *time.Time
	for i, r := range results {
		posts[i] = r.post
		if i == len(results)-1 {
			lastCursor = &r.bookmarkCreatedAt
		}
	}

	return posts, lastCursor, hasMore, nil
}
