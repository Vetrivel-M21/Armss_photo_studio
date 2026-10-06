package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"armss-poster-studio/internal/agents"
	"armss-poster-studio/internal/concurrency"
	"armss-poster-studio/internal/http/dto"
	"armss-poster-studio/internal/http/middleware"
	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/observability"
	"armss-poster-studio/internal/pipeline"
	"armss-poster-studio/internal/ratelimit"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
)

type ProjectsHandler struct {
	Orchestrator *pipeline.Orchestrator
	Projects     *mysql.ProjectsRepo
	StageOutputs *mysql.StageOutputsRepo
	Renders      *mysql.RendersRepo
	Store        *storage.Store
	Limiter      *ratelimit.Limiter
}

func NewProjectsHandler(o *pipeline.Orchestrator, projects *mysql.ProjectsRepo, stageOutputs *mysql.StageOutputsRepo, renders *mysql.RendersRepo, store *storage.Store, limiter *ratelimit.Limiter) *ProjectsHandler {
	return &ProjectsHandler{Orchestrator: o, Projects: projects, StageOutputs: stageOutputs, Renders: renders, Store: store, Limiter: limiter}
}

func (h *ProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if h.Limiter != nil && user != nil {
		chk, err := h.Limiter.CheckAndRecord(r.Context(), user.ID, mysql.ActionPipeline)
		if err == nil && !chk.Allowed {
			if chk.RetryAfterSeconds > 0 {
				w.Header().Set("Retry-After", strconv.Itoa(chk.RetryAfterSeconds))
			}
			respond.Error(w, http.StatusTooManyRequests, chk.ErrorMessage)
			return
		}
	}
	var req dto.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Request body must be valid JSON with at least a \"text\" field.")
		return
	}
	if req.Text == "" {
		respond.Error(w, http.StatusBadRequest, "Describe the occasion in \"text\" before generating.")
		return
	}

	logoUploadID := normalizeOptional(req.LogoUploadID)
	footerText := normalizeOptional(req.FooterText)
	project, err := h.Orchestrator.CreateProject(r.Context(), user.ID, req.Text, req.TemplateID, req.UploadIDs, logoUploadID, footerText)
	if !writeAgentError(w, r.Context(), err) {
		return
	}

	detail, err := buildProjectDetail(r.Context(), h.StageOutputs, h.Renders, h.Store, project)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Project was created but its details could not be loaded. Refresh to see it.")
		return
	}
	respond.JSON(w, http.StatusCreated, detail)
}

func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	projects, err := h.Projects.ListByUser(r.Context(), user.ID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load your projects right now.")
		return
	}

	summaries := make([]dto.ProjectSummary, 0, len(projects))
	for _, p := range projects {
		summaries = append(summaries, toProjectSummary(p))
	}
	respond.JSON(w, http.StatusOK, summaries)
}

func (h *ProjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	project, ok := h.loadOwnedProject(w, r)
	if !ok {
		return
	}
	detail, err := buildProjectDetail(r.Context(), h.StageOutputs, h.Renders, h.Store, project)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load this project's details right now.")
		return
	}
	respond.JSON(w, http.StatusOK, detail)
}

// Download streams the latest render of the requested kind (default
// "composited") as a file, so the user gets a finished poster with one click.
func (h *ProjectsHandler) Download(w http.ResponseWriter, r *http.Request) {
	project, ok := h.loadOwnedProject(w, r)
	if !ok {
		return
	}
	kind := r.URL.Query().Get("kind")
	if kind == "" {
		kind = mysql.RenderKindComposited
	}

	renders, err := h.Renders.ListByProject(r.Context(), project.ID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load renders for this project.")
		return
	}
	var latest *mysql.Render
	for i := range renders {
		if renders[i].Kind == kind {
			latest = &renders[i]
			break // newest first
		}
	}
	if latest == nil {
		respond.Error(w, http.StatusNotFound, fmt.Sprintf("No %s render exists for this project yet.", kind))
		return
	}

	data, err := h.Store.Get(r.Context(), latest.ObjectKey)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not download the render file right now.")
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s.png"`, project.ID, kind))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *ProjectsHandler) loadOwnedProject(w http.ResponseWriter, r *http.Request) (*mysql.Project, bool) {
	user := middleware.UserFromContext(r.Context())
	id := chi.URLParam(r, "id")

	project, err := h.Projects.GetByID(r.Context(), id)
	if errors.Is(err, mysql.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "Project not found.")
		return nil, false
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load this project right now.")
		return nil, false
	}
	if project.UserID != user.ID && user.Role != mysql.RoleAdmin {
		respond.Error(w, http.StatusForbidden, "You don't have access to this project.")
		return nil, false
	}
	return project, true
}

// normalizeOptional treats an explicit empty string the same as an omitted
// field, so the frontend doesn't need to special-case "" vs. undefined.
func normalizeOptional(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

func toProjectSummary(p mysql.Project) dto.ProjectSummary {
	s := dto.ProjectSummary{
		ID:        p.ID,
		InputText: p.InputText,
		State:     p.State,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if p.Title.Valid {
		s.Title = &p.Title.String
	}
	if p.TemplateID.Valid {
		s.TemplateID = &p.TemplateID.String
	}
	return s
}

// buildProjectDetail assembles the full project state DTO: the current
// (non-superseded) row per stage, and every render with a computed staleness
// flag. Shared by Create/Get and every pipeline-stage handler that returns the
// project after mutating it.
func buildProjectDetail(ctx context.Context, stageOutputs *mysql.StageOutputsRepo, renders *mysql.RendersRepo, store *storage.Store, project *mysql.Project) (*dto.ProjectDetail, error) {
	detail := &dto.ProjectDetail{
		ID:        project.ID,
		InputText: project.InputText,
		State:     project.State,
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
		Stages:    map[string]dto.StageOutputView{},
		// Initialized non-nil so a project with zero renders serializes as
		// "renders": [] rather than null — the frontend calls array methods
		// on this field directly.
		Renders: []dto.RenderView{},
	}
	if project.Title.Valid {
		detail.Title = &project.Title.String
	}
	if project.TemplateID.Valid {
		detail.TemplateID = &project.TemplateID.String
	}
	if project.FooterText.Valid {
		detail.FooterText = &project.FooterText.String
	}
	if project.LogoObjectKey.Valid {
		url, err := store.SignedURL(ctx, project.LogoObjectKey.String, 15*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("sign project logo url: %w", err)
		}
		detail.LogoURL = &url
	}

	stages := []string{
		mysql.StageStyle, mysql.StageBrief, mysql.StageConcepts,
		mysql.StageConceptSelection, mysql.StageCopy, mysql.StageArt,
	}
	var latestArtCreatedAt time.Time
	for _, stage := range stages {
		row, err := stageOutputs.GetLatest(ctx, project.ID, stage)
		if errors.Is(err, mysql.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("load stage %s: %w", stage, err)
		}
		view := dto.StageOutputView{ID: row.ID, Payload: row.Payload, CreatedAt: row.CreatedAt}
		if row.Model.Valid {
			view.Model = &row.Model.String
		}
		if row.TokensIn.Valid {
			v := int(row.TokensIn.Int32)
			view.TokensIn = &v
		}
		if row.TokensOut.Valid {
			v := int(row.TokensOut.Int32)
			view.TokensOut = &v
		}
		detail.Stages[wireStageName(stage)] = view
		if stage == mysql.StageArt {
			latestArtCreatedAt = row.CreatedAt
		}
	}

	renderRows, err := renders.ListByProject(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("list renders: %w", err)
	}
	for _, render := range renderRows {
		url, err := store.SignedURL(ctx, render.ObjectKey, 15*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("sign render url: %w", err)
		}
		view := dto.RenderView{
			ID:        render.ID,
			Kind:      render.Kind,
			URL:       url,
			CreatedAt: render.CreatedAt,
			Stale:     !latestArtCreatedAt.IsZero() && render.CreatedAt.Before(latestArtCreatedAt),
		}
		if render.Provider.Valid {
			view.Provider = &render.Provider.String
		}
		if render.Prompt.Valid {
			view.Prompt = &render.Prompt.String
		}
		if render.CostCents.Valid {
			v := int(render.CostCents.Int32)
			view.CostCents = &v
		}
		detail.Renders = append(detail.Renders, view)
	}

	return detail, nil
}

// wireStageName matches SSE's own renaming (spec §9's example uses
// "art_direction") so the same stage key is used in both the DTO and the
// stream, instead of the internal DB token "art".
func wireStageName(stage string) string {
	if stage == mysql.StageArt {
		return "art_direction"
	}
	return stage
}

// writeAgentError maps a pipeline error to a user-facing response. Returns
// true if err was nil (caller should continue), false if a response was
// already written (caller should return).
func writeAgentError(w http.ResponseWriter, ctx context.Context, err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, concurrency.ErrAlreadyInProgress) {
		respond.Error(w, http.StatusConflict, "A pipeline stage is already processing for this project. Please wait for it to complete.")
		return false
	}
	if errors.Is(err, agents.ErrAgentsUnavailable) {
		respond.Error(w, http.StatusServiceUnavailable, "AI generation isn't configured yet. An admin needs to add an API key for the selected provider before this can run.")
		return false
	}
	// Never leak the raw error to the client (spec §13) — but always log it
	// server-side with the request id, or a failure like this is unfixable.
	observability.FromContext(ctx, slog.Default()).Error("pipeline_step_failed", "error", err)
	respond.Error(w, http.StatusInternalServerError, "That step failed. You can retry it.")
	return false
}
