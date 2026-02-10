package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type PinRepository struct {
	db *sqlx.DB
}

func NewPinRepository(db *sqlx.DB) *PinRepository {
	return &PinRepository{db: db}
}

func (r *PinRepository) Create(ctx context.Context, pin *models.ChannelPin) error {
	query := `INSERT INTO channel_pins (id, channel_id, message_id, pinned_by, pinned_at)
		VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, pin.ID, pin.ChannelID, pin.MessageID, pin.PinnedBy, pin.PinnedAt)
	return err
}

func (r *PinRepository) GetByChannelAndMessage(ctx context.Context, channelID, messageID string) (*models.ChannelPin, error) {
	var pin models.ChannelPin
	query := `SELECT * FROM channel_pins WHERE channel_id = ? AND message_id = ?`
	err := r.db.GetContext(ctx, &pin, query, channelID, messageID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &pin, err
}

func (r *PinRepository) ListByChannel(ctx context.Context, channelID string) ([]*models.ChannelPin, error) {
	var pins []*models.ChannelPin
	query := `SELECT * FROM channel_pins WHERE channel_id = ? ORDER BY pinned_at DESC`
	err := r.db.SelectContext(ctx, &pins, query, channelID)
	return pins, err
}

func (r *PinRepository) Delete(ctx context.Context, channelID, messageID string) error {
	query := `DELETE FROM channel_pins WHERE channel_id = ? AND message_id = ?`
	_, err := r.db.ExecContext(ctx, query, channelID, messageID)
	return err
}

func (r *PinRepository) Count(ctx context.Context, channelID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM channel_pins WHERE channel_id = ?`
	err := r.db.GetContext(ctx, &count, query, channelID)
	return count, err
}
