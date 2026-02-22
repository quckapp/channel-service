package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type BookmarkRepository struct {
	db *sqlx.DB
}

func NewBookmarkRepository(db *sqlx.DB) *BookmarkRepository {
	return &BookmarkRepository{db: db}
}

func (r *BookmarkRepository) Create(ctx context.Context, bookmark *models.ChannelBookmark) error {
	query := `INSERT INTO channel_bookmarks (id, channel_id, user_id, title, url, entity_type, entity_id, position, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		bookmark.ID, bookmark.ChannelID, bookmark.UserID, bookmark.Title,
		bookmark.URL, bookmark.EntityType, bookmark.EntityID, bookmark.Position,
		bookmark.CreatedAt, bookmark.UpdatedAt)
	return err
}

func (r *BookmarkRepository) GetByID(ctx context.Context, id string) (*models.ChannelBookmark, error) {
	var bookmark models.ChannelBookmark
	query := `SELECT * FROM channel_bookmarks WHERE id = ?`
	err := r.db.GetContext(ctx, &bookmark, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &bookmark, err
}

func (r *BookmarkRepository) ListByUserInChannel(ctx context.Context, channelID, userID string) ([]*models.ChannelBookmark, error) {
	var bookmarks []*models.ChannelBookmark
	query := `SELECT * FROM channel_bookmarks WHERE channel_id = ? AND user_id = ? ORDER BY position ASC`
	err := r.db.SelectContext(ctx, &bookmarks, query, channelID, userID)
	return bookmarks, err
}

func (r *BookmarkRepository) Update(ctx context.Context, bookmark *models.ChannelBookmark) error {
	query := `UPDATE channel_bookmarks SET title = ?, url = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, bookmark.Title, bookmark.URL, bookmark.ID)
	return err
}

func (r *BookmarkRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM channel_bookmarks WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *BookmarkRepository) GetMaxPosition(ctx context.Context, channelID, userID string) (int, error) {
	var pos sql.NullInt64
	query := `SELECT MAX(position) FROM channel_bookmarks WHERE channel_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &pos, query, channelID, userID)
	if err != nil || !pos.Valid {
		return 0, err
	}
	return int(pos.Int64), nil
}

func (r *BookmarkRepository) CountByUser(ctx context.Context, channelID, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM channel_bookmarks WHERE channel_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &count, query, channelID, userID)
	return count, err
}
