package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type ModerationRepository struct {
	db *sqlx.DB
}

func NewModerationRepository(db *sqlx.DB) *ModerationRepository {
	return &ModerationRepository{db: db}
}

// ── Bans ──

func (r *ModerationRepository) CreateBan(ctx context.Context, ban *models.ChannelBan) error {
	query := `INSERT INTO channel_bans (id, channel_id, user_id, banned_by, reason, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		ban.ID, ban.ChannelID, ban.UserID, ban.BannedBy, ban.Reason, ban.ExpiresAt, ban.CreatedAt)
	return err
}

func (r *ModerationRepository) RemoveBan(ctx context.Context, channelID, userID string) error {
	query := `DELETE FROM channel_bans WHERE channel_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, channelID, userID)
	return err
}

func (r *ModerationRepository) IsBanned(ctx context.Context, channelID, userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM channel_bans WHERE channel_id = ? AND user_id = ? AND (expires_at IS NULL OR expires_at > NOW())`
	err := r.db.GetContext(ctx, &count, query, channelID, userID)
	return count > 0, err
}

func (r *ModerationRepository) ListBans(ctx context.Context, channelID string) ([]*models.ChannelBan, error) {
	var bans []*models.ChannelBan
	query := `SELECT * FROM channel_bans WHERE channel_id = ? ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &bans, query, channelID)
	return bans, err
}

func (r *ModerationRepository) GetBan(ctx context.Context, channelID, userID string) (*models.ChannelBan, error) {
	var ban models.ChannelBan
	query := `SELECT * FROM channel_bans WHERE channel_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &ban, query, channelID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &ban, err
}

// ── Mutes ──

func (r *ModerationRepository) CreateMute(ctx context.Context, mute *models.ChannelMute) error {
	query := `INSERT INTO channel_mutes (id, channel_id, user_id, muted_by, reason, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		mute.ID, mute.ChannelID, mute.UserID, mute.MutedBy, mute.Reason, mute.ExpiresAt, mute.CreatedAt)
	return err
}

func (r *ModerationRepository) RemoveMute(ctx context.Context, channelID, userID string) error {
	query := `DELETE FROM channel_mutes WHERE channel_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, channelID, userID)
	return err
}

func (r *ModerationRepository) IsMuted(ctx context.Context, channelID, userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM channel_mutes WHERE channel_id = ? AND user_id = ? AND (expires_at IS NULL OR expires_at > NOW())`
	err := r.db.GetContext(ctx, &count, query, channelID, userID)
	return count > 0, err
}

func (r *ModerationRepository) ListMutes(ctx context.Context, channelID string) ([]*models.ChannelMute, error) {
	var mutes []*models.ChannelMute
	query := `SELECT * FROM channel_mutes WHERE channel_id = ? ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &mutes, query, channelID)
	return mutes, err
}

// ── Moderation History ──

func (r *ModerationRepository) LogAction(ctx context.Context, entry *models.ModerationEntry) error {
	query := `INSERT INTO channel_moderation_log (id, channel_id, user_id, action, actor_id, reason, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.ChannelID, entry.UserID, entry.Action, entry.ActorID,
		entry.Reason, entry.ExpiresAt, entry.CreatedAt)
	return err
}

func (r *ModerationRepository) GetHistory(ctx context.Context, channelID string, limit, offset int) ([]*models.ModerationEntry, error) {
	var entries []*models.ModerationEntry
	query := `SELECT * FROM channel_moderation_log WHERE channel_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err := r.db.SelectContext(ctx, &entries, query, channelID, limit, offset)
	return entries, err
}
