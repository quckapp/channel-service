package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type AnnouncementRepository struct {
	db *sqlx.DB
}

func NewAnnouncementRepository(db *sqlx.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

func (r *AnnouncementRepository) Create(ctx context.Context, ann *models.ChannelAnnouncement) error {
	query := `INSERT INTO channel_announcements (id, channel_id, title, content, priority, author_id, is_pinned, expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		ann.ID, ann.ChannelID, ann.Title, ann.Content, ann.Priority,
		ann.AuthorID, ann.IsPinned, ann.ExpiresAt, ann.CreatedAt, ann.UpdatedAt)
	return err
}

func (r *AnnouncementRepository) GetByID(ctx context.Context, id string) (*models.ChannelAnnouncement, error) {
	var ann models.ChannelAnnouncement
	query := `SELECT * FROM channel_announcements WHERE id = ?`
	err := r.db.GetContext(ctx, &ann, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &ann, err
}

func (r *AnnouncementRepository) ListByChannel(ctx context.Context, channelID string) ([]*models.ChannelAnnouncement, error) {
	var anns []*models.ChannelAnnouncement
	query := `SELECT * FROM channel_announcements WHERE channel_id = ? ORDER BY is_pinned DESC, created_at DESC`
	err := r.db.SelectContext(ctx, &anns, query, channelID)
	return anns, err
}

func (r *AnnouncementRepository) Update(ctx context.Context, ann *models.ChannelAnnouncement) error {
	query := `UPDATE channel_announcements SET title = ?, content = ?, priority = ?, expires_at = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, ann.Title, ann.Content, ann.Priority, ann.ExpiresAt, ann.ID)
	return err
}

func (r *AnnouncementRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM channel_announcements WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *AnnouncementRepository) TogglePin(ctx context.Context, id string, pinned bool) error {
	query := `UPDATE channel_announcements SET is_pinned = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, pinned, id)
	return err
}
