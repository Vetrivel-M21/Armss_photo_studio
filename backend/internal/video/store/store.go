package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"armss-poster-studio/internal/video/render"
)

var (
	ErrNotFound       = errors.New("video entity not found")
	ErrConflict       = errors.New("optimistic lock revision conflict")
	ErrUnauthorized   = errors.New("unauthorized access to video project")
)

// ProjectRow represents a row in video_projects table
type ProjectRow struct {
	ID          string    `db:"id" json:"id"`
	UserID      string    `db:"user_id" json:"user_id"`
	Title       string    `db:"title" json:"title"`
	SpecJSON    string    `db:"spec" json:"-"`
	SpecVersion int       `db:"spec_version" json:"spec_version"`
	Revision    int       `db:"revision" json:"revision"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	Spec        *render.VideoSpec `db:"-" json:"spec,omitempty"`
}

// JobRow represents a row in video_jobs table
type JobRow struct {
	ID          string         `db:"id" json:"id"`
	ProjectID   sql.NullString `db:"project_id" json:"project_id,omitempty"`
	UserID      string         `db:"user_id" json:"user_id"`
	SpecJSON    string         `db:"spec" json:"-"`
	Resolution  string         `db:"resolution" json:"resolution"`
	Status      string         `db:"status" json:"status"`
	Progress    int            `db:"progress" json:"progress"`
	OutputKey   sql.NullString `db:"output_key" json:"output_key,omitempty"`
	Error       sql.NullString `db:"error" json:"error,omitempty"`
	StartedAt   *time.Time     `db:"started_at" json:"started_at,omitempty"`
	FinishedAt  *time.Time     `db:"finished_at" json:"finished_at,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
	Spec        *render.VideoSpec `db:"-" json:"spec,omitempty"`
}

// AudioPresetRow represents a row in video_audio_presets table
type AudioPresetRow struct {
	ID          string         `db:"id" json:"id"`
	Title       string         `db:"title" json:"title"`
	Artist      sql.NullString `db:"artist" json:"artist,omitempty"`
	S3Key       string         `db:"s3_key" json:"s3_key"`
	Duration    float64        `db:"duration" json:"duration"`
	License     string         `db:"license" json:"license"`
	SourceURL   string         `db:"source_url" json:"source_url"`
	Attribution sql.NullString `db:"attribution" json:"attribution,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
}

type Store struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Store {
	return &Store{db: db}
}

// ---- Projects ----

func (s *Store) CreateProject(ctx context.Context, userID, title string, spec *render.VideoSpec) (*ProjectRow, error) {
	if spec == nil {
		d := render.DefaultSpec()
		spec = &d
	}
	specRaw, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal spec: %w", err)
	}

	id := uuid.NewString()
	now := time.Now().UTC()

	query := `
		INSERT INTO video_projects (id, user_id, title, spec, spec_version, revision, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 1, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query, id, userID, title, string(specRaw), render.SpecVersion, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert project: %w", err)
	}

	return &ProjectRow{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Spec:        spec,
		SpecVersion: render.SpecVersion,
		Revision:    1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Store) GetProject(ctx context.Context, id, userID string) (*ProjectRow, error) {
	var row ProjectRow
	query := `
		SELECT id, user_id, title, spec, spec_version, revision, created_at, updated_at
		FROM video_projects
		WHERE id = ? AND user_id = ?
	`
	err := s.db.GetContext(ctx, &row, query, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}

	var spec render.VideoSpec
	if err := json.Unmarshal([]byte(row.SpecJSON), &spec); err != nil {
		return nil, fmt.Errorf("unmarshal spec: %w", err)
	}
	row.Spec = &spec
	return &row, nil
}

func (s *Store) ListProjects(ctx context.Context, userID string, limit, offset int) ([]ProjectRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var rows []ProjectRow
	query := `
		SELECT id, user_id, title, spec, spec_version, revision, created_at, updated_at
		FROM video_projects
		WHERE user_id = ?
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`
	err := s.db.SelectContext(ctx, &rows, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	for i := range rows {
		var spec render.VideoSpec
		if err := json.Unmarshal([]byte(rows[i].SpecJSON), &spec); err == nil {
			rows[i].Spec = &spec
		}
	}
	return rows, nil
}

func (s *Store) UpdateProject(ctx context.Context, id, userID, title string, spec *render.VideoSpec, expectedRevision int) (*ProjectRow, error) {
	specRaw, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal spec: %w", err)
	}

	query := `
		UPDATE video_projects
		SET title = ?, spec = ?, revision = revision + 1, updated_at = ?
		WHERE id = ? AND user_id = ? AND revision = ?
	`
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, query, title, string(specRaw), now, id, userID, expectedRevision)
	if err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if aff == 0 {
		// Either not found or revision conflict
		existing, getErr := s.GetProject(ctx, id, userID)
		if getErr != nil {
			return nil, ErrNotFound
		}
		if existing.Revision != expectedRevision {
			return nil, ErrConflict
		}
		return nil, ErrNotFound
	}

	return s.GetProject(ctx, id, userID)
}

func (s *Store) DeleteProject(ctx context.Context, id, userID string) error {
	query := `DELETE FROM video_projects WHERE id = ? AND user_id = ?`
	res, err := s.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

// ---- Jobs ----

func (s *Store) CreateJob(ctx context.Context, userID, projectID, resolution string, spec *render.VideoSpec) (*JobRow, error) {
	specRaw, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal spec: %w", err)
	}

	id := uuid.NewString()
	now := time.Now().UTC()

	var pID sql.NullString
	if projectID != "" {
		pID = sql.NullString{String: projectID, Valid: true}
	}

	query := `
		INSERT INTO video_jobs (id, project_id, user_id, spec, resolution, status, progress, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'queued', 0, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query, id, pID, userID, string(specRaw), resolution, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert job: %w", err)
	}

	return &JobRow{
		ID:         id,
		ProjectID:  pID,
		UserID:     userID,
		Spec:       spec,
		Resolution: resolution,
		Status:     "queued",
		Progress:   0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (s *Store) GetJob(ctx context.Context, id, userID string) (*JobRow, error) {
	var row JobRow
	query := `
		SELECT id, project_id, user_id, spec, resolution, status, progress, output_key, error, started_at, finished_at, created_at, updated_at
		FROM video_jobs
		WHERE id = ? AND user_id = ?
	`
	err := s.db.GetContext(ctx, &row, query, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get job: %w", err)
	}

	var spec render.VideoSpec
	if err := json.Unmarshal([]byte(row.SpecJSON), &spec); err == nil {
		row.Spec = &spec
	}
	return &row, nil
}

func (s *Store) UpdateJobStatus(ctx context.Context, id, status string, progress int, outputKey, errStr string) error {
	now := time.Now().UTC()
	var startedAt, finishedAt *time.Time
	if status == "processing" {
		startedAt = &now
	} else if status == "done" || status == "failed" {
		finishedAt = &now
	}

	query := `
		UPDATE video_jobs
		SET status = ?,
		    progress = ?,
		    output_key = COALESCE(?, output_key),
		    error = ?,
		    started_at = COALESCE(?, started_at),
		    finished_at = COALESCE(?, finished_at),
		    updated_at = ?
		WHERE id = ?
	`
	var outKeyNull sql.NullString
	if outputKey != "" {
		outKeyNull = sql.NullString{String: outputKey, Valid: true}
	}
	var errNull sql.NullString
	if errStr != "" {
		errNull = sql.NullString{String: errStr, Valid: true}
	}

	_, err := s.db.ExecContext(ctx, query, status, progress, outKeyNull, errNull, startedAt, finishedAt, now, id)
	return err
}

// ---- Audio Presets ----

func (s *Store) ListAudioPresets(ctx context.Context) ([]AudioPresetRow, error) {
	var rows []AudioPresetRow
	query := `
		SELECT id, title, artist, s3_key, duration, license, source_url, attribution, created_at
		FROM video_audio_presets
		ORDER BY title ASC
	`
	err := s.db.SelectContext(ctx, &rows, query)
	if err != nil {
		return nil, fmt.Errorf("list audio presets: %w", err)
	}
	return rows, nil
}
