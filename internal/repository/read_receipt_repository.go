package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type ReadReceiptRepository struct {
	db *sqlx.DB
}

func NewReadReceiptRepository(db *sqlx.DB) *ReadReceiptRepository {
	return &ReadReceiptRepository{db: db}
}

func (r *ReadReceiptRepository) Upsert(ctx context.Context, receipt *models.ReadReceipt) error {
	query := `INSERT INTO channel_read_receipts (id, channel_id, user_id, message_id, read_at) VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE message_id = VALUES(message_id), read_at = VALUES(read_at)`
	_, err := r.db.ExecContext(ctx, query, receipt.ID, receipt.ChannelID, receipt.UserID, receipt.MessageID, receipt.ReadAt)
	return err
}

func (r *ReadReceiptRepository) GetByChannelAndUser(ctx context.Context, channelID, userID string) (*models.ReadReceipt, error) {
	var receipt models.ReadReceipt
	err := r.db.GetContext(ctx, &receipt, `SELECT * FROM channel_read_receipts WHERE channel_id = ? AND user_id = ?`, channelID, userID)
	if err != nil {
		return nil, nil
	}
	return &receipt, nil
}

func (r *ReadReceiptRepository) ListByMessage(ctx context.Context, channelID, messageID string) ([]*models.ReadReceipt, error) {
	var receipts []*models.ReadReceipt
	err := r.db.SelectContext(ctx, &receipts, `SELECT * FROM channel_read_receipts WHERE channel_id = ? AND message_id >= ? ORDER BY read_at DESC`, channelID, messageID)
	return receipts, err
}

func (r *ReadReceiptRepository) GetReadCount(ctx context.Context, channelID, messageID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM channel_read_receipts WHERE channel_id = ? AND message_id >= ?`, channelID, messageID)
	return count, err
}
