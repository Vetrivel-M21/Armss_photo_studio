package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Stage names. "concept_selection" is an addition beyond the spec's literal
// enum (style|brief|concepts|copy|art|proof|review) — it keeps which concept was
// picked in the same append-only, superseded-on-rerun history as everything else,
// instead of adding a separate column to projects.
const (
	StageStyle            = "style"
	StageBrief            = "brief"
	StageConcepts         = "concepts"
	StageConceptSelection = "concept_selection"
	StageCopy             = "copy"
	StageArt              = "art"
	StageProof            = "proof"
	StageReview           = "review"
)

type StageOutput struct {
	ID         string          `db:"id"`
	ProjectID  string          `db:"project_id"`
	Stage      string          `db:"stage"`
	Payload    json.RawMessage `db:"payload"`
	Model      sql.NullString  `db:"model"`
	TokensIn   sql.NullInt32   `db:"tokens_in"`
	TokensOut  sql.NullInt32   `db:"tokens_out"`
	Superseded bool            `db:"superseded"`
	CreatedAt  time.Time       `db:"created_at"`
}

type StageOutputsRepo struct {
	db *sqlx.DB
}

func NewStageOutputsRepo(db *sqlx.DB) *StageOutputsRepo {
	return &StageOutputsRepo{db: db}
}

func (r *StageOutputsRepo) DB() *sqlx.DB { return r.db }

// Supersede marks the current non-superseded row for (projectID, stage) as
// superseded, if one exists. Callers run this inside the same transaction as
// the following Insert — see pipeline.InsertSuperseding.
func (r *StageOutputsRepo) Supersede(ctx context.Context, tx *sqlx.Tx, projectID, stage string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE stage_outputs SET superseded = TRUE WHERE project_id = ? AND stage = ? AND superseded = FALSE`,
		projectID, stage)
	if err != nil {
		return fmt.Errorf("supersede stage output: %w", err)
	}
	return nil
}

// Insert writes a new stage_outputs row. model is empty for a manual/human edit,
// distinguishing agent-authored rows from user-authored ones.
func (r *StageOutputsRepo) Insert(ctx context.Context, tx *sqlx.Tx, projectID, stage string, payload any, model string, tokensIn, tokensOut int) (*StageOutput, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal stage payload: %w", err)
	}
	id := uuid.NewString()

	var modelVal any
	var tokensInVal, tokensOutVal any
	if model != "" {
		modelVal = model
	}
	if tokensIn > 0 || tokensOut > 0 {
		tokensInVal, tokensOutVal = tokensIn, tokensOut
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO stage_outputs (id, project_id, stage, payload, model, tokens_in, tokens_out, superseded)
		 VALUES (?, ?, ?, ?, ?, ?, ?, FALSE)`,
		id, projectID, stage, payloadJSON, modelVal, tokensInVal, tokensOutVal)
	if err != nil {
		return nil, fmt.Errorf("insert stage output: %w", err)
	}

	var out StageOutput
	if err := tx.GetContext(ctx, &out,
		`SELECT id, project_id, stage, payload, model, tokens_in, tokens_out, superseded, created_at
		 FROM stage_outputs WHERE id = ?`, id); err != nil {
		return nil, fmt.Errorf("read back stage output: %w", err)
	}
	return &out, nil
}

// GetLatest returns the current non-superseded row for (projectID, stage).
func (r *StageOutputsRepo) GetLatest(ctx context.Context, projectID, stage string) (*StageOutput, error) {
	var out StageOutput
	err := r.db.GetContext(ctx, &out,
		`SELECT id, project_id, stage, payload, model, tokens_in, tokens_out, superseded, created_at
		 FROM stage_outputs WHERE project_id = ? AND stage = ? AND superseded = FALSE
		 ORDER BY created_at DESC LIMIT 1`, projectID, stage)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get latest stage output: %w", err)
	}
	return &out, nil
}

// ListByProject returns the full history (including superseded rows) for a
// project, oldest first — the auditable re-run trail (spec §10).
func (r *StageOutputsRepo) ListByProject(ctx context.Context, projectID string) ([]StageOutput, error) {
	var outs []StageOutput
	err := r.db.SelectContext(ctx, &outs,
		`SELECT id, project_id, stage, payload, model, tokens_in, tokens_out, superseded, created_at
		 FROM stage_outputs WHERE project_id = ? ORDER BY created_at ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list stage outputs: %w", err)
	}
	return outs, nil
}

// BeginTx starts a transaction for the supersede-then-insert pattern.
func (r *StageOutputsRepo) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}
