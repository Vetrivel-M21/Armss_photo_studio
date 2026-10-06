package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	ActionPipeline = "pipeline"
	ActionRender   = "render"
)

type RateRecordsRepo struct {
	db *sqlx.DB
}

func NewRateRecordsRepo(db *sqlx.DB) *RateRecordsRepo {
	return &RateRecordsRepo{db: db}
}

// RecordAction logs a usage event for the user in the sliding window.
func (r *RateRecordsRepo) RecordAction(ctx context.Context, userID, actionType string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO api_rate_records (user_id, action_type, created_at) VALUES (?, ?, NOW(6))`,
		userID, actionType)
	if err != nil {
		return fmt.Errorf("record rate action: %w", err)
	}
	return nil
}

// CountInPastHour returns the count of actions recorded in the past 60 minutes.
func (r *RateRecordsRepo) CountInPastHour(ctx context.Context, userID, actionType string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM api_rate_records
		 WHERE user_id = ? AND action_type = ? AND created_at >= DATE_SUB(NOW(6), INTERVAL 1 HOUR)`,
		userID, actionType)
	if err != nil {
		return 0, fmt.Errorf("count rate records in past hour: %w", err)
	}
	return count, nil
}

// OldestInPastHour finds the earliest action within the past hour to calculate exact quota reset time.
func (r *RateRecordsRepo) OldestInPastHour(ctx context.Context, userID, actionType string) (time.Time, error) {
	var t time.Time
	err := r.db.GetContext(ctx, &t,
		`SELECT created_at FROM api_rate_records
		 WHERE user_id = ? AND action_type = ? AND created_at >= DATE_SUB(NOW(6), INTERVAL 1 HOUR)
		 ORDER BY created_at ASC LIMIT 1`,
		userID, actionType)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, ErrNotFound
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("get oldest rate record in past hour: %w", err)
	}
	return t, nil
}

// PurgeOld deletes rate records older than maxAge (default 24h) to keep the table compact.
func (r *RateRecordsRepo) PurgeOld(ctx context.Context, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)
	_, err := r.db.ExecContext(ctx, `DELETE FROM api_rate_records WHERE created_at < ?`, cutoff)
	if err != nil {
		return fmt.Errorf("purge old rate records: %w", err)
	}
	return nil
}

type UserUsageRow struct {
	UserID         string         `db:"user_id"`
	Email          string         `db:"email"`
	Role           string         `db:"role"`
	CreatedAt      time.Time      `db:"created_at"`
	TotalProjects  int            `db:"total_projects"`
	TotalTokensIn  sql.NullInt64  `db:"total_tokens_in"`
	TotalTokensOut sql.NullInt64  `db:"total_tokens_out"`
	TotalRenders   int            `db:"total_renders"`
	TotalCostCents sql.NullInt64  `db:"total_cost_cents"`
}

// GetUserUsageOverview compiles aggregate usage stats across all users by joining
// users with projects, stage_outputs, and renders.
func (r *RateRecordsRepo) GetUserUsageOverview(ctx context.Context) ([]UserUsageRow, error) {
	query := `
		SELECT
		  u.id AS user_id,
		  u.email,
		  u.role,
		  u.created_at,
		  (SELECT COUNT(*) FROM projects p WHERE p.user_id = u.id) AS total_projects,
		  (SELECT COALESCE(SUM(so.tokens_in), 0) FROM stage_outputs so JOIN projects p ON so.project_id = p.id WHERE p.user_id = u.id) AS total_tokens_in,
		  (SELECT COALESCE(SUM(so.tokens_out), 0) FROM stage_outputs so JOIN projects p ON so.project_id = p.id WHERE p.user_id = u.id) AS total_tokens_out,
		  (SELECT COUNT(*) FROM renders r JOIN projects p ON r.project_id = p.id WHERE p.user_id = u.id) AS total_renders,
		  (SELECT COALESCE(SUM(r.cost_cents), 0) FROM renders r JOIN projects p ON r.project_id = p.id WHERE p.user_id = u.id) AS total_cost_cents
		FROM users u
		ORDER BY total_renders DESC, total_projects DESC
	`
	var rows []UserUsageRow
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("get user usage overview: %w", err)
	}
	return rows, nil
}
