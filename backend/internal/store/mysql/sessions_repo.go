package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Session struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type SessionsRepo struct {
	db *sqlx.DB
}

func NewSessionsRepo(db *sqlx.DB) *SessionsRepo {
	return &SessionsRepo{db: db}
}

func (r *SessionsRepo) Create(ctx context.Context, id, userID string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`, id, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *SessionsRepo) Get(ctx context.Context, id string) (*Session, error) {
	var s Session
	err := r.db.GetContext(ctx, &s,
		`SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = ? AND expires_at > UTC_TIMESTAMP(6)`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &s, nil
}

func (r *SessionsRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
