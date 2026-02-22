package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type ReactionRepository struct {
	db *sqlx.DB
}

func NewReactionRepository(db *sqlx.DB) *ReactionRepository {
	return &ReactionRepository{db: db}
}

func (r *ReactionRepository) Create(ctx context.Context, reaction *models.ChannelReaction) error {
	query := `INSERT INTO channel_reactions (id, channel_id, message_id, user_id, emoji, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		reaction.ID, reaction.ChannelID, reaction.MessageID, reaction.UserID, reaction.Emoji, reaction.CreatedAt)
	return err
}

func (r *ReactionRepository) Delete(ctx context.Context, channelID, messageID, userID, emoji string) error {
	query := `DELETE FROM channel_reactions WHERE channel_id = ? AND message_id = ? AND user_id = ? AND emoji = ?`
	_, err := r.db.ExecContext(ctx, query, channelID, messageID, userID, emoji)
	return err
}

func (r *ReactionRepository) Exists(ctx context.Context, channelID, messageID, userID, emoji string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM channel_reactions WHERE channel_id = ? AND message_id = ? AND user_id = ? AND emoji = ?`
	err := r.db.GetContext(ctx, &count, query, channelID, messageID, userID, emoji)
	return count > 0, err
}

func (r *ReactionRepository) ListByMessage(ctx context.Context, channelID, messageID string) ([]*models.ChannelReaction, error) {
	var reactions []*models.ChannelReaction
	query := `SELECT * FROM channel_reactions WHERE channel_id = ? AND message_id = ? ORDER BY created_at ASC`
	err := r.db.SelectContext(ctx, &reactions, query, channelID, messageID)
	return reactions, err
}

func (r *ReactionRepository) GetSummary(ctx context.Context, channelID, messageID string) ([]models.ReactionSummary, error) {
	var summary []models.ReactionSummary
	query := `SELECT emoji, COUNT(*) as count FROM channel_reactions WHERE channel_id = ? AND message_id = ? GROUP BY emoji ORDER BY count DESC`
	err := r.db.SelectContext(ctx, &summary, query, channelID, messageID)
	return summary, err
}

func (r *ReactionRepository) GetByID(ctx context.Context, id string) (*models.ChannelReaction, error) {
	var reaction models.ChannelReaction
	query := `SELECT * FROM channel_reactions WHERE id = ?`
	err := r.db.GetContext(ctx, &reaction, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &reaction, err
}
