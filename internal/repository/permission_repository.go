package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type PermissionRepository struct {
	db *sqlx.DB
}

func NewPermissionRepository(db *sqlx.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) Set(ctx context.Context, perm *models.ChannelPermission) error {
	query := `INSERT INTO channel_permissions (id, channel_id, permission_type, target_type, target_id, allow, deny, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE allow = VALUES(allow), deny = VALUES(deny), updated_at = NOW()`
	_, err := r.db.ExecContext(ctx, query,
		perm.ID, perm.ChannelID, perm.PermissionType, perm.TargetType, perm.TargetID,
		perm.Allow, perm.Deny, perm.CreatedAt, perm.UpdatedAt)
	return err
}

func (r *PermissionRepository) GetByID(ctx context.Context, id string) (*models.ChannelPermission, error) {
	var perm models.ChannelPermission
	query := `SELECT * FROM channel_permissions WHERE id = ?`
	err := r.db.GetContext(ctx, &perm, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &perm, err
}

func (r *PermissionRepository) ListByChannel(ctx context.Context, channelID string) ([]*models.ChannelPermission, error) {
	var perms []*models.ChannelPermission
	query := `SELECT * FROM channel_permissions WHERE channel_id = ? ORDER BY permission_type, target_type`
	err := r.db.SelectContext(ctx, &perms, query, channelID)
	return perms, err
}

func (r *PermissionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM channel_permissions WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PermissionRepository) GetEffective(ctx context.Context, channelID, userID, role string) ([]*models.ChannelPermission, error) {
	var perms []*models.ChannelPermission
	query := `SELECT * FROM channel_permissions WHERE channel_id = ? AND (
		(target_type = 'user' AND target_id = ?) OR
		(target_type = 'role' AND target_id = ?)
	) ORDER BY target_type DESC`
	err := r.db.SelectContext(ctx, &perms, query, channelID, userID, role)
	return perms, err
}
