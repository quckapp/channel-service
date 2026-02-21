package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type ChannelLinkRepository struct {
	db *sqlx.DB
}

func NewChannelLinkRepository(db *sqlx.DB) *ChannelLinkRepository {
	return &ChannelLinkRepository{db: db}
}

func (r *ChannelLinkRepository) Create(ctx context.Context, link *models.ChannelLink) error {
	query := `INSERT INTO channel_links (id, source_channel_id, target_channel_id, created_by, link_type, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, link.ID, link.SourceChannelID, link.TargetChannelID, link.CreatedBy, link.LinkType, link.IsActive, link.CreatedAt, link.UpdatedAt)
	return err
}

func (r *ChannelLinkRepository) GetByID(ctx context.Context, id string) (*models.ChannelLink, error) {
	var link models.ChannelLink
	err := r.db.GetContext(ctx, &link, "SELECT * FROM channel_links WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &link, err
}

func (r *ChannelLinkRepository) ListByChannel(ctx context.Context, channelID string) ([]*models.ChannelLink, error) {
	var links []*models.ChannelLink
	err := r.db.SelectContext(ctx, &links, "SELECT * FROM channel_links WHERE (source_channel_id = ? OR target_channel_id = ?) AND is_active = TRUE ORDER BY created_at DESC", channelID, channelID)
	return links, err
}

func (r *ChannelLinkRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE channel_links SET is_active = FALSE, updated_at = NOW() WHERE id = ?", id)
	return err
}
