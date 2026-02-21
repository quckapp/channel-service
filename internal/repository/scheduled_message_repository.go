package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type ScheduledMessageRepository struct {
	db *sqlx.DB
}

func NewScheduledMessageRepository(db *sqlx.DB) *ScheduledMessageRepository {
	return &ScheduledMessageRepository{db: db}
}

func (r *ScheduledMessageRepository) Create(ctx context.Context, msg *models.ScheduledMessage) error {
	query := `INSERT INTO scheduled_messages (id, channel_id, user_id, content, scheduled_at, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, msg.ID, msg.ChannelID, msg.UserID, msg.Content, msg.ScheduledAt, msg.Status, msg.CreatedAt, msg.UpdatedAt)
	return err
}

func (r *ScheduledMessageRepository) GetByID(ctx context.Context, id string) (*models.ScheduledMessage, error) {
	var msg models.ScheduledMessage
	err := r.db.GetContext(ctx, &msg, "SELECT * FROM scheduled_messages WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &msg, err
}

func (r *ScheduledMessageRepository) ListByChannel(ctx context.Context, channelID string) ([]*models.ScheduledMessage, error) {
	var msgs []*models.ScheduledMessage
	err := r.db.SelectContext(ctx, &msgs, "SELECT * FROM scheduled_messages WHERE channel_id = ? AND status = 'pending' ORDER BY scheduled_at ASC", channelID)
	return msgs, err
}

func (r *ScheduledMessageRepository) Update(ctx context.Context, msg *models.ScheduledMessage) error {
	query := `UPDATE scheduled_messages SET content = ?, scheduled_at = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, msg.Content, msg.ScheduledAt, msg.ID)
	return err
}

func (r *ScheduledMessageRepository) Cancel(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE scheduled_messages SET status = 'cancelled', updated_at = NOW() WHERE id = ?", id)
	return err
}
