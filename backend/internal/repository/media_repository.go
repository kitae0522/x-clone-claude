package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type MediaRepository interface {
	Create(ctx context.Context, media *model.Media) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Media, error)
	FindByPostID(ctx context.Context, postID uuid.UUID) ([]model.Media, error)
	LinkToPost(ctx context.Context, mediaIDs []uuid.UUID, postID uuid.UUID) error
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Media, error)
	UnlinkByPostID(ctx context.Context, postID uuid.UUID) error
}

type mediaRepository struct {
	pool *pgxpool.Pool
}

func NewMediaRepository(pool *pgxpool.Pool) MediaRepository {
	return &mediaRepository{pool: pool}
}

func (r *mediaRepository) Create(ctx context.Context, media *model.Media) error {
	if media.ID != uuid.Nil {
		// Insert with explicit ID (media-service assigned)
		query := `
			INSERT INTO post_media (id, post_id, uploader_id, url, media_type, mime_type, width, height, size_bytes, duration_seconds, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING created_at`

		return r.pool.QueryRow(ctx, query,
			media.ID,
			media.PostID,
			media.UploaderID,
			media.URL,
			string(media.MediaType),
			media.MimeType,
			media.Width,
			media.Height,
			media.SizeBytes,
			media.DurationSeconds,
			media.SortOrder,
		).Scan(&media.CreatedAt)
	}

	// Auto-generate ID (legacy local upload)
	query := `
		INSERT INTO post_media (uploader_id, url, media_type, mime_type, width, height, size_bytes, duration_seconds, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		media.UploaderID,
		media.URL,
		string(media.MediaType),
		media.MimeType,
		media.Width,
		media.Height,
		media.SizeBytes,
		media.DurationSeconds,
		media.SortOrder,
	).Scan(&media.ID, &media.CreatedAt)
}

func (r *mediaRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Media, error) {
	m := &model.Media{}
	query := `
		SELECT id, post_id, uploader_id, url, media_type, mime_type, width, height, size_bytes, duration_seconds, sort_order, created_at
		FROM post_media
		WHERE id = $1`

	var mediaType string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.PostID, &m.UploaderID, &m.URL, &mediaType, &m.MimeType,
		&m.Width, &m.Height, &m.SizeBytes, &m.DurationSeconds, &m.SortOrder, &m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find media by id: %w", err)
	}
	m.MediaType = model.MediaType(mediaType)
	return m, nil
}

func (r *mediaRepository) FindByPostID(ctx context.Context, postID uuid.UUID) ([]model.Media, error) {
	query := `
		SELECT id, post_id, uploader_id, url, media_type, mime_type, width, height, size_bytes, duration_seconds, sort_order, created_at
		FROM post_media
		WHERE post_id = $1
		ORDER BY sort_order ASC`

	rows, err := r.pool.Query(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to find media by post id: %w", err)
	}
	defer rows.Close()

	var media []model.Media
	for rows.Next() {
		var m model.Media
		var mediaType string
		if err := rows.Scan(
			&m.ID, &m.PostID, &m.UploaderID, &m.URL, &mediaType, &m.MimeType,
			&m.Width, &m.Height, &m.SizeBytes, &m.DurationSeconds, &m.SortOrder, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan media row: %w", err)
		}
		m.MediaType = model.MediaType(mediaType)
		media = append(media, m)
	}
	return media, rows.Err()
}

func (r *mediaRepository) LinkToPost(ctx context.Context, mediaIDs []uuid.UUID, postID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for i, id := range mediaIDs {
		_, err := tx.Exec(ctx,
			`UPDATE post_media SET post_id = $1, sort_order = $2 WHERE id = $3`,
			postID, i, id,
		)
		if err != nil {
			return fmt.Errorf("failed to link media %s to post: %w", id, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *mediaRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Media, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, post_id, uploader_id, url, media_type, mime_type, width, height, size_bytes, duration_seconds, sort_order, created_at
		FROM post_media
		WHERE id = ANY($1)
		ORDER BY sort_order ASC`

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to find media by ids: %w", err)
	}
	defer rows.Close()

	var media []model.Media
	for rows.Next() {
		var m model.Media
		var mediaType string
		if err := rows.Scan(
			&m.ID, &m.PostID, &m.UploaderID, &m.URL, &mediaType, &m.MimeType,
			&m.Width, &m.Height, &m.SizeBytes, &m.DurationSeconds, &m.SortOrder, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan media row: %w", err)
		}
		m.MediaType = model.MediaType(mediaType)
		media = append(media, m)
	}
	return media, rows.Err()
}

func (r *mediaRepository) UnlinkByPostID(ctx context.Context, postID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM post_media WHERE post_id = $1`, postID)
	if err != nil {
		return fmt.Errorf("failed to unlink media by post_id: %w", err)
	}
	return nil
}
