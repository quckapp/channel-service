package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type AnalyticsRepository struct {
	db *sqlx.DB
}

func NewAnalyticsRepository(db *sqlx.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) GetChannelStats(ctx context.Context, channelID string) (*models.ChannelStats, error) {
	stats := &models.ChannelStats{}

	// Member count
	memberQuery := `SELECT COUNT(*) FROM channel_members WHERE channel_id = ?`
	r.db.GetContext(ctx, &stats.MemberCount, memberQuery, channelID)

	// Pin count
	pinQuery := `SELECT COUNT(*) FROM channel_pins WHERE channel_id = ?`
	r.db.GetContext(ctx, &stats.PinCount, pinQuery, channelID)

	// Active members this week (members who updated last_read_at in last 7 days)
	activeQuery := `SELECT COUNT(*) FROM channel_members WHERE channel_id = ? AND last_read_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)`
	r.db.GetContext(ctx, &stats.ActiveMembersWeek, activeQuery, channelID)

	return stats, nil
}

func (r *AnalyticsRepository) GetDailyActivity(ctx context.Context, channelID string, days int) ([]models.ChannelActivity, error) {
	var activities []models.ChannelActivity
	query := `SELECT DATE(last_read_at) as date, COUNT(DISTINCT user_id) as active_users, 0 as message_count
		FROM channel_members
		WHERE channel_id = ? AND last_read_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY DATE(last_read_at)
		ORDER BY date DESC`
	err := r.db.SelectContext(ctx, &activities, query, channelID, days)
	return activities, err
}

func (r *AnalyticsRepository) GetMostActiveMembers(ctx context.Context, channelID string, limit int) ([]models.MostActiveMember, error) {
	var members []models.MostActiveMember
	// Using reaction count as activity proxy since we don't have messages table
	query := `SELECT user_id, COUNT(*) as message_count FROM channel_reactions
		WHERE channel_id = ? GROUP BY user_id ORDER BY message_count DESC LIMIT ?`
	err := r.db.SelectContext(ctx, &members, query, channelID, limit)
	return members, err
}

func (r *AnalyticsRepository) GetTopChannels(ctx context.Context, workspaceID string, limit int) ([]*models.Channel, error) {
	var channels []*models.Channel
	query := `SELECT c.* FROM channels c
		LEFT JOIN channel_members cm ON c.id = cm.channel_id
		WHERE c.workspace_id = ? AND c.deleted_at IS NULL
		GROUP BY c.id
		ORDER BY COUNT(cm.id) DESC
		LIMIT ?`
	err := r.db.SelectContext(ctx, &channels, query, workspaceID, limit)
	return channels, err
}
