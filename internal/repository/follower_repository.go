package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type FollowerRepository struct {
	db *sqlx.DB
}

func NewFollowerRepository(db *sqlx.DB) *FollowerRepository {
	return &FollowerRepository{db: db}
}

func (r *FollowerRepository) Follow(ctx context.Context, follower *models.ChannelFollower) error {
	query := `INSERT INTO channel_followers (id, channel_id, user_id, followed_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, follower.ID, follower.ChannelID, follower.UserID, follower.FollowedAt)
	return err
}

func (r *FollowerRepository) Unfollow(ctx context.Context, channelID, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM channel_followers WHERE channel_id = ? AND user_id = ?`, channelID, userID)
	return err
}

func (r *FollowerRepository) IsFollowing(ctx context.Context, channelID, userID string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM channel_followers WHERE channel_id = ? AND user_id = ?`, channelID, userID)
	return count > 0, err
}

func (r *FollowerRepository) ListByChannel(ctx context.Context, channelID string, limit, offset int) ([]*models.ChannelFollower, error) {
	var followers []*models.ChannelFollower
	err := r.db.SelectContext(ctx, &followers, `SELECT * FROM channel_followers WHERE channel_id = ? ORDER BY followed_at DESC LIMIT ? OFFSET ?`, channelID, limit, offset)
	return followers, err
}

func (r *FollowerRepository) ListByUser(ctx context.Context, userID string) ([]*models.ChannelFollower, error) {
	var followers []*models.ChannelFollower
	err := r.db.SelectContext(ctx, &followers, `SELECT * FROM channel_followers WHERE user_id = ? ORDER BY followed_at DESC`, userID)
	return followers, err
}

func (r *FollowerRepository) CountByChannel(ctx context.Context, channelID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM channel_followers WHERE channel_id = ?`, channelID)
	return count, err
}

func (r *FollowerRepository) GetByChannelAndUser(ctx context.Context, channelID, userID string) (*models.ChannelFollower, error) {
	var follower models.ChannelFollower
	err := r.db.GetContext(ctx, &follower, `SELECT * FROM channel_followers WHERE channel_id = ? AND user_id = ?`, channelID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &follower, err
}
