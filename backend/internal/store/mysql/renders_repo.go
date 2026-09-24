package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	RenderKindRaw        = "raw"
	RenderKindComposited = "composited"
)

type Render struct {
	ID        string          `db:"id"`
	ProjectID string          `db:"project_id"`
	Kind      string          `db:"kind"`
	Provider  sql.NullString  `db:"provider"`
	ObjectKey string          `db:"object_key"`
	Prompt    sql.NullString  `db:"prompt"`
	CostCents sql.NullInt32   `db:"cost_cents"`
	Meta      json.RawMessage `db:"meta"`
	CreatedAt time.Time       `db:"created_at"`
}

type RendersRepo struct {
	db *sqlx.DB
}

func NewRendersRepo(db *sqlx.DB) *RendersRepo {
	return &RendersRepo{db: db}
}

func (r *RendersRepo) Create(ctx context.Context, projectID, kind, provider, objectKey, prompt string, costCents int, meta any) (*Render, error) {
	id := uuid.NewString()
	var metaJSON []byte
	if meta != nil {
		var err error
		metaJSON, err = json.Marshal(meta)
		if err != nil {
			return nil, fmt.Errorf("marshal render meta: %w", err)
		}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO renders (id, project_id, kind, provider, object_key, prompt, cost_cents, meta)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, projectID, kind, provider, objectKey, prompt, costCents, metaJSON)
	if err != nil {
		return nil, fmt.Errorf("insert render: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *RendersRepo) GetByID(ctx context.Context, id string) (*Render, error) {
	var render Render
	err := r.db.GetContext(ctx, &render,
		`SELECT id, project_id, kind, provider, object_key, prompt, cost_cents, meta, created_at
		 FROM renders WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("get render: %w", err)
	}
	return &render, nil
}

// ListByProject returns every render for a project, newest first, for the
// raw/composited toggle and staleness comparison against the current art stage.
func (r *RendersRepo) ListByProject(ctx context.Context, projectID string) ([]Render, error) {
	var renders []Render
	err := r.db.SelectContext(ctx, &renders,
		`SELECT id, project_id, kind, provider, object_key, prompt, cost_cents, meta, created_at
		 FROM renders WHERE project_id = ? ORDER BY created_at DESC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list renders: %w", err)
	}
	return renders, nil
}
