package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type ThreadRepository struct {
	db *sqlx.DB
}

func NewThreadRepository(db *sqlx.DB) *ThreadRepository {
	return &ThreadRepository{db: db}
}

func (r *ThreadRepository) Create(ctx context.Context, thread *models.ChannelThread) error {
	query := `INSERT INTO channel_threads (id, channel_id, message_id, title, created_by, is_locked, is_resolved, reply_count, last_reply_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, thread.ID, thread.ChannelID, thread.MessageID, thread.Title, thread.CreatedBy, thread.IsLocked, thread.IsResolved, thread.ReplyCount, thread.LastReplyAt, thread.CreatedAt, thread.UpdatedAt)
	return err
}

func (r *ThreadRepository) GetByID(ctx context.Context, id string) (*models.ChannelThread, error) {
	var thread models.ChannelThread
	err := r.db.GetContext(ctx, &thread, `SELECT * FROM channel_threads WHERE id = ?`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &thread, err
}

func (r *ThreadRepository) ListByChannel(ctx context.Context, channelID string, limit, offset int) ([]*models.ChannelThread, error) {
	var threads []*models.ChannelThread
	err := r.db.SelectContext(ctx, &threads, `SELECT * FROM channel_threads WHERE channel_id = ? ORDER BY last_reply_at DESC LIMIT ? OFFSET ?`, channelID, limit, offset)
	return threads, err
}

func (r *ThreadRepository) Update(ctx context.Context, thread *models.ChannelThread) error {
	_, err := r.db.ExecContext(ctx, `UPDATE channel_threads SET title = ?, is_locked = ?, is_resolved = ?, updated_at = ? WHERE id = ?`,
		thread.Title, thread.IsLocked, thread.IsResolved, time.Now(), thread.ID)
	return err
}

func (r *ThreadRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM channel_threads WHERE id = ?`, id)
	return err
}

func (r *ThreadRepository) IncrementReplyCount(ctx context.Context, threadID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE channel_threads SET reply_count = reply_count + 1, last_reply_at = ?, updated_at = ? WHERE id = ?`,
		time.Now(), time.Now(), threadID)
	return err
}

func (r *ThreadRepository) DecrementReplyCount(ctx context.Context, threadID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE channel_threads SET reply_count = reply_count - 1, updated_at = ? WHERE id = ?`,
		time.Now(), threadID)
	return err
}

func (r *ThreadRepository) CreateReply(ctx context.Context, reply *models.ThreadReply) error {
	query := `INSERT INTO thread_replies (id, thread_id, user_id, content, parent_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, reply.ID, reply.ThreadID, reply.UserID, reply.Content, reply.ParentID, reply.CreatedAt, reply.UpdatedAt)
	return err
}

func (r *ThreadRepository) GetReplyByID(ctx context.Context, id string) (*models.ThreadReply, error) {
	var reply models.ThreadReply
	err := r.db.GetContext(ctx, &reply, `SELECT * FROM thread_replies WHERE id = ?`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &reply, err
}

func (r *ThreadRepository) ListReplies(ctx context.Context, threadID string, limit, offset int) ([]*models.ThreadReply, error) {
	var replies []*models.ThreadReply
	err := r.db.SelectContext(ctx, &replies, `SELECT * FROM thread_replies WHERE thread_id = ? ORDER BY created_at ASC LIMIT ? OFFSET ?`, threadID, limit, offset)
	return replies, err
}

func (r *ThreadRepository) UpdateReply(ctx context.Context, reply *models.ThreadReply) error {
	_, err := r.db.ExecContext(ctx, `UPDATE thread_replies SET content = ?, updated_at = ? WHERE id = ?`,
		reply.Content, time.Now(), reply.ID)
	return err
}

func (r *ThreadRepository) DeleteReply(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM thread_replies WHERE id = ?`, id)
	return err
}

func (r *ThreadRepository) AddFollower(ctx context.Context, follower *models.ThreadFollower) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO thread_followers (id, thread_id, user_id, created_at) VALUES (?, ?, ?, ?)`,
		follower.ID, follower.ThreadID, follower.UserID, follower.CreatedAt)
	return err
}

func (r *ThreadRepository) RemoveFollower(ctx context.Context, threadID, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM thread_followers WHERE thread_id = ? AND user_id = ?`, threadID, userID)
	return err
}

func (r *ThreadRepository) IsFollowing(ctx context.Context, threadID, userID string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM thread_followers WHERE thread_id = ? AND user_id = ?`, threadID, userID)
	return count > 0, err
}

func (r *ThreadRepository) ListFollowers(ctx context.Context, threadID string) ([]*models.ThreadFollower, error) {
	var followers []*models.ThreadFollower
	err := r.db.SelectContext(ctx, &followers, `SELECT * FROM thread_followers WHERE thread_id = ? ORDER BY created_at`, threadID)
	return followers, err
}

func (r *ThreadRepository) CountByChannel(ctx context.Context, channelID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM channel_threads WHERE channel_id = ?`, channelID)
	return count, err
}
