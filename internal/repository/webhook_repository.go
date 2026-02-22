package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type WebhookRepository struct {
	db *sqlx.DB
}

func NewWebhookRepository(db *sqlx.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) Create(ctx context.Context, webhook *models.ChannelWebhook) error {
	query := `INSERT INTO channel_webhooks (id, channel_id, name, url, avatar_url, events, is_active, created_by, last_triggered_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		webhook.ID, webhook.ChannelID, webhook.Name, webhook.URL, webhook.AvatarURL,
		webhook.Events, webhook.IsActive, webhook.CreatedBy, webhook.LastTriggeredAt,
		webhook.CreatedAt, webhook.UpdatedAt)
	return err
}

func (r *WebhookRepository) GetByID(ctx context.Context, id string) (*models.ChannelWebhook, error) {
	var webhook models.ChannelWebhook
	query := `SELECT * FROM channel_webhooks WHERE id = ?`
	err := r.db.GetContext(ctx, &webhook, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &webhook, err
}

func (r *WebhookRepository) ListByChannel(ctx context.Context, channelID string) ([]*models.ChannelWebhook, error) {
	var webhooks []*models.ChannelWebhook
	query := `SELECT * FROM channel_webhooks WHERE channel_id = ? ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &webhooks, query, channelID)
	return webhooks, err
}

func (r *WebhookRepository) Update(ctx context.Context, webhook *models.ChannelWebhook) error {
	query := `UPDATE channel_webhooks SET name = ?, url = ?, avatar_url = ?, events = ?, is_active = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query,
		webhook.Name, webhook.URL, webhook.AvatarURL, webhook.Events, webhook.IsActive, webhook.ID)
	return err
}

func (r *WebhookRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM channel_webhooks WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *WebhookRepository) UpdateLastTriggered(ctx context.Context, id string) error {
	query := `UPDATE channel_webhooks SET last_triggered_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *WebhookRepository) ListActiveByEvent(ctx context.Context, channelID, event string) ([]*models.ChannelWebhook, error) {
	var webhooks []*models.ChannelWebhook
	query := `SELECT * FROM channel_webhooks WHERE channel_id = ? AND is_active = TRUE AND events LIKE ?`
	pattern := "%" + event + "%"
	err := r.db.SelectContext(ctx, &webhooks, query, channelID, pattern)
	return webhooks, err
}
