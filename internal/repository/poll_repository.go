package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type PollRepository struct {
	db *sqlx.DB
}

func NewPollRepository(db *sqlx.DB) *PollRepository {
	return &PollRepository{db: db}
}

func (r *PollRepository) Create(ctx context.Context, poll *models.ChannelPoll) error {
	query := `INSERT INTO channel_polls (id, channel_id, created_by, question, is_anonymous, multi_choice, is_closed, expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, poll.ID, poll.ChannelID, poll.CreatedBy, poll.Question, poll.IsAnonymous, poll.MultiChoice, poll.IsClosed, poll.ExpiresAt, poll.CreatedAt, poll.UpdatedAt)
	return err
}

func (r *PollRepository) CreateOption(ctx context.Context, option *models.PollOption) error {
	query := `INSERT INTO poll_options (id, poll_id, text, position) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, option.ID, option.PollID, option.Text, option.Position)
	return err
}

func (r *PollRepository) GetByID(ctx context.Context, id string) (*models.ChannelPoll, error) {
	var poll models.ChannelPoll
	err := r.db.GetContext(ctx, &poll, "SELECT * FROM channel_polls WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &poll, err
}

func (r *PollRepository) GetOptions(ctx context.Context, pollID string) ([]models.PollOption, error) {
	var options []models.PollOption
	err := r.db.SelectContext(ctx, &options, "SELECT * FROM poll_options WHERE poll_id = ? ORDER BY position ASC", pollID)
	return options, err
}

func (r *PollRepository) ListByChannel(ctx context.Context, channelID string) ([]*models.ChannelPoll, error) {
	var polls []*models.ChannelPoll
	err := r.db.SelectContext(ctx, &polls, "SELECT * FROM channel_polls WHERE channel_id = ? ORDER BY created_at DESC", channelID)
	return polls, err
}

func (r *PollRepository) CreateVote(ctx context.Context, vote *models.PollVote) error {
	query := `INSERT INTO poll_votes (id, poll_id, option_id, user_id, voted_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, vote.ID, vote.PollID, vote.OptionID, vote.UserID, vote.VotedAt)
	return err
}

func (r *PollRepository) HasVoted(ctx context.Context, pollID, userID string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM poll_votes WHERE poll_id = ? AND user_id = ?", pollID, userID)
	return count > 0, err
}

func (r *PollRepository) ClosePoll(ctx context.Context, pollID string, closedAt interface{}) error {
	_, err := r.db.ExecContext(ctx, "UPDATE channel_polls SET is_closed = TRUE, closed_at = ?, updated_at = NOW() WHERE id = ?", closedAt, pollID)
	return err
}

func (r *PollRepository) GetResults(ctx context.Context, pollID string) ([]models.PollResult, error) {
	var results []models.PollResult
	query := `SELECT po.id AS option_id, po.text AS option_text, COUNT(pv.id) AS vote_count
		FROM poll_options po
		LEFT JOIN poll_votes pv ON pv.option_id = po.id
		WHERE po.poll_id = ?
		GROUP BY po.id, po.text, po.position
		ORDER BY po.position ASC`
	err := r.db.SelectContext(ctx, &results, query, pollID)
	return results, err
}

func (r *PollRepository) GetTotalVotes(ctx context.Context, pollID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM poll_votes WHERE poll_id = ?", pollID)
	return count, err
}
