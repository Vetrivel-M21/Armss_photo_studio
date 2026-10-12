package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"armss-poster-studio/internal/agents"
	"armss-poster-studio/internal/concurrency"
	"armss-poster-studio/internal/http/dto"
	"armss-poster-studio/internal/http/middleware"
	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/imagegen"
	"armss-poster-studio/internal/pipeline"
	"armss-poster-studio/internal/ratelimit"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
)

type RenderHandler struct {
	Registry     *imagegen.Registry
	Projects     *mysql.ProjectsRepo
	StageOutputs *mysql.StageOutputsRepo
	Renders      *mysql.RendersRepo
	Idempotency  *mysql.IdempotencyRepo
	Store        *storage.Store
	Bus          *pipeline.Bus
	Locks        *concurrency.Registry
	Limiter      *ratelimit.Limiter
}

func NewRenderHandler(registry *imagegen.Registry, projects *mysql.ProjectsRepo, stageOutputs *mysql.StageOutputsRepo, renders *mysql.RendersRepo, idempotency *mysql.IdempotencyRepo, store *storage.Store, bus *pipeline.Bus, locks *concurrency.Registry, limiter *ratelimit.Limiter) *RenderHandler {
	return &RenderHandler{Registry: registry, Projects: projects, StageOutputs: stageOutputs, Renders: renders, Idempotency: idempotency, Store: store, Bus: bus, Locks: locks, Limiter: limiter}
}

// Render implements POST /projects/:id/render (spec §9). Image generation is
// the one step that costs real money per click, so it always requires an
// Idempotency-Key (spec §13) and never runs more than once per user action.
func (h *RenderHandler) Render(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		respond.Error(w, http.StatusBadRequest, "An Idempotency-Key header is required so a retry can't generate a second image.")
		return
	}

	if cached, err := h.Idempotency.Get(r.Context(), idempotencyKey); err == nil {
		respond.JSON(w, http.StatusOK, json.RawMessage(cached))
		return
	} else if !errors.Is(err, mysql.ErrNotFound) {
		respond.Error(w, http.StatusInternalServerError, "Could not check whether this request was already processed.")
		return
	}

	if h.Locks != nil {
		unlockKey, ok := h.Locks.TryLock("idempotency", idempotencyKey)
		if !ok {
			respond.Error(w, http.StatusConflict, "A generation request with this idempotency key is already in progress. Please wait.")
			return
		}
		defer unlockKey()

		unlockProject, ok := h.Locks.TryLock(projectID, "render")
		if !ok {
			respond.Error(w, http.StatusConflict, "An image generation is already in progress for this project. Please wait for it to complete.")
			return
		}
		defer unlockProject()
	}

	user := middleware.UserFromContext(r.Context())
	if h.Limiter != nil && user != nil {
		chk, err := h.Limiter.CheckAndRecord(r.Context(), user.ID, mysql.ActionRender)
		if err == nil && !chk.Allowed {
			if chk.RetryAfterSeconds > 0 {
				w.Header().Set("Retry-After", strconv.Itoa(chk.RetryAfterSeconds))
			}
			respond.Error(w, http.StatusTooManyRequests, chk.ErrorMessage)
			return
		}
	}

	var req dto.RenderRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // provider is optional; a missing/empty body picks the server default

	project, err := h.Projects.GetByID(r.Context(), projectID)
	if errors.Is(err, mysql.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "Project not found.")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load this project right now.")
		return
	}

	artRow, err := h.StageOutputs.GetLatest(r.Context(), projectID, mysql.StageArt)
	if errors.Is(err, mysql.ErrNotFound) {
		respond.Error(w, http.StatusConflict, "This project doesn't have art direction yet. Select a concept first.")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load this project's art direction.")
		return
	}
	var art agents.ArtDirection
	if err := json.Unmarshal(artRow.Payload, &art); err != nil {
		respond.Error(w, http.StatusInternalServerError, "This project's art direction could not be read.")
		return
	}

	provider, err := h.Registry.Get(req.Provider)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Unknown image provider requested.")
		return
	}
	if !provider.Enabled() {
		respond.Error(w, http.StatusBadRequest, fmt.Sprintf("%s image generation is disabled (no API key configured). Please add an API key in Admin Settings.", provider.Name()))
		return
	}

	prompt := art.FullPosterPrompt
	if !provider.SupportsTextRendering() {
		prompt = art.BackgroundPrompt
	}

	h.Bus.Publish(projectID, pipeline.Event{Stage: "render", Status: pipeline.StatusRunning})
	images, err := provider.Generate(r.Context(), imagegen.GenerateRequest{
		Prompt:         prompt,
		NegativePrompt: art.NegativePrompt,
		AspectRatio:    art.AspectRatio,
		N:              1,
	})
	if !writeProviderError(w, provider.Name(), err) {
		h.Bus.Publish(projectID, pipeline.Event{Stage: "render", Status: pipeline.StatusError})
		return
	}
	if len(images) == 0 {
		h.Bus.Publish(projectID, pipeline.Event{Stage: "render", Status: pipeline.StatusError})
		respond.Error(w, http.StatusBadGateway, fmt.Sprintf("%s did not return an image. You can retry.", provider.Name()))
		return
	}
	image := images[0]

	objectKey := fmt.Sprintf("renders/%s/%s-raw.png", projectID, uuid.NewString())
	if err := h.Store.Put(r.Context(), objectKey, image.Data, image.ContentType); err != nil {
		respond.Error(w, http.StatusInternalServerError, "The image was generated but could not be saved. You can retry.")
		return
	}

	if _, err := h.Renders.Create(r.Context(), projectID, mysql.RenderKindRaw, provider.Name(), objectKey, prompt, 0, nil); err != nil {
		respond.Error(w, http.StatusInternalServerError, fmt.Sprintf("The image was generated but could not be recorded (%v). You can retry.", err))
		return
	}
	if err := h.Projects.SetState(r.Context(), projectID, mysql.StateRenderReady); err != nil {
		respond.Error(w, http.StatusInternalServerError, fmt.Sprintf("The image was saved but the project state could not be updated (%v).", err))
		return
	}
	h.Bus.Publish(projectID, pipeline.Event{Stage: "render", Status: pipeline.StatusDone})

	detail, err := buildProjectDetail(r.Context(), h.StageOutputs, h.Renders, h.Store, project)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, fmt.Sprintf("Render finished, but its details could not be loaded (%v). Refresh to see it.", err))
		return
	}
	if err := h.Idempotency.Save(r.Context(), idempotencyKey, projectID, "render", detail); err != nil {
		respond.Error(w, http.StatusInternalServerError, "Render finished, but could not be recorded for retry-safety.")
		return
	}
	respond.JSON(w, http.StatusOK, detail)
}

// writeProviderError maps an ImageProvider error to a user-facing response
// naming the failing provider (spec §7/§13). Returns true if err was nil.
func writeProviderError(w http.ResponseWriter, providerName string, err error) bool {
	if err == nil {
		return true
	}
	var perr *imagegen.ProviderError
	if errors.As(err, &perr) {
		if strings.Contains(perr.Message, "disabled") || strings.Contains(perr.Message, "no API key") {
			respond.Error(w, http.StatusServiceUnavailable, fmt.Sprintf("Image generation isn't configured yet. An admin needs to add an API key for %s.", perr.Provider))
			return false
		}
		status := http.StatusBadGateway
		if perr.Retryable {
			status = http.StatusServiceUnavailable
		}
		respond.Error(w, status, fmt.Sprintf("%s image generation failed: %s. You can retry.", perr.Provider, perr.Message))
		return false
	}
	respond.Error(w, http.StatusInternalServerError, fmt.Sprintf("%s failed unexpectedly: %s. You can retry.", providerName, err.Error()))
	return false
}
