package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"armss-poster-studio/internal/agents"
	"armss-poster-studio/internal/compositor"
	"armss-poster-studio/internal/concurrency"
	"armss-poster-studio/internal/http/dto"
	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/pipeline"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
)

type CompositeHandler struct {
	Compositor   *compositor.Compositor
	Assets       *compositor.AssetLoader
	Projects     *mysql.ProjectsRepo
	StageOutputs *mysql.StageOutputsRepo
	Renders      *mysql.RendersRepo
	Idempotency  *mysql.IdempotencyRepo
	Store        *storage.Store
	Bus          *pipeline.Bus
	Locks        *concurrency.Registry
}

func NewCompositeHandler(c *compositor.Compositor, assets *compositor.AssetLoader, projects *mysql.ProjectsRepo, stageOutputs *mysql.StageOutputsRepo, renders *mysql.RendersRepo, idempotency *mysql.IdempotencyRepo, store *storage.Store, bus *pipeline.Bus, locks *concurrency.Registry) *CompositeHandler {
	return &CompositeHandler{Compositor: c, Assets: assets, Projects: projects, StageOutputs: stageOutputs, Renders: renders, Idempotency: idempotency, Store: store, Bus: bus, Locks: locks}
}

// Composite implements POST /projects/:id/composite (spec §6/§9): overlays
// the real brand assets into the reserved zones of the latest raw render,
// keeping the raw version alongside for the UI's raw/composited toggle.
func (h *CompositeHandler) Composite(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		respond.Error(w, http.StatusBadRequest, "An Idempotency-Key header is required so a retry can't double-composite.")
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
			respond.Error(w, http.StatusConflict, "A compositing request with this idempotency key is already in progress. Please wait.")
			return
		}
		defer unlockKey()

		unlockProject, ok := h.Locks.TryLock(projectID, "composite")
		if !ok {
			respond.Error(w, http.StatusConflict, "A compositing job is already in progress for this project. Please wait for it to complete.")
			return
		}
		defer unlockProject()
	}

	var req dto.CompositeRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // zoneOverrides is optional

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
		respond.Error(w, http.StatusConflict, "This project doesn't have art direction yet.")
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

	rawRender, err := latestRenderOfKind(r.Context(), h.Renders, projectID, mysql.RenderKindRaw)
	if err != nil {
		respond.Error(w, http.StatusConflict, "Generate an image before compositing it.")
		return
	}

	rawBytes, err := h.Store.Get(r.Context(), rawRender.ObjectKey)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load the generated image right now.")
		return
	}

	assets, err := h.Assets.Load(r.Context())
	if err != nil {
		respond.Error(w, http.StatusServiceUnavailable, "Brand assets haven't been uploaded yet. An admin needs to add them first.")
		return
	}
	// Copy before mutating: assets is a pointer into AssetLoader's shared,
	// mutex-guarded cache. Overriding .Logo in place would leak this
	// project's logo into every other concurrent/future composite call.
	resolvedAssets := *assets
	if project.LogoObjectKey.Valid {
		logoBytes, err := h.Store.Get(r.Context(), project.LogoObjectKey.String)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not load this poster's logo right now.")
			return
		}
		resolvedAssets.Logo = logoBytes
	}
	footerText := ""
	if project.FooterText.Valid {
		footerText = project.FooterText.String
	}

	overrides := toCompositorOverrides(req.ZoneOverrides)

	h.Bus.Publish(projectID, pipeline.Event{Stage: "composite", Status: pipeline.StatusRunning})
	composited, err := h.Compositor.Composite(rawBytes, art.ReservedZones, overrides, resolvedAssets, footerText)
	if err != nil {
		h.Bus.Publish(projectID, pipeline.Event{Stage: "composite", Status: pipeline.StatusError})
		respond.Error(w, http.StatusInternalServerError, "Could not composite the brand assets onto this image. You can retry.")
		return
	}

	objectKey := fmt.Sprintf("renders/%s/%s-composited.png", projectID, uuid.NewString())
	if err := h.Store.Put(r.Context(), objectKey, composited, "image/png"); err != nil {
		h.Bus.Publish(projectID, pipeline.Event{Stage: "composite", Status: pipeline.StatusError})
		respond.Error(w, http.StatusInternalServerError, "The composited image could not be saved. You can retry.")
		return
	}

	provider := ""
	if rawRender.Provider.Valid {
		provider = rawRender.Provider.String
	}
	prompt := ""
	if rawRender.Prompt.Valid {
		prompt = rawRender.Prompt.String
	}
	if _, err := h.Renders.Create(r.Context(), projectID, mysql.RenderKindComposited, provider, objectKey, prompt, 0, req.ZoneOverrides); err != nil {
		respond.Error(w, http.StatusInternalServerError, "The composited image was saved but could not be recorded.")
		return
	}
	if err := h.Projects.SetState(r.Context(), projectID, mysql.StateComposited); err != nil {
		respond.Error(w, http.StatusInternalServerError, "The composited image was saved but the project state could not be updated.")
		return
	}
	h.Bus.Publish(projectID, pipeline.Event{Stage: "composite", Status: pipeline.StatusDone})

	detail, err := buildProjectDetail(r.Context(), h.StageOutputs, h.Renders, h.Store, project)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Composite finished, but its details could not be loaded. Refresh to see it.")
		return
	}
	if err := h.Idempotency.Save(r.Context(), idempotencyKey, projectID, "composite", detail); err != nil {
		respond.Error(w, http.StatusInternalServerError, "Composite finished, but could not be recorded for retry-safety.")
		return
	}
	respond.JSON(w, http.StatusOK, detail)
}

// latestRenderOfKind returns the newest render of the given kind. Renders are
// listed newest-first (see RendersRepo.ListByProject), so the first match wins.
func latestRenderOfKind(ctx context.Context, renders *mysql.RendersRepo, projectID, kind string) (*mysql.Render, error) {
	rows, err := renders.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list renders: %w", err)
	}
	for i := range rows {
		if rows[i].Kind == kind {
			return &rows[i], nil
		}
	}
	return nil, mysql.ErrNotFound
}

func toCompositorOverrides(reqOverrides map[string]dto.ZoneOverrideRequest) compositor.Overrides {
	if len(reqOverrides) == 0 {
		return nil
	}
	out := make(compositor.Overrides, len(reqOverrides))
	for zoneID, o := range reqOverrides {
		out[zoneID] = compositor.ZoneOverride{X: o.X, Y: o.Y, W: o.W, H: o.H}
	}
	return out
}
