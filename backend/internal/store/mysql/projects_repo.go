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
	ID             string         `db:"id"`
	UserID         string         `db:"user_id"`
	Title          sql.NullString `db:"title"`
	InputText      string         `db:"input_text"`
	TemplateID     sql.NullString `db:"template_id"`
	LogoObjectKey  sql.NullString `db:"logo_object_key"`
	FooterText     sql.NullString `db:"footer_text"`
	// Zone hint fields — set at project creation, drive Art Director zone geometry.
	AspectRatio    string         `db:"aspect_ratio"`
	LogoPosition   string         `db:"logo_position"`
	LogoScale      string         `db:"logo_scale"`
	SloganPosition string         `db:"slogan_position"`
	SloganScale    string         `db:"slogan_scale"`
	FooterVisible  bool           `db:"footer_visible"`
	State          string         `db:"state"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

type ProjectsRepo struct {
	db *sqlx.DB
}

func NewProjectsRepo(db *sqlx.DB) *ProjectsRepo {
	return &ProjectsRepo{db: db}
}

// ZoneHints carries the user's logo/slogan placement choices from the start
// screen so the Art Director can encode them as hard geometry constraints
// rather than parsing them out of free text.
type ZoneHints struct {
	AspectRatio    string // e.g. "3:4"
	LogoPosition   string // e.g. "top-center"
	LogoScale      string // e.g. "standard"
	SloganPosition string // e.g. "below-logo"
	SloganScale    string // e.g. "standard"
	FooterVisible  bool
}

func (r *ProjectsRepo) Create(ctx context.Context, userID, inputText string, templateID, logoObjectKey, footerText *string, hints ZoneHints) (*Project, error) {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO projects
			(id, user_id, input_text, template_id, logo_object_key, footer_text,
			 aspect_ratio, logo_position, logo_scale, slogan_position, slogan_scale, footer_visible,
			 state)
		 VALUES (?, ?, ?, ?, ?, ?,  ?, ?, ?, ?, ?, ?,  ?)`,
		id, userID, inputText, templateID, logoObjectKey, footerText,
		hints.AspectRatio, hints.LogoPosition, hints.LogoScale,
		hints.SloganPosition, hints.SloganScale, hints.FooterVisible,
		StateDraft)
	if err != nil {
		return nil, fmt.Errorf("insert project: %w", err)
	}
	return r.GetByID(ctx, id)
}

const projectColumns = `id, user_id, title, input_text, template_id, logo_object_key, footer_text,
			aspect_ratio, logo_position, logo_scale, slogan_position, slogan_scale, footer_visible,
			state, created_at, updated_at`

func (r *ProjectsRepo) GetByID(ctx context.Context, id string) (*Project, error) {
	var p Project
	err := r.db.GetContext(ctx, &p,
		`SELECT `+projectColumns+` FROM projects WHERE id = ?`, id)
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
		`SELECT `+projectColumns+` FROM projects WHERE user_id = ? ORDER BY created_at DESC`, userID)
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
