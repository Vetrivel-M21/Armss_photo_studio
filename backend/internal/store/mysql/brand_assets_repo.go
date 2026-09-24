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

const (
	BrandAssetLogo    = "logo"
	BrandAssetTagline = "tagline"
	BrandAssetFooter  = "footer"
)

type BrandAsset struct {
	ID        string          `db:"id"`
	Kind      string          `db:"kind"`
	ObjectKey string          `db:"object_key"`
	Meta      json.RawMessage `db:"meta"`
	Version   int             `db:"version"`
	IsActive  bool            `db:"is_active"`
	CreatedAt time.Time       `db:"created_at"`
}

type BrandAssetsRepo struct {
	db *sqlx.DB
}

func NewBrandAssetsRepo(db *sqlx.DB) *BrandAssetsRepo {
	return &BrandAssetsRepo{db: db}
}

// Replace deactivates the current active asset of this kind and inserts a new
// active one at version+1 — brand assets are versioned per spec §10, and PUT
// /brand-assets always replaces rather than mutates in place.
func (r *BrandAssetsRepo) Replace(ctx context.Context, kind, objectKey string, meta json.RawMessage) (*BrandAsset, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var currentVersion int
	err = tx.GetContext(ctx, &currentVersion,
		`SELECT COALESCE(MAX(version), 0) FROM brand_assets WHERE kind = ?`, kind)
	if err != nil {
		return nil, fmt.Errorf("get current version: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `UPDATE brand_assets SET is_active = FALSE WHERE kind = ? AND is_active = TRUE`, kind); err != nil {
		return nil, fmt.Errorf("deactivate previous asset: %w", err)
	}

	id := uuid.NewString()
	metaJSON := meta
	if len(metaJSON) == 0 {
		metaJSON = json.RawMessage(`{}`)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO brand_assets (id, kind, object_key, meta, version, is_active) VALUES (?, ?, ?, ?, ?, TRUE)`,
		id, kind, objectKey, []byte(metaJSON), currentVersion+1); err != nil {
		return nil, fmt.Errorf("insert brand asset: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *BrandAssetsRepo) GetByID(ctx context.Context, id string) (*BrandAsset, error) {
	var a BrandAsset
	err := r.db.GetContext(ctx, &a,
		`SELECT id, kind, object_key, meta, version, is_active, created_at FROM brand_assets WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get brand asset: %w", err)
	}
	return &a, nil
}

// ListActive returns the one active asset per kind (logo, tagline, footer).
func (r *BrandAssetsRepo) ListActive(ctx context.Context) ([]BrandAsset, error) {
	var assets []BrandAsset
	err := r.db.SelectContext(ctx, &assets,
		`SELECT id, kind, object_key, meta, version, is_active, created_at FROM brand_assets WHERE is_active = TRUE`)
	if err != nil {
		return nil, fmt.Errorf("list active brand assets: %w", err)
	}
	return assets, nil
}
