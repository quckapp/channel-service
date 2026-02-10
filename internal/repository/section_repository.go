package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type SectionRepository struct {
	db *sqlx.DB
}

func NewSectionRepository(db *sqlx.DB) *SectionRepository {
	return &SectionRepository{db: db}
}

func (r *SectionRepository) Create(ctx context.Context, section *models.ChannelSection) error {
	query := `INSERT INTO channel_sections (id, workspace_id, user_id, name, position, is_collapsed, channel_ids, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		section.ID, section.WorkspaceID, section.UserID, section.Name,
		section.Position, section.IsCollapsed, section.ChannelIDs,
		section.CreatedAt, section.UpdatedAt)
	return err
}

func (r *SectionRepository) GetByID(ctx context.Context, id string) (*models.ChannelSection, error) {
	var section models.ChannelSection
	query := `SELECT * FROM channel_sections WHERE id = ?`
	err := r.db.GetContext(ctx, &section, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &section, err
}

func (r *SectionRepository) ListByUser(ctx context.Context, workspaceID, userID string) ([]*models.ChannelSection, error) {
	var sections []*models.ChannelSection
	query := `SELECT * FROM channel_sections WHERE workspace_id = ? AND user_id = ? ORDER BY position ASC`
	err := r.db.SelectContext(ctx, &sections, query, workspaceID, userID)
	return sections, err
}

func (r *SectionRepository) Update(ctx context.Context, section *models.ChannelSection) error {
	query := `UPDATE channel_sections SET name = ?, is_collapsed = ?, channel_ids = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, section.Name, section.IsCollapsed, section.ChannelIDs, section.ID)
	return err
}

func (r *SectionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM channel_sections WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *SectionRepository) GetMaxPosition(ctx context.Context, workspaceID, userID string) (int, error) {
	var pos sql.NullInt64
	query := `SELECT MAX(position) FROM channel_sections WHERE workspace_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &pos, query, workspaceID, userID)
	if err != nil || !pos.Valid {
		return 0, err
	}
	return int(pos.Int64), nil
}
