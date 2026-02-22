package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type MemberRepository struct {
	db *sqlx.DB
}

func NewMemberRepository(db *sqlx.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) Create(ctx context.Context, m *models.ChannelMember) error {
	query := `INSERT INTO channel_members (id, channel_id, user_id, role, notifications, joined_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, m.ID, m.ChannelID, m.UserID, m.Role, m.Notifications, m.JoinedAt)
	return err
}

func (r *MemberRepository) GetByChannelAndUser(ctx context.Context, channelID, userID string) (*models.ChannelMember, error) {
	var m models.ChannelMember
	query := `SELECT * FROM channel_members WHERE channel_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &m, query, channelID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func (r *MemberRepository) ListByChannel(ctx context.Context, channelID string) ([]*models.ChannelMember, error) {
	var members []*models.ChannelMember
	query := `SELECT * FROM channel_members WHERE channel_id = ? ORDER BY joined_at`
	err := r.db.SelectContext(ctx, &members, query, channelID)
	return members, err
}

func (r *MemberRepository) Remove(ctx context.Context, channelID, userID string) error {
	query := `DELETE FROM channel_members WHERE channel_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, channelID, userID)
	return err
}

func (r *MemberRepository) UpdateRole(ctx context.Context, channelID, userID, role string) error {
	query := `UPDATE channel_members SET role = ? WHERE channel_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, role, channelID, userID)
	return err
}

func (r *MemberRepository) UpdateNotifications(ctx context.Context, channelID, userID, level string) error {
	query := `UPDATE channel_members SET notifications = ? WHERE channel_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, level, channelID, userID)
	return err
}

func (r *MemberRepository) UpdateLastRead(ctx context.Context, channelID, userID string) error {
	query := `UPDATE channel_members SET last_read_at = NOW() WHERE channel_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, channelID, userID)
	return err
}

func (r *MemberRepository) IsMember(ctx context.Context, channelID, userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM channel_members WHERE channel_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &count, query, channelID, userID)
	return count > 0, err
}

func (r *MemberRepository) GetRole(ctx context.Context, channelID, userID string) (string, error) {
	var role string
	query := `SELECT role FROM channel_members WHERE channel_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &role, query, channelID, userID)
	return role, err
}
