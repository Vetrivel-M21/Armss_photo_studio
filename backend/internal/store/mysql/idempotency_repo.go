package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// IdempotencyRepo backs the Idempotency-Key header required on render and
// composite (spec §13) so a retried request replays the stored response
// instead of double-charging an image provider.
type IdempotencyRepo struct {
	db *sqlx.DB
}

func NewIdempotencyRepo(db *sqlx.DB) *IdempotencyRepo {
	return &IdempotencyRepo{db: db}
}

// Get returns the stored response for a key, if this exact key was already used.
func (r *IdempotencyRepo) Get(ctx context.Context, key string) (json.RawMessage, error) {
	var raw json.RawMessage
	err := r.db.GetContext(ctx, &raw, "SELECT response_snapshot FROM idempotency_keys WHERE `key` = ?", key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get idempotency key: %w", err)
	}
	return raw, nil
}

// Save records the response for a key, so a repeated request can replay it.
func (r *IdempotencyRepo) Save(ctx context.Context, key, projectID, endpoint string, response any) error {
	snapshot, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal idempotency snapshot: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		"INSERT INTO idempotency_keys (`key`, project_id, endpoint, response_snapshot) VALUES (?, ?, ?, ?)",
		key, projectID, endpoint, snapshot)
	if err != nil {
		return fmt.Errorf("save idempotency key: %w", err)
	}
	return nil
}
