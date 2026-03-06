package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitae0522/twitter-clone-claude/backend/internal/model"
)

type PollRepository interface {
	CreatePoll(ctx context.Context, poll *model.Poll, options []model.PollOption) error
	FindByPostID(ctx context.Context, postID uuid.UUID) (*model.Poll, []model.PollOption, error)
	Vote(ctx context.Context, pollID, userID uuid.UUID, optionIndex int16) error
	Unvote(ctx context.Context, pollID, userID uuid.UUID, optionIndex int16) error
	GetUserVote(ctx context.Context, pollID, userID uuid.UUID) (*int16, error)
	FindByPostIDs(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]model.Poll, map[uuid.UUID][]model.PollOption, error)
}

type pollRepository struct {
	pool *pgxpool.Pool
}

func NewPollRepository(pool *pgxpool.Pool) PollRepository {
	return &pollRepository{pool: pool}
}

func (r *pollRepository) CreatePoll(ctx context.Context, poll *model.Poll, options []model.PollOption) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	pollQuery := `
		INSERT INTO polls (post_id, expires_at)
		VALUES ($1, $2)
		RETURNING id, created_at`

	err = tx.QueryRow(ctx, pollQuery, poll.PostID, poll.ExpiresAt).Scan(&poll.ID, &poll.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert poll: %w", err)
	}

	optionQuery := `
		INSERT INTO poll_options (poll_id, option_index, text)
		VALUES ($1, $2, $3)
		RETURNING id`

	for i := range options {
		options[i].PollID = poll.ID
		err = tx.QueryRow(ctx, optionQuery, options[i].PollID, options[i].OptionIndex, options[i].Text).Scan(&options[i].ID)
		if err != nil {
			return fmt.Errorf("failed to insert poll option: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *pollRepository) FindByPostID(ctx context.Context, postID uuid.UUID) (*model.Poll, []model.PollOption, error) {
	poll := &model.Poll{}
	pollQuery := `
		SELECT id, post_id, expires_at, total_votes, created_at
		FROM polls
		WHERE post_id = $1`

	err := r.pool.QueryRow(ctx, pollQuery, postID).Scan(
		&poll.ID, &poll.PostID, &poll.ExpiresAt, &poll.TotalVotes, &poll.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to find poll: %w", err)
	}

	optionsQuery := `
		SELECT id, poll_id, option_index, text, vote_count
		FROM poll_options
		WHERE poll_id = $1
		ORDER BY option_index ASC`

	rows, err := r.pool.Query(ctx, optionsQuery, poll.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find poll options: %w", err)
	}
	defer rows.Close()

	var options []model.PollOption
	for rows.Next() {
		var opt model.PollOption
		if err := rows.Scan(&opt.ID, &opt.PollID, &opt.OptionIndex, &opt.Text, &opt.VoteCount); err != nil {
			return nil, nil, fmt.Errorf("failed to scan poll option: %w", err)
		}
		options = append(options, opt)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to iterate poll options: %w", err)
	}

	return poll, options, nil
}

func (r *pollRepository) Vote(ctx context.Context, pollID, userID uuid.UUID, optionIndex int16) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	voteQuery := `
		INSERT INTO poll_votes (poll_id, user_id, option_index)
		VALUES ($1, $2, $3)`

	_, err = tx.Exec(ctx, voteQuery, pollID, userID, optionIndex)
	if err != nil {
		return fmt.Errorf("failed to insert vote: %w", err)
	}

	optionUpdateQuery := `
		UPDATE poll_options
		SET vote_count = vote_count + 1
		WHERE poll_id = $1 AND option_index = $2`

	_, err = tx.Exec(ctx, optionUpdateQuery, pollID, optionIndex)
	if err != nil {
		return fmt.Errorf("failed to update option vote count: %w", err)
	}

	pollUpdateQuery := `
		UPDATE polls
		SET total_votes = total_votes + 1
		WHERE id = $1`

	_, err = tx.Exec(ctx, pollUpdateQuery, pollID)
	if err != nil {
		return fmt.Errorf("failed to update poll total votes: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *pollRepository) Unvote(ctx context.Context, pollID, userID uuid.UUID, optionIndex int16) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	deleteQuery := `
		DELETE FROM poll_votes
		WHERE poll_id = $1 AND user_id = $2`

	res, err := tx.Exec(ctx, deleteQuery, pollID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete vote: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("vote not found")
	}

	optionUpdateQuery := `
		UPDATE poll_options
		SET vote_count = vote_count - 1
		WHERE poll_id = $1 AND option_index = $2`

	_, err = tx.Exec(ctx, optionUpdateQuery, pollID, optionIndex)
	if err != nil {
		return fmt.Errorf("failed to update option vote count: %w", err)
	}

	pollUpdateQuery := `
		UPDATE polls
		SET total_votes = total_votes - 1
		WHERE id = $1`

	_, err = tx.Exec(ctx, pollUpdateQuery, pollID)
	if err != nil {
		return fmt.Errorf("failed to update poll total votes: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *pollRepository) GetUserVote(ctx context.Context, pollID, userID uuid.UUID) (*int16, error) {
	query := `
		SELECT option_index
		FROM poll_votes
		WHERE poll_id = $1 AND user_id = $2`

	var optionIndex int16
	err := r.pool.QueryRow(ctx, query, pollID, userID).Scan(&optionIndex)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user vote: %w", err)
	}
	return &optionIndex, nil
}

func (r *pollRepository) FindByPostIDs(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]model.Poll, map[uuid.UUID][]model.PollOption, error) {
	if len(postIDs) == 0 {
		return make(map[uuid.UUID]model.Poll), make(map[uuid.UUID][]model.PollOption), nil
	}

	pollsQuery := `
		SELECT id, post_id, expires_at, total_votes, created_at
		FROM polls
		WHERE post_id = ANY($1)`

	rows, err := r.pool.Query(ctx, pollsQuery, postIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find polls: %w", err)
	}
	defer rows.Close()

	pollsByPostID := make(map[uuid.UUID]model.Poll)
	var pollIDs []uuid.UUID
	for rows.Next() {
		var p model.Poll
		if err := rows.Scan(&p.ID, &p.PostID, &p.ExpiresAt, &p.TotalVotes, &p.CreatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan poll: %w", err)
		}
		pollsByPostID[p.PostID] = p
		pollIDs = append(pollIDs, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to iterate polls: %w", err)
	}

	optionsByPollID := make(map[uuid.UUID][]model.PollOption)
	if len(pollIDs) > 0 {
		optionsQuery := `
			SELECT id, poll_id, option_index, text, vote_count
			FROM poll_options
			WHERE poll_id = ANY($1)
			ORDER BY option_index ASC`

		optRows, err := r.pool.Query(ctx, optionsQuery, pollIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to find poll options: %w", err)
		}
		defer optRows.Close()

		for optRows.Next() {
			var opt model.PollOption
			if err := optRows.Scan(&opt.ID, &opt.PollID, &opt.OptionIndex, &opt.Text, &opt.VoteCount); err != nil {
				return nil, nil, fmt.Errorf("failed to scan poll option: %w", err)
			}
			optionsByPollID[opt.PollID] = append(optionsByPollID[opt.PollID], opt)
		}
		if err := optRows.Err(); err != nil {
			return nil, nil, fmt.Errorf("failed to iterate poll options: %w", err)
		}
	}

	return pollsByPostID, optionsByPollID, nil
}
