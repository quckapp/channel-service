package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type ActivityLogRepository struct {
	db *sqlx.DB
}

func NewActivityLogRepository(db *sqlx.DB) *ActivityLogRepository {
	return &ActivityLogRepository{db: db}
}

func (r *ActivityLogRepository) Create(ctx context.Context, entry *models.ChannelActivityLog) error {
	query := `INSERT INTO channel_activity_log (id, channel_id, user_id, action, target_id, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, entry.ID, entry.ChannelID, entry.UserID, entry.Action, entry.TargetID, entry.Details, entry.CreatedAt)
	return err
}

func (r *ActivityLogRepository) ListByChannel(ctx context.Context, channelID string, limit, offset int) ([]*models.ChannelActivityLog, error) {
	var entries []*models.ChannelActivityLog
	err := r.db.SelectContext(ctx, &entries, `SELECT * FROM channel_activity_log WHERE channel_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, channelID, limit, offset)
	return entries, err
}

func (r *ActivityLogRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.ChannelActivityLog, error) {
	var entries []*models.ChannelActivityLog
	err := r.db.SelectContext(ctx, &entries, `SELECT * FROM channel_activity_log WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, userID, limit, offset)
	return entries, err
}

func (r *ActivityLogRepository) ListByAction(ctx context.Context, channelID, action string, limit, offset int) ([]*models.ChannelActivityLog, error) {
	var entries []*models.ChannelActivityLog
	err := r.db.SelectContext(ctx, &entries, `SELECT * FROM channel_activity_log WHERE channel_id = ? AND action = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, channelID, action, limit, offset)
	return entries, err
}
