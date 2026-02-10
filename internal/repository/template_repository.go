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

func (r *TemplateRepository) Create(ctx context.Context, tmpl *models.ChannelTemplate) error {
	query := `INSERT INTO channel_templates (id, workspace_id, name, description, type, topic, settings, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, tmpl.ID, tmpl.WorkspaceID, tmpl.Name, tmpl.Description, tmpl.Type, tmpl.Topic, tmpl.Settings, tmpl.CreatedBy, tmpl.CreatedAt, tmpl.UpdatedAt)
	return err
}

func (r *TemplateRepository) GetByID(ctx context.Context, id string) (*models.ChannelTemplate, error) {
	var tmpl models.ChannelTemplate
	err := r.db.GetContext(ctx, &tmpl, `SELECT * FROM channel_templates WHERE id = ?`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tmpl, err
}

func (r *TemplateRepository) ListByWorkspace(ctx context.Context, workspaceID string) ([]*models.ChannelTemplate, error) {
	var templates []*models.ChannelTemplate
	err := r.db.SelectContext(ctx, &templates, `SELECT * FROM channel_templates WHERE workspace_id = ? ORDER BY name`, workspaceID)
	return templates, err
}

func (r *TemplateRepository) Update(ctx context.Context, tmpl *models.ChannelTemplate) error {
	_, err := r.db.ExecContext(ctx, `UPDATE channel_templates SET name = ?, description = ?, type = ?, topic = ?, settings = ?, updated_at = ? WHERE id = ?`,
		tmpl.Name, tmpl.Description, tmpl.Type, tmpl.Topic, tmpl.Settings, time.Now(), tmpl.ID)
	return err
}

func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM channel_templates WHERE id = ?`, id)
	return err
}
