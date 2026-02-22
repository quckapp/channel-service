package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type SettingsRepository struct {
	db *sqlx.DB
}

func NewSettingsRepository(db *sqlx.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Get(ctx context.Context, channelID string) (*models.ChannelSetting, error) {
	var setting models.ChannelSetting
	err := r.db.GetContext(ctx, &setting, `SELECT * FROM channel_settings WHERE channel_id = ?`, channelID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &setting, err
}

func (r *SettingsRepository) Upsert(ctx context.Context, setting *models.ChannelSetting) error {
	query := `INSERT INTO channel_settings (id, channel_id, slow_mode_interval, max_pins, max_bookmarks, allow_threads, allow_reactions, allow_invites, auto_archive_days, default_notification, custom_emoji, link_previews, member_limit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE slow_mode_interval = VALUES(slow_mode_interval), max_pins = VALUES(max_pins), max_bookmarks = VALUES(max_bookmarks), allow_threads = VALUES(allow_threads), allow_reactions = VALUES(allow_reactions), allow_invites = VALUES(allow_invites), auto_archive_days = VALUES(auto_archive_days), default_notification = VALUES(default_notification), custom_emoji = VALUES(custom_emoji), link_previews = VALUES(link_previews), member_limit = VALUES(member_limit), updated_at = ?`
	_, err := r.db.ExecContext(ctx, query, setting.ID, setting.ChannelID, setting.SlowModeInterval, setting.MaxPins, setting.MaxBookmarks, setting.AllowThreads, setting.AllowReactions, setting.AllowInvites, setting.AutoArchiveDays, setting.DefaultNotification, setting.CustomEmoji, setting.LinkPreviews, setting.MemberLimit, setting.CreatedAt, setting.UpdatedAt, time.Now())
	return err
}

func (r *SettingsRepository) Delete(ctx context.Context, channelID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM channel_settings WHERE channel_id = ?`, channelID)
	return err
}
