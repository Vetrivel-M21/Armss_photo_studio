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

type UserLimit struct {
	ID                  string         `db:"id"`
	UserID              sql.NullString `db:"user_id"`
	MaxAPICallsPerHour  int            `db:"max_api_calls_per_hour"`
	MaxRendersPerHour   int            `db:"max_renders_per_hour"`
	MonthlySpendCapCents int           `db:"monthly_spend_cap_cents"`
	IsUnlimited         bool           `db:"is_unlimited"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at"`
}

type UserLimitsRepo struct {
	db *sqlx.DB
}

func NewUserLimitsRepo(db *sqlx.DB) *UserLimitsRepo {
	return &UserLimitsRepo{db: db}
}

// GetGlobalDefault returns the system-wide fallback limit (user_id IS NULL),
// or a hardcoded default (30 calls/hr, 6 renders/hr, $15.00/mo) if not in DB.
func (r *UserLimitsRepo) GetGlobalDefault(ctx context.Context) (UserLimit, error) {
	var l UserLimit
	err := r.db.GetContext(ctx, &l,
		`SELECT id, user_id, max_api_calls_per_hour, max_renders_per_hour, monthly_spend_cap_cents, is_unlimited, created_at, updated_at
		 FROM user_limits WHERE user_id IS NULL LIMIT 1`)
	if errors.Is(err, sql.ErrNoRows) {
		return UserLimit{
			ID:                  "global-default",
			MaxAPICallsPerHour:  30,
			MaxRendersPerHour:   6,
			MonthlySpendCapCents: 1500,
			IsUnlimited:         false,
		}, nil
	}
	if err != nil {
		return UserLimit{}, fmt.Errorf("get global limit: %w", err)
	}
	return l, nil
}

// SetGlobalDefault inserts or updates the global default row (user_id IS NULL).
func (r *UserLimitsRepo) SetGlobalDefault(ctx context.Context, maxCalls, maxRenders, monthlyCap int) error {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_limits (id, user_id, max_api_calls_per_hour, max_renders_per_hour, monthly_spend_cap_cents, is_unlimited)
		 VALUES (?, NULL, ?, ?, ?, FALSE)
		 ON DUPLICATE KEY UPDATE
		   max_api_calls_per_hour = VALUES(max_api_calls_per_hour),
		   max_renders_per_hour = VALUES(max_renders_per_hour),
		   monthly_spend_cap_cents = VALUES(monthly_spend_cap_cents)`,
		id, maxCalls, maxRenders, monthlyCap)
	if err != nil {
		return fmt.Errorf("set global limits: %w", err)
	}
	return nil
}

// GetEffectiveLimit returns the user-specific limit if set, otherwise the global default.
func (r *UserLimitsRepo) GetEffectiveLimit(ctx context.Context, userID string) (UserLimit, error) {
	var l UserLimit
	err := r.db.GetContext(ctx, &l,
		`SELECT id, user_id, max_api_calls_per_hour, max_renders_per_hour, monthly_spend_cap_cents, is_unlimited, created_at, updated_at
		 FROM user_limits WHERE user_id = ?`, userID)
	if err == nil {
		return l, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return UserLimit{}, fmt.Errorf("get user limit for %s: %w", userID, err)
	}
	return r.GetGlobalDefault(ctx)
}

// SetForUser sets or updates limits for a specific user.
func (r *UserLimitsRepo) SetForUser(ctx context.Context, userID string, maxCalls, maxRenders, monthlyCap int, isUnlimited bool) error {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_limits (id, user_id, max_api_calls_per_hour, max_renders_per_hour, monthly_spend_cap_cents, is_unlimited)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   max_api_calls_per_hour = VALUES(max_api_calls_per_hour),
		   max_renders_per_hour = VALUES(max_renders_per_hour),
		   monthly_spend_cap_cents = VALUES(monthly_spend_cap_cents),
		   is_unlimited = VALUES(is_unlimited)`,
		id, userID, maxCalls, maxRenders, monthlyCap, isUnlimited)
	if err != nil {
		return fmt.Errorf("set limits for user %s: %w", userID, err)
	}
	return nil
}

// ListAll returns all user limits currently configured in DB.
func (r *UserLimitsRepo) ListAll(ctx context.Context) ([]UserLimit, error) {
	var limits []UserLimit
	err := r.db.SelectContext(ctx, &limits,
		`SELECT id, user_id, max_api_calls_per_hour, max_renders_per_hour, monthly_spend_cap_cents, is_unlimited, created_at, updated_at
		 FROM user_limits ORDER BY user_id IS NULL DESC, created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all user limits: %w", err)
	}
	return limits, nil
}
