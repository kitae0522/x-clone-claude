package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LikeRepository interface {
	Like(ctx context.Context, userID, postID uuid.UUID) error
	Unlike(ctx context.Context, userID, postID uuid.UUID) error
	IsLiked(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	SoftDeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

type likeRepository struct {
	pool *pgxpool.Pool
}

func NewLikeRepository(pool *pgxpool.Pool) LikeRepository {
	return &likeRepository{pool: pool}
}

func (r *likeRepository) Like(ctx context.Context, userID, postID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`INSERT INTO likes (user_id, post_id)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id, post_id)
		 DO UPDATE SET deleted_at = NULL, created_at = NOW()
		 WHERE likes.deleted_at IS NOT NULL`,
		userID, postID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert like: %w", err)
	}

	if tag.RowsAffected() > 0 {
		_, err = tx.Exec(ctx,
			`UPDATE posts SET like_count = like_count + 1 WHERE id = $1`,
			postID,
		)
		if err != nil {
			return fmt.Errorf("failed to update like_count: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *likeRepository) Unlike(ctx context.Context, userID, postID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`DELETE FROM likes WHERE user_id = $1 AND post_id = $2 AND deleted_at IS NULL`,
		userID, postID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete like: %w", err)
	}

	if tag.RowsAffected() > 0 {
		_, err = tx.Exec(ctx,
			`UPDATE posts SET like_count = like_count - 1 WHERE id = $1`,
			postID,
		)
		if err != nil {
			return fmt.Errorf("failed to update like_count: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *likeRepository) IsLiked(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND post_id = $2 AND deleted_at IS NULL)`,
		userID, postID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check like status: %w", err)
	}
	return exists, nil
}

func (r *likeRepository) SoftDeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		WITH deleted_likes AS (
			UPDATE likes
			SET deleted_at = NOW()
			WHERE user_id = $1 AND deleted_at IS NULL
			RETURNING post_id
		)
		UPDATE posts p
		SET like_count = GREATEST(like_count - sub.cnt, 0)
		FROM (
			SELECT post_id, COUNT(*) AS cnt
			FROM deleted_likes
			GROUP BY post_id
		) sub
		WHERE p.id = sub.post_id`,
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to soft delete likes: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return tag.RowsAffected(), nil
}
