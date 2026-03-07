package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepostRepository interface {
	Repost(ctx context.Context, userID, postID uuid.UUID) error
	Unrepost(ctx context.Context, userID, postID uuid.UUID) error
	IsReposted(ctx context.Context, userID, postID uuid.UUID) (bool, error)
}

type repostRepository struct {
	pool *pgxpool.Pool
}

func NewRepostRepository(pool *pgxpool.Pool) RepostRepository {
	return &repostRepository{pool: pool}
}

func (r *repostRepository) Repost(ctx context.Context, userID, postID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO reposts (user_id, post_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, postID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert repost: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE posts SET repost_count = repost_count + 1 WHERE id = $1`,
		postID,
	)
	if err != nil {
		return fmt.Errorf("failed to update repost_count: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *repostRepository) Unrepost(ctx context.Context, userID, postID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`DELETE FROM reposts WHERE user_id = $1 AND post_id = $2`,
		userID, postID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete repost: %w", err)
	}

	if tag.RowsAffected() > 0 {
		_, err = tx.Exec(ctx,
			`UPDATE posts SET repost_count = GREATEST(repost_count - 1, 0) WHERE id = $1`,
			postID,
		)
		if err != nil {
			return fmt.Errorf("failed to update repost_count: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *repostRepository) IsReposted(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM reposts WHERE user_id = $1 AND post_id = $2)`,
		userID, postID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check repost status: %w", err)
	}
	return exists, nil
}
