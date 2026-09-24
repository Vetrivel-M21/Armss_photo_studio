package mysql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type ProjectReference struct {
	ProjectID    string          `db:"project_id"`
	ObjectKey    string          `db:"object_key"`
	StyleProfile json.RawMessage `db:"style_profile"`
}

type ProjectReferencesRepo struct {
	db *sqlx.DB
}

func NewProjectReferencesRepo(db *sqlx.DB) *ProjectReferencesRepo {
	return &ProjectReferencesRepo{db: db}
}

func (r *ProjectReferencesRepo) Create(ctx context.Context, projectID, objectKey string, styleProfile json.RawMessage) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO project_references (project_id, object_key, style_profile) VALUES (?, ?, ?)`,
		projectID, objectKey, styleProfile)
	if err != nil {
		return fmt.Errorf("insert project reference: %w", err)
	}
	return nil
}

func (r *ProjectReferencesRepo) ListByProject(ctx context.Context, projectID string) ([]ProjectReference, error) {
	var refs []ProjectReference
	err := r.db.SelectContext(ctx, &refs,
		`SELECT project_id, object_key, style_profile FROM project_references WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project references: %w", err)
	}
	return refs, nil
}
