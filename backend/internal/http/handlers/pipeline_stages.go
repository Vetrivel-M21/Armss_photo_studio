package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"armss-poster-studio/internal/agents"
	"armss-poster-studio/internal/http/dto"
	"armss-poster-studio/internal/http/middleware"
	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/pipeline"
	"armss-poster-studio/internal/ratelimit"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
)

type PipelineStagesHandler struct {
	Orchestrator *pipeline.Orchestrator
	Projects     *mysql.ProjectsRepo
	StageOutputs *mysql.StageOutputsRepo
	Renders      *mysql.RendersRepo
	Store        *storage.Store
	Limiter      *ratelimit.Limiter
}

func NewPipelineStagesHandler(o *pipeline.Orchestrator, projects *mysql.ProjectsRepo, stageOutputs *mysql.StageOutputsRepo, renders *mysql.RendersRepo, store *storage.Store, limiter *ratelimit.Limiter) *PipelineStagesHandler {
	return &PipelineStagesHandler{Orchestrator: o, Projects: projects, StageOutputs: stageOutputs, Renders: renders, Store: store, Limiter: limiter}
}

func (h *PipelineStagesHandler) RerunBrief(w http.ResponseWriter, r *http.Request) {
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
	id := chi.URLParam(r, "id")
	var req dto.RerunBriefRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // extraDetail is optional; a missing/empty body is fine

	project, err := h.Orchestrator.RerunBrief(r.Context(), id, req.ExtraDetail)
	if !writeAgentError(w, r.Context(), err) {
		return
	}
	h.respondDetail(w, r, project)
}

func (h *PipelineStagesHandler) SelectConcept(w http.ResponseWriter, r *http.Request) {
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
	id := chi.URLParam(r, "id")
	n, err := strconv.Atoi(chi.URLParam(r, "n"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Concept number must be 0, 1, or 2.")
		return
	}

	project, err := h.Orchestrator.SelectConcept(r.Context(), id, n)
	if !writeAgentError(w, r.Context(), err) {
		return
	}
	h.respondDetail(w, r, project)
}

func (h *PipelineStagesHandler) PatchCopy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body agents.Copy
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "Request body must be valid copy JSON (headline, subhead, cta, ...).")
		return
	}

	project, err := h.Orchestrator.PatchCopy(r.Context(), id, body)
	if !writeAgentError(w, r.Context(), err) {
		return
	}
	h.respondDetail(w, r, project)
}

func (h *PipelineStagesHandler) RerunArtDirection(w http.ResponseWriter, r *http.Request) {
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
	id := chi.URLParam(r, "id")
	var req dto.RerunArtDirectionRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // excludedObjects is optional; a missing/empty body is fine

	project, err := h.Orchestrator.RerunArtDirection(r.Context(), id, req.ExcludedObjects)
	if !writeAgentError(w, r.Context(), err) {
		return
	}
	h.respondDetail(w, r, project)
}

func (h *PipelineStagesHandler) PatchArtDirection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body agents.ArtDirection
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "Request body must be valid art direction JSON.")
		return
	}

	project, err := h.Orchestrator.PatchArtDirection(r.Context(), id, body)
	if !writeAgentError(w, r.Context(), err) {
		return
	}
	h.respondDetail(w, r, project)
}

func (h *PipelineStagesHandler) respondDetail(w http.ResponseWriter, r *http.Request, project *mysql.Project) {
	detail, err := buildProjectDetail(r.Context(), h.StageOutputs, h.Renders, h.Store, project)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Stage updated, but its details could not be loaded. Refresh to see it.")
		return
	}
	respond.JSON(w, http.StatusOK, detail)
}
