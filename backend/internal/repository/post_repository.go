package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	IncrementViewCountBatch(ctx context.Context, ids []uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, content string, visibility model.Visibility, locationLat *float64, locationLng *float64, locationName *string) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	SoftDeleteReply(ctx context.Context, id uuid.UUID, parentID uuid.UUID) error
	ExistsIncludingDeleted(ctx context.Context, id uuid.UUID) (exists bool, isDeleted bool, err error)
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*model.PostWithAuthor, error)
	FindDeletedByAuthor(ctx context.Context, authorID uuid.UUID, limit int, cursor *time.Time) ([]model.PostWithAuthor, error)
	Restore(ctx context.Context, id uuid.UUID) error
	RestoreReply(ctx context.Context, id uuid.UUID, parentID uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error
}

type postRepository struct {
	pool *pgxpool.Pool
}

func NewPostRepository(pool *pgxpool.Pool) PostRepository {
	return &postRepository{pool: pool}
}

func (r *postRepository) Create(ctx context.Context, post *model.Post) error {
	query := `
		INSERT INTO posts (author_id, content, visibility, location_lat, location_lng, location_name)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		post.AuthorID, post.Content, string(post.Visibility),
		post.LocationLat, post.LocationLng, post.LocationName,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
}

func (r *postRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.PostWithAuthor, error) {
	p := &model.PostWithAuthor{}
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       p.location_lat, p.location_lng, p.location_name
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.id = $1 AND p.deleted_at IS NULL`

	var visibility string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
		&p.AuthorDeleted,
		&p.LocationLat, &p.LocationLng, &p.LocationName,
	)
	if err != nil {
		return nil, err
	}
	p.Visibility = model.Visibility(visibility)
	return p, nil
}

func (r *postRepository) FindAll(ctx context.Context, limit, offset int) ([]model.PostWithAuthor, error) {
	query := `
		SELECT id, author_id, parent_id, content, visibility, like_count, reply_count, view_count, repost_count, created_at, updated_at,
		       username, display_name, profile_image_url, author_deleted,
		       location_lat, location_lng, location_name,
		       reposted_by_username, reposted_by_display_name, reposted_at
		FROM (
		  SELECT DISTINCT ON (id) id, author_id, parent_id, content, visibility, like_count, reply_count, view_count, repost_count, created_at, updated_at,
		         username, display_name, profile_image_url, author_deleted,
		         location_lat, location_lng, location_name,
		         reposted_by_username, reposted_by_display_name, reposted_at, sort_time
		  FROM (
		    SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		           COALESCE(u.username, '') AS username, COALESCE(u.display_name, '') AS display_name, COALESCE(u.profile_image_url, '') AS profile_image_url,
		           (u.deleted_at IS NOT NULL OR u.id IS NULL) AS author_deleted,
		           p.location_lat, p.location_lng, p.location_name,
		           NULL::TEXT AS reposted_by_username, NULL::TEXT AS reposted_by_display_name, NULL::TIMESTAMPTZ AS reposted_at,
		           p.created_at AS sort_time
		    FROM posts p
		    LEFT JOIN users u ON p.author_id = u.id
		    WHERE p.parent_id IS NULL AND p.visibility = 'public' AND p.deleted_at IS NULL

		    UNION ALL

		    SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		           COALESCE(u.username, '') AS username, COALESCE(u.display_name, '') AS display_name, COALESCE(u.profile_image_url, '') AS profile_image_url,
		           (u.deleted_at IS NOT NULL OR u.id IS NULL) AS author_deleted,
		           p.location_lat, p.location_lng, p.location_name,
		           ru.username AS reposted_by_username, ru.display_name AS reposted_by_display_name, rp.created_at AS reposted_at,
		           rp.created_at AS sort_time
		    FROM reposts rp
		    JOIN posts p ON p.id = rp.post_id
		    LEFT JOIN users u ON p.author_id = u.id
		    JOIN users ru ON rp.user_id = ru.id
		    WHERE p.parent_id IS NULL AND p.visibility = 'public' AND p.deleted_at IS NULL
		  ) sub
		  ORDER BY id, sort_time DESC
		) deduped
		ORDER BY sort_time DESC
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
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
			&p.AuthorDeleted,
			&p.LocationLat, &p.LocationLng, &p.LocationName,
			&p.RepostedByUsername, &p.RepostedByDisplayName, &p.RepostedAt,
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $2 AND l.post_id = p.id AND l.deleted_at IS NULL) AS is_liked,
		       EXISTS(SELECT 1 FROM bookmarks b WHERE b.user_id = $2 AND b.post_id = p.id) AS is_bookmarked,
		       EXISTS(SELECT 1 FROM reposts r WHERE r.user_id = $2 AND r.post_id = p.id) AS is_reposted,
		       p.location_lat, p.location_lng, p.location_name
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.id = $1 AND p.deleted_at IS NULL`

	var visibility string
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
		&p.AuthorDeleted,
		&p.IsLiked, &p.IsBookmarked, &p.IsReposted,
		&p.LocationLat, &p.LocationLng, &p.LocationName,
	)
	if err != nil {
		return nil, err
	}
	p.Visibility = model.Visibility(visibility)
	return p, nil
}

func (r *postRepository) FindAllWithUser(ctx context.Context, limit, offset int, userID uuid.UUID) ([]model.PostWithAuthor, error) {
	query := `
		SELECT id, author_id, parent_id, content, visibility, like_count, reply_count, view_count, repost_count, created_at, updated_at,
		       username, display_name, profile_image_url, author_deleted,
		       is_liked, is_bookmarked, is_reposted,
		       location_lat, location_lng, location_name,
		       reposted_by_username, reposted_by_display_name, reposted_at
		FROM (
		  SELECT DISTINCT ON (id) id, author_id, parent_id, content, visibility, like_count, reply_count, view_count, repost_count, created_at, updated_at,
		         username, display_name, profile_image_url, author_deleted,
		         is_liked, is_bookmarked, is_reposted,
		         location_lat, location_lng, location_name,
		         reposted_by_username, reposted_by_display_name, reposted_at, sort_time
		  FROM (
		    SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		           COALESCE(u.username, '') AS username, COALESCE(u.display_name, '') AS display_name, COALESCE(u.profile_image_url, '') AS profile_image_url,
		           (u.deleted_at IS NOT NULL OR u.id IS NULL) AS author_deleted,
		           EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $3 AND l.post_id = p.id AND l.deleted_at IS NULL) AS is_liked,
		           EXISTS(SELECT 1 FROM bookmarks b WHERE b.user_id = $3 AND b.post_id = p.id) AS is_bookmarked,
		           EXISTS(SELECT 1 FROM reposts r WHERE r.user_id = $3 AND r.post_id = p.id) AS is_reposted,
		           p.location_lat, p.location_lng, p.location_name,
		           NULL::TEXT AS reposted_by_username, NULL::TEXT AS reposted_by_display_name, NULL::TIMESTAMPTZ AS reposted_at,
		           p.created_at AS sort_time
		    FROM posts p
		    LEFT JOIN users u ON p.author_id = u.id
		    WHERE p.parent_id IS NULL
		      AND (
		        p.visibility = 'public'
		        OR (p.visibility = 'follower' AND (
		          p.author_id = $3
		          OR EXISTS (SELECT 1 FROM follows f WHERE f.follower_id = $3 AND f.following_id = p.author_id)
		        ))
		        OR (p.visibility = 'private' AND p.author_id = $3)
		      )
		      AND p.deleted_at IS NULL

		    UNION ALL

		    SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		           COALESCE(u.username, '') AS username, COALESCE(u.display_name, '') AS display_name, COALESCE(u.profile_image_url, '') AS profile_image_url,
		           (u.deleted_at IS NOT NULL OR u.id IS NULL) AS author_deleted,
		           EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $3 AND l.post_id = p.id AND l.deleted_at IS NULL) AS is_liked,
		           EXISTS(SELECT 1 FROM bookmarks b WHERE b.user_id = $3 AND b.post_id = p.id) AS is_bookmarked,
		           EXISTS(SELECT 1 FROM reposts r WHERE r.user_id = $3 AND r.post_id = p.id) AS is_reposted,
		           p.location_lat, p.location_lng, p.location_name,
		           ru.username AS reposted_by_username, ru.display_name AS reposted_by_display_name, rp.created_at AS reposted_at,
		           rp.created_at AS sort_time
		    FROM reposts rp
		    JOIN posts p ON p.id = rp.post_id
		    LEFT JOIN users u ON p.author_id = u.id
		    JOIN users ru ON rp.user_id = ru.id
		    WHERE p.parent_id IS NULL AND p.deleted_at IS NULL
		      AND (
		        p.visibility = 'public'
		        OR (p.visibility = 'follower' AND (
		          p.author_id = $3
		          OR EXISTS (SELECT 1 FROM follows f WHERE f.follower_id = $3 AND f.following_id = p.author_id)
		        ))
		        OR (p.visibility = 'private' AND p.author_id = $3)
		      )
		  ) sub
		  ORDER BY id, sort_time DESC
		) deduped
		ORDER BY sort_time DESC
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
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
			&p.AuthorDeleted,
			&p.IsLiked, &p.IsBookmarked, &p.IsReposted,
			&p.LocationLat, &p.LocationLng, &p.LocationName,
			&p.RepostedByUsername, &p.RepostedByDisplayName, &p.RepostedAt,
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
		INSERT INTO posts (author_id, parent_id, content, visibility, location_lat, location_lng, location_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(ctx, query,
		post.AuthorID, post.ParentID, post.Content, string(post.Visibility),
		post.LocationLat, post.LocationLng, post.LocationName,
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       p.location_lat, p.location_lng, p.location_name
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.parent_id = $1
		  AND p.deleted_at IS NULL
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
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
			&p.AuthorDeleted,
			&p.LocationLat, &p.LocationLng, &p.LocationName,
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       p.location_lat, p.location_lng, p.location_name
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.parent_id = $1 AND p.author_id = $2
		  AND p.deleted_at IS NULL
		ORDER BY p.created_at ASC
		LIMIT 1`

	var visibility string
	err := r.pool.QueryRow(ctx, query, postID, authorID).Scan(
		&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
		&p.AuthorDeleted,
		&p.LocationLat, &p.LocationLng, &p.LocationName,
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $3 AND l.post_id = p.id AND l.deleted_at IS NULL) AS is_liked,
		       EXISTS(SELECT 1 FROM bookmarks b WHERE b.user_id = $3 AND b.post_id = p.id) AS is_bookmarked,
		       EXISTS(SELECT 1 FROM reposts r WHERE r.user_id = $3 AND r.post_id = p.id) AS is_reposted,
		       p.location_lat, p.location_lng, p.location_name
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.parent_id = $1 AND p.author_id = $2
		  AND p.deleted_at IS NULL
		ORDER BY p.created_at ASC
		LIMIT 1`

	var visibility string
	err := r.pool.QueryRow(ctx, query, postID, authorID, userID).Scan(
		&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
		&p.AuthorDeleted,
		&p.IsLiked, &p.IsBookmarked, &p.IsReposted,
		&p.LocationLat, &p.LocationLng, &p.LocationName,
	)
	if err != nil {
		return nil, err
	}
	p.Visibility = model.Visibility(visibility)
	return p, nil
}

func (r *postRepository) FindRepliesByPostIDWithUser(ctx context.Context, postID, userID uuid.UUID, limit, offset int) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $3 AND l.post_id = p.id AND l.deleted_at IS NULL) AS is_liked,
		       EXISTS(SELECT 1 FROM bookmarks b WHERE b.user_id = $3 AND b.post_id = p.id) AS is_bookmarked,
		       EXISTS(SELECT 1 FROM reposts r WHERE r.user_id = $3 AND r.post_id = p.id) AS is_reposted,
		       p.location_lat, p.location_lng, p.location_name
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.parent_id = $1
		  AND p.deleted_at IS NULL
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
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
			&p.AuthorDeleted,
			&p.IsLiked, &p.IsBookmarked, &p.IsReposted,
			&p.LocationLat, &p.LocationLng, &p.LocationName,
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
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
			&p.AuthorDeleted,
		)
		if withIsLiked {
			scanArgs = append(scanArgs, &p.IsLiked, &p.IsBookmarked, &p.IsReposted)
		}
		scanArgs = append(scanArgs, &p.LocationLat, &p.LocationLng, &p.LocationName)
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
		SELECT id, author_id, parent_id, content, visibility, like_count, reply_count, view_count, repost_count, created_at, updated_at,
		       username, display_name, profile_image_url, author_deleted,
		       location_lat, location_lng, location_name,
		       reposted_by_username, reposted_by_display_name, reposted_at
		FROM (
		  SELECT DISTINCT ON (id) id, author_id, parent_id, content, visibility, like_count, reply_count, view_count, repost_count, created_at, updated_at,
		         username, display_name, profile_image_url, author_deleted,
		         location_lat, location_lng, location_name,
		         reposted_by_username, reposted_by_display_name, reposted_at, sort_time
		  FROM (
		    SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		           COALESCE(u.username, '') AS username, COALESCE(u.display_name, '') AS display_name, COALESCE(u.profile_image_url, '') AS profile_image_url,
		           (u.deleted_at IS NOT NULL OR u.id IS NULL) AS author_deleted,
		           p.location_lat, p.location_lng, p.location_name,
		           NULL::TEXT AS reposted_by_username, NULL::TEXT AS reposted_by_display_name, NULL::TIMESTAMPTZ AS reposted_at,
		           p.created_at AS sort_time
		    FROM posts p
		    LEFT JOIN users u ON p.author_id = u.id
		    WHERE u.username = $1 AND u.deleted_at IS NULL AND p.parent_id IS NULL
		      AND p.visibility = 'public' AND p.deleted_at IS NULL

		    UNION ALL

		    SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		           COALESCE(u.username, '') AS username, COALESCE(u.display_name, '') AS display_name, COALESCE(u.profile_image_url, '') AS profile_image_url,
		           (u.deleted_at IS NOT NULL OR u.id IS NULL) AS author_deleted,
		           p.location_lat, p.location_lng, p.location_name,
		           ru.username AS reposted_by_username, ru.display_name AS reposted_by_display_name, rp.created_at AS reposted_at,
		           rp.created_at AS sort_time
		    FROM reposts rp
		    JOIN users ru ON rp.user_id = ru.id
		    JOIN posts p ON p.id = rp.post_id
		    LEFT JOIN users u ON p.author_id = u.id
		    WHERE ru.username = $1 AND ru.deleted_at IS NULL AND p.parent_id IS NULL
		      AND p.visibility = 'public' AND p.deleted_at IS NULL
		  ) sub
		  ORDER BY id, sort_time DESC
		) deduped
		ORDER BY sort_time DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, handle, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.PostWithAuthor
	for rows.Next() {
		var p model.PostWithAuthor
		var visibility string
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
			&p.AuthorDeleted,
			&p.LocationLat, &p.LocationLng, &p.LocationName,
			&p.RepostedByUsername, &p.RepostedByDisplayName, &p.RepostedAt,
		); err != nil {
			return nil, err
		}
		p.Visibility = model.Visibility(visibility)
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) FindByAuthorHandleWithUser(ctx context.Context, handle string, limit, offset int, userID uuid.UUID) ([]model.PostWithAuthor, error) {
	query := `
		SELECT id, author_id, parent_id, content, visibility, like_count, reply_count, view_count, repost_count, created_at, updated_at,
		       username, display_name, profile_image_url, author_deleted,
		       is_liked, is_bookmarked, is_reposted,
		       location_lat, location_lng, location_name,
		       reposted_by_username, reposted_by_display_name, reposted_at
		FROM (
		  SELECT DISTINCT ON (id) id, author_id, parent_id, content, visibility, like_count, reply_count, view_count, repost_count, created_at, updated_at,
		         username, display_name, profile_image_url, author_deleted,
		         is_liked, is_bookmarked, is_reposted,
		         location_lat, location_lng, location_name,
		         reposted_by_username, reposted_by_display_name, reposted_at, sort_time
		  FROM (
		    SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		           COALESCE(u.username, '') AS username, COALESCE(u.display_name, '') AS display_name, COALESCE(u.profile_image_url, '') AS profile_image_url,
		           (u.deleted_at IS NOT NULL OR u.id IS NULL) AS author_deleted,
		           EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $4 AND l.post_id = p.id AND l.deleted_at IS NULL) AS is_liked,
		           EXISTS(SELECT 1 FROM bookmarks b WHERE b.user_id = $4 AND b.post_id = p.id) AS is_bookmarked,
		           EXISTS(SELECT 1 FROM reposts r WHERE r.user_id = $4 AND r.post_id = p.id) AS is_reposted,
		           p.location_lat, p.location_lng, p.location_name,
		           NULL::TEXT AS reposted_by_username, NULL::TEXT AS reposted_by_display_name, NULL::TIMESTAMPTZ AS reposted_at,
		           p.created_at AS sort_time
		    FROM posts p
		    LEFT JOIN users u ON p.author_id = u.id
		    WHERE u.username = $1 AND u.deleted_at IS NULL AND p.parent_id IS NULL AND p.deleted_at IS NULL
		      AND (
		        p.visibility = 'public'
		        OR (p.visibility = 'follower' AND (
		          p.author_id = $4
		          OR EXISTS (SELECT 1 FROM follows f WHERE f.follower_id = $4 AND f.following_id = p.author_id)
		        ))
		        OR (p.visibility = 'private' AND p.author_id = $4)
		      )

		    UNION ALL

		    SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		           COALESCE(u.username, '') AS username, COALESCE(u.display_name, '') AS display_name, COALESCE(u.profile_image_url, '') AS profile_image_url,
		           (u.deleted_at IS NOT NULL OR u.id IS NULL) AS author_deleted,
		           EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $4 AND l.post_id = p.id AND l.deleted_at IS NULL) AS is_liked,
		           EXISTS(SELECT 1 FROM bookmarks b WHERE b.user_id = $4 AND b.post_id = p.id) AS is_bookmarked,
		           EXISTS(SELECT 1 FROM reposts r WHERE r.user_id = $4 AND r.post_id = p.id) AS is_reposted,
		           p.location_lat, p.location_lng, p.location_name,
		           ru.username AS reposted_by_username, ru.display_name AS reposted_by_display_name, rp.created_at AS reposted_at,
		           rp.created_at AS sort_time
		    FROM reposts rp
		    JOIN users ru ON rp.user_id = ru.id
		    JOIN posts p ON p.id = rp.post_id
		    LEFT JOIN users u ON p.author_id = u.id
		    WHERE ru.username = $1 AND ru.deleted_at IS NULL AND p.parent_id IS NULL AND p.deleted_at IS NULL
		      AND (
		        p.visibility = 'public'
		        OR (p.visibility = 'follower' AND (
		          p.author_id = $4
		          OR EXISTS (SELECT 1 FROM follows f WHERE f.follower_id = $4 AND f.following_id = p.author_id)
		        ))
		        OR (p.visibility = 'private' AND p.author_id = $4)
		      )
		  ) sub
		  ORDER BY id, sort_time DESC
		) deduped
		ORDER BY sort_time DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, handle, limit, offset, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.PostWithAuthor
	for rows.Next() {
		var p model.PostWithAuthor
		var visibility string
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
			&p.AuthorDeleted,
			&p.IsLiked, &p.IsBookmarked, &p.IsReposted,
			&p.LocationLat, &p.LocationLng, &p.LocationName,
			&p.RepostedByUsername, &p.RepostedByDisplayName, &p.RepostedAt,
		); err != nil {
			return nil, err
		}
		p.Visibility = model.Visibility(visibility)
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) scanReplyWithParentRows(rows scannable, withIsLiked bool) ([]model.PostWithAuthor, error) {
	defer rows.Close()
	var posts []model.PostWithAuthor
	for rows.Next() {
		var p model.PostWithAuthor
		var visibility string
		var scanArgs []any
		scanArgs = append(scanArgs,
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
			&p.AuthorDeleted,
		)
		if withIsLiked {
			scanArgs = append(scanArgs, &p.IsLiked, &p.IsBookmarked, &p.IsReposted)
		}
		scanArgs = append(scanArgs, &p.LocationLat, &p.LocationLng, &p.LocationName)
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       p.location_lat, p.location_lng, p.location_name,
		       pp.id, pp.content,
		       pu.username, pu.display_name, pu.profile_image_url
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		LEFT JOIN posts pp ON pp.id = p.parent_id
		LEFT JOIN users pu ON pu.id = pp.author_id
		WHERE u.username = $1 AND u.deleted_at IS NULL AND p.parent_id IS NOT NULL
		  AND p.visibility = 'public' AND p.deleted_at IS NULL
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $4 AND l.post_id = p.id AND l.deleted_at IS NULL) AS is_liked,
		       EXISTS(SELECT 1 FROM bookmarks b WHERE b.user_id = $4 AND b.post_id = p.id) AS is_bookmarked,
		       EXISTS(SELECT 1 FROM reposts r WHERE r.user_id = $4 AND r.post_id = p.id) AS is_reposted,
		       p.location_lat, p.location_lng, p.location_name,
		       pp.id, pp.content,
		       pu.username, pu.display_name, pu.profile_image_url
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		LEFT JOIN posts pp ON pp.id = p.parent_id
		LEFT JOIN users pu ON pu.id = pp.author_id
		WHERE u.username = $1 AND u.deleted_at IS NULL AND p.parent_id IS NOT NULL AND p.deleted_at IS NULL
		  AND (
		    p.visibility = 'public'
		    OR (p.visibility = 'follower' AND (
		      p.author_id = $4
		      OR EXISTS (SELECT 1 FROM follows f WHERE f.follower_id = $4 AND f.following_id = p.author_id)
		    ))
		    OR (p.visibility = 'private' AND p.author_id = $4)
		  )
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       p.location_lat, p.location_lng, p.location_name
		FROM likes lk
		JOIN users target ON target.username = $1 AND target.deleted_at IS NULL
		JOIN posts p ON p.id = lk.post_id
		LEFT JOIN users u ON p.author_id = u.id
		WHERE lk.user_id = target.id AND lk.deleted_at IS NULL
		  AND p.visibility = 'public' AND p.deleted_at IS NULL
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
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $4 AND l.post_id = p.id AND l.deleted_at IS NULL) AS is_liked,
		       EXISTS(SELECT 1 FROM bookmarks b WHERE b.user_id = $4 AND b.post_id = p.id) AS is_bookmarked,
		       EXISTS(SELECT 1 FROM reposts r WHERE r.user_id = $4 AND r.post_id = p.id) AS is_reposted,
		       p.location_lat, p.location_lng, p.location_name
		FROM likes lk
		JOIN users target ON target.username = $1 AND target.deleted_at IS NULL
		JOIN posts p ON p.id = lk.post_id
		LEFT JOIN users u ON p.author_id = u.id
		WHERE lk.user_id = target.id AND lk.deleted_at IS NULL AND p.deleted_at IS NULL
		  AND (
		    p.visibility = 'public'
		    OR (p.visibility = 'follower' AND (
		      p.author_id = $4
		      OR EXISTS (SELECT 1 FROM follows f WHERE f.follower_id = $4 AND f.following_id = p.author_id)
		    ))
		    OR (p.visibility = 'private' AND p.author_id = $4)
		  )
		ORDER BY lk.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, handle, limit, offset, viewerID)
	if err != nil {
		return nil, err
	}
	return r.scanPostRows(rows, true)
}

func (r *postRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE posts SET view_count = view_count + 1 WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}
	return nil
}

func (r *postRepository) IncrementViewCountBatch(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.pool.Exec(ctx, `UPDATE posts SET view_count = view_count + 1 WHERE id = ANY($1)`, ids)
	if err != nil {
		return fmt.Errorf("failed to batch increment view count: %w", err)
	}
	return nil
}

func (r *postRepository) Update(ctx context.Context, id uuid.UUID, content string, visibility model.Visibility, locationLat *float64, locationLng *float64, locationName *string) error {
	query := `UPDATE posts SET content = $1, visibility = $2, location_lat = $3, location_lng = $4, location_name = $5, updated_at = NOW() WHERE id = $6 AND deleted_at IS NULL`
	result, err := r.pool.Exec(ctx, query, content, string(visibility), locationLat, locationLng, locationName, id)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *postRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `UPDATE posts SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete post: %w", err)
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *postRepository) ExistsIncludingDeleted(ctx context.Context, id uuid.UUID) (bool, bool, error) {
	var exists, isDeleted bool
	query := `
		SELECT
			EXISTS(SELECT 1 FROM posts WHERE id = $1) AS exists,
			EXISTS(SELECT 1 FROM posts WHERE id = $1 AND deleted_at IS NOT NULL) AS is_deleted`
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists, &isDeleted)
	if err != nil {
		return false, false, fmt.Errorf("failed to check post existence: %w", err)
	}
	return exists, isDeleted, nil
}

func (r *postRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*model.PostWithAuthor, error) {
	p := &model.PostWithAuthor{}
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at, p.deleted_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       p.location_lat, p.location_lng, p.location_name
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.id = $1`

	var visibility string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
		&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
		&p.AuthorDeleted,
		&p.LocationLat, &p.LocationLng, &p.LocationName,
	)
	if err != nil {
		return nil, err
	}
	p.Visibility = model.Visibility(visibility)
	return p, nil
}

func (r *postRepository) FindDeletedByAuthor(ctx context.Context, authorID uuid.UUID, limit int, cursor *time.Time) ([]model.PostWithAuthor, error) {
	query := `
		SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility, p.like_count, p.reply_count, p.view_count, p.repost_count, p.created_at, p.updated_at, p.deleted_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
		       (u.deleted_at IS NOT NULL OR u.id IS NULL),
		       p.location_lat, p.location_lng, p.location_name
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.author_id = $1
		  AND p.deleted_at IS NOT NULL
		  AND ($2::TIMESTAMPTZ IS NULL OR p.deleted_at < $2)
		ORDER BY p.deleted_at DESC
		LIMIT $3`

	rows, err := r.pool.Query(ctx, query, authorID, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.PostWithAuthor
	for rows.Next() {
		var p model.PostWithAuthor
		var visibility string
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.ParentID, &p.Content, &visibility, &p.LikeCount, &p.ReplyCount, &p.ViewCount, &p.RepostCount, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
			&p.AuthorUsername, &p.AuthorDisplayName, &p.AuthorProfileImageURL,
			&p.AuthorDeleted,
			&p.LocationLat, &p.LocationLng, &p.LocationName,
		); err != nil {
			return nil, err
		}
		p.Visibility = model.Visibility(visibility)
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) Restore(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `UPDATE posts SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return fmt.Errorf("failed to restore post: %w", err)
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *postRepository) RestoreReply(ctx context.Context, id uuid.UUID, parentID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `UPDATE posts SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return fmt.Errorf("failed to restore reply: %w", err)
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	_, err = tx.Exec(ctx,
		`UPDATE posts SET reply_count = reply_count + 1 WHERE id = $1`,
		parentID,
	)
	if err != nil {
		return fmt.Errorf("failed to increment reply_count: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *postRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM posts WHERE id = $1 AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete post: %w", err)
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *postRepository) SoftDeleteReply(ctx context.Context, id uuid.UUID, parentID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `UPDATE posts SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete reply: %w", err)
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	_, err = tx.Exec(ctx,
		`UPDATE posts SET reply_count = GREATEST(reply_count - 1, 0) WHERE id = $1`,
		parentID,
	)
	if err != nil {
		return fmt.Errorf("failed to decrement reply_count: %w", err)
	}

	return tx.Commit(ctx)
}
