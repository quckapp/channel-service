package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type StarredRepository struct {
	db *sqlx.DB
}

func NewStarredRepository(db *sqlx.DB) *StarredRepository {
	return &StarredRepository{db: db}
}

func (r *StarredRepository) Star(ctx context.Context, star *models.StarredChannel) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO starred_channels (id, user_id, channel_id, position, created_at) VALUES (?, ?, ?, ?, ?)`,
		star.ID, star.UserID, star.ChannelID, star.Position, star.CreatedAt)
	return err
}

func (r *StarredRepository) Unstar(ctx context.Context, userID, channelID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM starred_channels WHERE user_id = ? AND channel_id = ?`, userID, channelID)
	return err
}

func (r *StarredRepository) IsStarred(ctx context.Context, userID, channelID string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM starred_channels WHERE user_id = ? AND channel_id = ?`, userID, channelID)
	return count > 0, err
}

func (r *StarredRepository) ListByUser(ctx context.Context, userID string) ([]*models.StarredChannel, error) {
	var starred []*models.StarredChannel
	err := r.db.SelectContext(ctx, &starred, `SELECT * FROM starred_channels WHERE user_id = ? ORDER BY position`, userID)
	return starred, err
}

func (r *StarredRepository) GetMaxPosition(ctx context.Context, userID string) (int, error) {
	var maxPos sql.NullInt64
	err := r.db.GetContext(ctx, &maxPos, `SELECT MAX(position) FROM starred_channels WHERE user_id = ?`, userID)
	if !maxPos.Valid {
		return 0, err
	}
	return int(maxPos.Int64), err
}

func (r *StarredRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM starred_channels WHERE user_id = ?`, userID)
	return count, err
}
