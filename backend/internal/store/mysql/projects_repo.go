package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	StateDraft         = "draft"
	StateBriefing      = "briefing"
	StateConceptsReady = "concepts_ready"
	StateDeveloping    = "developing"
	StateRenderReady   = "render_ready"
	StateComposited    = "composited"
	StateApproved      = "approved"
)

type Project struct {
	ID            string         `db:"id"`
	UserID        string         `db:"user_id"`
	Title         sql.NullString `db:"title"`
	InputText     string         `db:"input_text"`
	TemplateID    sql.NullString `db:"template_id"`
	LogoObjectKey sql.NullString `db:"logo_object_key"`
	FooterText    sql.NullString `db:"footer_text"`
	State         string         `db:"state"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}

type ProjectsRepo struct {
	db *sqlx.DB
}

func NewProjectsRepo(db *sqlx.DB) *ProjectsRepo {
	return &ProjectsRepo{db: db}
}

func (r *ProjectsRepo) Create(ctx context.Context, userID, inputText string, templateID, logoObjectKey, footerText *string) (*Project, error) {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO projects (id, user_id, input_text, template_id, logo_object_key, footer_text, state) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, userID, inputText, templateID, logoObjectKey, footerText, StateDraft)
	if err != nil {
		return nil, fmt.Errorf("insert project: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *ProjectsRepo) GetByID(ctx context.Context, id string) (*Project, error) {
	var p Project
	err := r.db.GetContext(ctx, &p,
		`SELECT id, user_id, title, input_text, template_id, logo_object_key, footer_text, state, created_at, updated_at
		 FROM projects WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	return &p, nil
}

func (r *ProjectsRepo) ListByUser(ctx context.Context, userID string) ([]Project, error) {
	var projects []Project
	err := r.db.SelectContext(ctx, &projects,
		`SELECT id, user_id, title, input_text, template_id, logo_object_key, footer_text, state, created_at, updated_at
		 FROM projects WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return projects, nil
}

func (r *ProjectsRepo) SetState(ctx context.Context, id, state string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE projects SET state = ? WHERE id = ?`, state, id)
	if err != nil {
		return fmt.Errorf("update project state: %w", err)
	}
	return nil
}
