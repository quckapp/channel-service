package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type InviteRepository struct {
	db *sqlx.DB
}

func NewInviteRepository(db *sqlx.DB) *InviteRepository {
	return &InviteRepository{db: db}
}

func (r *InviteRepository) Create(ctx context.Context, invite *models.ChannelInvite) error {
	query := `INSERT INTO channel_invites (id, channel_id, created_by, code, max_uses, use_count, expires_at, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		invite.ID, invite.ChannelID, invite.CreatedBy, invite.Code,
		invite.MaxUses, invite.UseCount, invite.ExpiresAt, invite.IsActive, invite.CreatedAt)
	return err
}

func (r *InviteRepository) GetByCode(ctx context.Context, code string) (*models.ChannelInvite, error) {
	var invite models.ChannelInvite
	query := `SELECT * FROM channel_invites WHERE code = ? AND is_active = TRUE`
	err := r.db.GetContext(ctx, &invite, query, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &invite, err
}

func (r *InviteRepository) GetByID(ctx context.Context, id string) (*models.ChannelInvite, error) {
	var invite models.ChannelInvite
	query := `SELECT * FROM channel_invites WHERE id = ?`
	err := r.db.GetContext(ctx, &invite, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &invite, err
}

func (r *InviteRepository) ListByChannel(ctx context.Context, channelID string) ([]*models.ChannelInvite, error) {
	var invites []*models.ChannelInvite
	query := `SELECT * FROM channel_invites WHERE channel_id = ? ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &invites, query, channelID)
	return invites, err
}

func (r *InviteRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM channel_invites WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *InviteRepository) IncrementUseCount(ctx context.Context, id string) error {
	query := `UPDATE channel_invites SET use_count = use_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *InviteRepository) DeactivateExpired(ctx context.Context) error {
	query := `UPDATE channel_invites SET is_active = FALSE WHERE expires_at IS NOT NULL AND expires_at < NOW() AND is_active = TRUE`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
