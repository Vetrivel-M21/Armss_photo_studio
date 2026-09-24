package pipeline

import (
	"context"
	"fmt"

	"armss-poster-studio/internal/store/mysql"
)

// InsertSuperseding is the one helper every mutating pipeline handler goes
// through: it marks the current non-superseded row for (projectID, stage)
// superseded, then inserts the new row, atomically. This is what makes stage
// re-run/cascade auditable and centralized instead of scattered per-handler
// (spec §10's "re-running a stage inserts a new row and marks the previous
// one superseded, so the history is auditable and nothing is destroyed").
func InsertSuperseding(ctx context.Context, repo *mysql.StageOutputsRepo, projectID, stage string, payload any, model string, tokensIn, tokensOut int) (*mysql.StageOutput, error) {
	tx, err := repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin supersede tx: %w", err)
	}
	defer tx.Rollback()

	if err := repo.Supersede(ctx, tx, projectID, stage); err != nil {
		return nil, err
	}
	out, err := repo.Insert(ctx, tx, projectID, stage, payload, model, tokensIn, tokensOut)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit supersede tx: %w", err)
	}
	return out, nil
}

// SupersedeCascade marks each given stage's current row superseded without
// inserting a replacement — used when re-running an upstream stage (or
// editing copy) invalidates downstream stages that haven't been regenerated yet.
func SupersedeCascade(ctx context.Context, repo *mysql.StageOutputsRepo, projectID string, stages ...string) error {
	tx, err := repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin cascade tx: %w", err)
	}
	defer tx.Rollback()

	for _, stage := range stages {
		if err := repo.Supersede(ctx, tx, projectID, stage); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit cascade tx: %w", err)
	}
	return nil
}
