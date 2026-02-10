package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type ChannelRepository struct {
	db *sqlx.DB
}

func NewChannelRepository(db *sqlx.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) Create(ctx context.Context, ch *models.Channel) error {
	query := `INSERT INTO channels (id, workspace_id, name, type, description, topic, icon_url, is_archived, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, ch.ID, ch.WorkspaceID, ch.Name, ch.Type, ch.Description, ch.Topic, ch.IconURL, ch.IsArchived, ch.CreatedBy, ch.CreatedAt, ch.UpdatedAt)
	return err
}

func (r *ChannelRepository) GetByID(ctx context.Context, id string) (*models.Channel, error) {
	var ch models.Channel
	query := `SELECT * FROM channels WHERE id = ? AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &ch, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &ch, err
}

func (r *ChannelRepository) Update(ctx context.Context, ch *models.Channel) error {
	query := `UPDATE channels SET name = ?, description = ?, topic = ?, icon_url = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, ch.Name, ch.Description, ch.Topic, ch.IconURL, time.Now(), ch.ID)
	return err
}

func (r *ChannelRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE channels SET deleted_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ChannelRepository) ListByWorkspace(ctx context.Context, workspaceID string) ([]*models.Channel, error) {
	var channels []*models.Channel
	query := `SELECT * FROM channels WHERE workspace_id = ? AND deleted_at IS NULL ORDER BY name`
	err := r.db.SelectContext(ctx, &channels, query, workspaceID)
	return channels, err
}

func (r *ChannelRepository) ListByUser(ctx context.Context, userID, workspaceID string) ([]*models.Channel, error) {
	var channels []*models.Channel
	query := `SELECT c.* FROM channels c
		INNER JOIN channel_members cm ON c.id = cm.channel_id
		WHERE cm.user_id = ? AND c.workspace_id = ? AND c.deleted_at IS NULL
		ORDER BY c.name`
	err := r.db.SelectContext(ctx, &channels, query, userID, workspaceID)
	return channels, err
}

func (r *ChannelRepository) Archive(ctx context.Context, id string) error {
	query := `UPDATE channels SET is_archived = TRUE, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ChannelRepository) Unarchive(ctx context.Context, id string) error {
	query := `UPDATE channels SET is_archived = FALSE, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ChannelRepository) GetMemberCount(ctx context.Context, channelID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM channel_members WHERE channel_id = ?`
	err := r.db.GetContext(ctx, &count, query, channelID)
	return count, err
}

func (r *ChannelRepository) Search(ctx context.Context, workspaceID, query string) ([]*models.Channel, error) {
	var channels []*models.Channel
	searchQuery := `SELECT * FROM channels WHERE workspace_id = ? AND deleted_at IS NULL
		AND (name LIKE ? OR description LIKE ?) ORDER BY name LIMIT 50`
	pattern := "%" + query + "%"
	err := r.db.SelectContext(ctx, &channels, searchQuery, workspaceID, pattern, pattern)
	return channels, err
}
