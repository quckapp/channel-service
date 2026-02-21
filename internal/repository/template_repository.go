package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/channel-service/internal/models"
)

type TemplateRepository struct {
	db *sqlx.DB
}

func NewTemplateRepository(db *sqlx.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) Create(ctx context.Context, t *models.ChannelTemplate) error {
	query := `INSERT INTO channel_templates (id, name, description, created_by, channel_type, default_topic, default_tabs, default_settings, is_public, use_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.Name, t.Description, t.CreatedBy, t.ChannelType, t.DefaultTopic, t.DefaultTabs, t.DefaultSettings, t.IsPublic, t.UseCount, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *TemplateRepository) GetByID(ctx context.Context, id string) (*models.ChannelTemplate, error) {
	var t models.ChannelTemplate
	err := r.db.GetContext(ctx, &t, "SELECT * FROM channel_templates WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *TemplateRepository) List(ctx context.Context) ([]*models.ChannelTemplate, error) {
	var templates []*models.ChannelTemplate
	err := r.db.SelectContext(ctx, &templates, "SELECT * FROM channel_templates WHERE is_public = TRUE ORDER BY use_count DESC, created_at DESC")
	return templates, err
}

func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM channel_templates WHERE id = ?", id)
	return err
}

func (r *TemplateRepository) IncrementUseCount(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE channel_templates SET use_count = use_count + 1, updated_at = ? WHERE id = ?", time.Now(), id)
	return err
}
