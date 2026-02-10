package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type VoiceRepository struct {
	db *sqlx.DB
}

func NewVoiceRepository(db *sqlx.DB) *VoiceRepository {
	return &VoiceRepository{db: db}
}

func (r *VoiceRepository) Join(ctx context.Context, state *models.VoiceChannelState) error {
	query := `INSERT INTO voice_channel_states (id, channel_id, user_id, is_muted, is_deafened, is_screen_share, is_video_on, joined_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, state.ID, state.ChannelID, state.UserID, state.IsMuted, state.IsDeafened, state.IsScreenShare, state.IsVideoOn, state.JoinedAt)
	return err
}

func (r *VoiceRepository) Leave(ctx context.Context, channelID, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE voice_channel_states SET disconnected_at = ? WHERE channel_id = ? AND user_id = ? AND disconnected_at IS NULL`,
		time.Now(), channelID, userID)
	return err
}

func (r *VoiceRepository) GetState(ctx context.Context, channelID, userID string) (*models.VoiceChannelState, error) {
	var state models.VoiceChannelState
	err := r.db.GetContext(ctx, &state, `SELECT * FROM voice_channel_states WHERE channel_id = ? AND user_id = ? AND disconnected_at IS NULL`, channelID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &state, err
}

func (r *VoiceRepository) UpdateState(ctx context.Context, channelID, userID string, req *models.UpdateVoiceStateRequest) error {
	query := `UPDATE voice_channel_states SET is_muted = COALESCE(?, is_muted), is_deafened = COALESCE(?, is_deafened), is_screen_share = COALESCE(?, is_screen_share), is_video_on = COALESCE(?, is_video_on) WHERE channel_id = ? AND user_id = ? AND disconnected_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, req.IsMuted, req.IsDeafened, req.IsScreenShare, req.IsVideoOn, channelID, userID)
	return err
}

func (r *VoiceRepository) ListParticipants(ctx context.Context, channelID string) ([]*models.VoiceChannelState, error) {
	var states []*models.VoiceChannelState
	err := r.db.SelectContext(ctx, &states, `SELECT * FROM voice_channel_states WHERE channel_id = ? AND disconnected_at IS NULL ORDER BY joined_at`, channelID)
	return states, err
}

func (r *VoiceRepository) CountParticipants(ctx context.Context, channelID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM voice_channel_states WHERE channel_id = ? AND disconnected_at IS NULL`, channelID)
	return count, err
}
