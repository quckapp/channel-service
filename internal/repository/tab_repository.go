package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type TabRepository struct {
	db *sqlx.DB
}

func NewTabRepository(db *sqlx.DB) *TabRepository {
	return &TabRepository{db: db}
}

func (r *TabRepository) Create(ctx context.Context, tab *models.ChannelTab) error {
	query := `INSERT INTO channel_tabs (id, channel_id, name, tab_type, config, position, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, tab.ID, tab.ChannelID, tab.Name, tab.TabType, tab.Config, tab.Position, tab.CreatedBy, tab.CreatedAt, tab.UpdatedAt)
	return err
}

func (r *TabRepository) GetByID(ctx context.Context, id string) (*models.ChannelTab, error) {
	var tab models.ChannelTab
	err := r.db.GetContext(ctx, &tab, "SELECT * FROM channel_tabs WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tab, err
}

func (r *TabRepository) ListByChannel(ctx context.Context, channelID string) ([]*models.ChannelTab, error) {
	var tabs []*models.ChannelTab
	err := r.db.SelectContext(ctx, &tabs, "SELECT * FROM channel_tabs WHERE channel_id = ? ORDER BY position ASC", channelID)
	return tabs, err
}

func (r *TabRepository) Update(ctx context.Context, tab *models.ChannelTab) error {
	query := `UPDATE channel_tabs SET name = ?, config = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, tab.Name, tab.Config, tab.ID)
	return err
}

func (r *TabRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM channel_tabs WHERE id = ?", id)
	return err
}

func (r *TabRepository) GetMaxPosition(ctx context.Context, channelID string) (int, error) {
	var pos sql.NullInt64
	err := r.db.GetContext(ctx, &pos, "SELECT MAX(position) FROM channel_tabs WHERE channel_id = ?", channelID)
	if err != nil || !pos.Valid {
		return 0, err
	}
	return int(pos.Int64), nil
}

func (r *TabRepository) UpdatePositions(ctx context.Context, channelID string, tabIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, tabID := range tabIDs {
		_, err := tx.ExecContext(ctx, "UPDATE channel_tabs SET position = ?, updated_at = NOW() WHERE id = ? AND channel_id = ?", i, tabID, channelID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
