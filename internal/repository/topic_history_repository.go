package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type TopicHistoryRepository struct {
	db *sqlx.DB
}

func NewTopicHistoryRepository(db *sqlx.DB) *TopicHistoryRepository {
	return &TopicHistoryRepository{db: db}
}

func (r *TopicHistoryRepository) Create(ctx context.Context, entry *models.TopicHistory) error {
	query := `INSERT INTO channel_topic_history (id, channel_id, old_topic, new_topic, changed_by, changed_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.ChannelID, entry.OldTopic, entry.NewTopic, entry.ChangedBy, entry.ChangedAt)
	return err
}

func (r *TopicHistoryRepository) ListByChannel(ctx context.Context, channelID string, limit, offset int) ([]*models.TopicHistory, error) {
	var entries []*models.TopicHistory
	query := `SELECT * FROM channel_topic_history WHERE channel_id = ? ORDER BY changed_at DESC LIMIT ? OFFSET ?`
	err := r.db.SelectContext(ctx, &entries, query, channelID, limit, offset)
	return entries, err
}

func (r *TopicHistoryRepository) CountByChannel(ctx context.Context, channelID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM channel_topic_history WHERE channel_id = ?`
	err := r.db.GetContext(ctx, &count, query, channelID)
	return count, err
}
