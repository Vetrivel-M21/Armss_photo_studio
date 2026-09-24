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

type Template struct {
	ID           string          `db:"id"`
	Title        string          `db:"title"`
	ObjectKey    sql.NullString  `db:"object_key"`
	ThumbKey     sql.NullString  `db:"thumb_key"`
	OccasionTags json.RawMessage `db:"occasion_tags"` // JSON array of strings
	Season       sql.NullString  `db:"season"`
	StyleProfile json.RawMessage `db:"style_profile"` // nullable JSON object
	IsActive     bool            `db:"is_active"`
	CreatedBy    sql.NullString  `db:"created_by"`
	CreatedAt    time.Time       `db:"created_at"`
}

type TemplatesRepo struct {
	db *sqlx.DB
}

func NewTemplatesRepo(db *sqlx.DB) *TemplatesRepo {
	return &TemplatesRepo{db: db}
}

func (r *TemplatesRepo) Create(ctx context.Context, title, objectKey, thumbKey string, occasionTags []string, season string, styleProfile json.RawMessage, createdBy string) (*Template, error) {
	id := uuid.NewString()
	tagsJSON, err := json.Marshal(occasionTags)
	if err != nil {
		return nil, fmt.Errorf("marshal occasion tags: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO templates (id, title, object_key, thumb_key, occasion_tags, season, style_profile, is_active, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?, TRUE, ?)`,
		id, title, nullIfEmpty(objectKey), nullIfEmpty(thumbKey), tagsJSON, nullIfEmpty(season), nullIfEmptyJSON(styleProfile), nullIfEmpty(createdBy))
	if err != nil {
		return nil, fmt.Errorf("insert template: %w", err)
	}
	return r.GetByID(ctx, id)
}

// ExistsByTitle reports whether an active or inactive template with this
// exact title already exists — used to make bulk library imports idempotent
// on repeat clicks.
func (r *TemplatesRepo) ExistsByTitle(ctx context.Context, title string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM templates WHERE title = ?`, title)
	if err != nil {
		return false, fmt.Errorf("check template title: %w", err)
	}
	return count > 0, nil
}

func (r *TemplatesRepo) GetByID(ctx context.Context, id string) (*Template, error) {
	var t Template
	err := r.db.GetContext(ctx, &t,
		`SELECT id, title, object_key, thumb_key, occasion_tags, season, style_profile, is_active, created_by, created_at
		 FROM templates WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}
	return &t, nil
}

// ListActive returns all active templates, optionally filtered by occasion
// substring match on the query string. Used both for the gallery and as the
// candidate set for recommendation ranking.
func (r *TemplatesRepo) ListActive(ctx context.Context, occasion string) ([]Template, error) {
	var templates []Template
	if occasion != "" {
		err := r.db.SelectContext(ctx, &templates,
			`SELECT id, title, object_key, thumb_key, occasion_tags, season, style_profile, is_active, created_by, created_at
			 FROM templates WHERE is_active = TRUE AND JSON_SEARCH(occasion_tags, 'one', ?) IS NOT NULL
			 ORDER BY created_at DESC`, "%"+occasion+"%")
		if err != nil {
			return nil, fmt.Errorf("list active templates by occasion: %w", err)
		}
		return templates, nil
	}
	err := r.db.SelectContext(ctx, &templates,
		`SELECT id, title, object_key, thumb_key, occasion_tags, season, style_profile, is_active, created_by, created_at
		 FROM templates WHERE is_active = TRUE ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list active templates: %w", err)
	}
	return templates, nil
}

// ListAll returns every template regardless of is_active — admins need to see
// and reactivate deactivated templates, which ListActive (used by the public
// gallery) deliberately excludes.
func (r *TemplatesRepo) ListAll(ctx context.Context, occasion string) ([]Template, error) {
	var templates []Template
	if occasion != "" {
		err := r.db.SelectContext(ctx, &templates,
			`SELECT id, title, object_key, thumb_key, occasion_tags, season, style_profile, is_active, created_by, created_at
			 FROM templates WHERE JSON_SEARCH(occasion_tags, 'one', ?) IS NOT NULL
			 ORDER BY created_at DESC`, "%"+occasion+"%")
		if err != nil {
			return nil, fmt.Errorf("list all templates by occasion: %w", err)
		}
		return templates, nil
	}
	err := r.db.SelectContext(ctx, &templates,
		`SELECT id, title, object_key, thumb_key, occasion_tags, season, style_profile, is_active, created_by, created_at
		 FROM templates ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all templates: %w", err)
	}
	return templates, nil
}

// UpdateStyleProfile lets an admin's correction replace the model's guess (spec §8).
func (r *TemplatesRepo) UpdateStyleProfile(ctx context.Context, id string, styleProfile json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `UPDATE templates SET style_profile = ? WHERE id = ?`, styleProfile, id)
	if err != nil {
		return fmt.Errorf("update template style profile: %w", err)
	}
	return nil
}

func (r *TemplatesRepo) SetActive(ctx context.Context, id string, active bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE templates SET is_active = ? WHERE id = ?`, active, id)
	if err != nil {
		return fmt.Errorf("update template active flag: %w", err)
	}
	return nil
}

func (r *TemplatesRepo) UpdateTags(ctx context.Context, id string, occasionTags []string, season string) error {
	tagsJSON, err := json.Marshal(occasionTags)
	if err != nil {
		return fmt.Errorf("marshal occasion tags: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `UPDATE templates SET occasion_tags = ?, season = ? WHERE id = ?`, tagsJSON, nullIfEmpty(season), id)
	if err != nil {
		return fmt.Errorf("update template tags: %w", err)
	}
	return nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullIfEmptyJSON(b json.RawMessage) any {
	if len(b) == 0 {
		return nil
	}
	return []byte(b)
}
