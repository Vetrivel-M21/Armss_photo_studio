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

// Upload backs POST /uploads — spec §9 lists the endpoint but §10's schema has no
// table for it; this is the minimal one needed to hand back a stable upload id.
type Upload struct {
	ID          string    `db:"id"`
	ObjectKey   string    `db:"object_key"`
	ContentType string    `db:"content_type"`
	SizeBytes   int       `db:"size_bytes"`
	CreatedBy   string    `db:"created_by"`
	CreatedAt   time.Time `db:"created_at"`
}

type UploadsRepo struct {
	db *sqlx.DB
}

func NewUploadsRepo(db *sqlx.DB) *UploadsRepo {
	return &UploadsRepo{db: db}
}

func (r *UploadsRepo) Create(ctx context.Context, objectKey, contentType string, sizeBytes int, createdBy string) (*Upload, error) {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO uploads (id, object_key, content_type, size_bytes, created_by) VALUES (?, ?, ?, ?, ?)`,
		id, objectKey, contentType, sizeBytes, createdBy)
	if err != nil {
		return nil, fmt.Errorf("insert upload: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *UploadsRepo) GetByID(ctx context.Context, id string) (*Upload, error) {
	var u Upload
	err := r.db.GetContext(ctx, &u,
		`SELECT id, object_key, content_type, size_bytes, created_by, created_at FROM uploads WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get upload: %w", err)
	}
	return &u, nil
}
