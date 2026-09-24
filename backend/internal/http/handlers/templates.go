package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"armss-poster-studio/internal/agents"
	"armss-poster-studio/internal/http/dto"
	"armss-poster-studio/internal/http/middleware"
	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
	"armss-poster-studio/internal/templates"
)

const templateStyleProfileStatusPendingManual = "pending_manual"
const templateStyleProfileStatusReady = "ready"

type TemplatesHandler struct {
	Templates *mysql.TemplatesRepo
	Store     *storage.Store
	Agents    *agents.Client
}

func NewTemplatesHandler(templates *mysql.TemplatesRepo, store *storage.Store, agentsClient *agents.Client) *TemplatesHandler {
	return &TemplatesHandler{Templates: templates, Store: store, Agents: agentsClient}
}

func (h *TemplatesHandler) List(w http.ResponseWriter, r *http.Request) {
	occasion := r.URL.Query().Get("occasion")
	q := r.URL.Query().Get("q")
	if occasion == "" {
		occasion = q
	}

	// Admins manage the library and need to see (and reactivate) deactivated
	// templates too; the public gallery (staff) only ever sees active ones.
	user := middleware.UserFromContext(r.Context())
	var templates []mysql.Template
	var err error
	if user != nil && user.Role == mysql.RoleAdmin {
		templates, err = h.Templates.ListAll(r.Context(), occasion)
	} else {
		templates, err = h.Templates.ListActive(r.Context(), occasion)
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load templates right now.")
		return
	}

	out := make([]dto.TemplateResponse, 0, len(templates))
	for _, t := range templates {
		view, err := h.toResponse(r.Context(), t)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not load templates right now.")
			return
		}
		out = append(out, view)
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *TemplatesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	t, err := h.Templates.GetByID(r.Context(), id)
	if errors.Is(err, mysql.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "Template not found.")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load this template right now.")
		return
	}
	view, err := h.toResponse(r.Context(), *t)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load this template right now.")
		return
	}
	respond.JSON(w, http.StatusOK, view)
}

// Create implements the admin upload flow (spec §8): attach image → run Style
// Analyst → admin reviews and edits. If the Anthropic client is disabled, the
// template is still created with a null profile so the CRUD/gallery path stays
// usable with zero keys — an admin fills the profile in later via Patch.
//
// The "file" field is optional: a template can instead be created from a
// composition prompt alone (no reference image), via "compositionPrompt" —
// e.g. a hand-written art-direction template. It's stored directly as the
// style profile's Summary, skipping Style Analyst entirely since there's no
// image to analyze.
func (h *TemplatesHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		respond.Error(w, http.StatusBadRequest, "Upload is too large. The limit is 10MB.")
		return
	}
	title := r.FormValue("title")
	if title == "" {
		respond.Error(w, http.StatusBadRequest, "A title is required.")
		return
	}
	season := r.FormValue("season")
	var occasionTags []string
	if raw := r.FormValue("occasionTags"); raw != "" {
		for _, tag := range strings.Split(raw, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				occasionTags = append(occasionTags, tag)
			}
		}
	}

	file, _, fileErr := r.FormFile("file")
	if fileErr != nil {
		compositionPrompt := strings.TrimSpace(r.FormValue("compositionPrompt"))
		if compositionPrompt == "" {
			respond.Error(w, http.StatusBadRequest, "Attach a reference image or write a composition prompt.")
			return
		}
		created, err := h.createPromptOnlyTemplate(r.Context(), title, occasionTags, season, compositionPrompt, user.ID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not save the template right now.")
			return
		}
		view, err := h.toResponse(r.Context(), *created)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "Template was created but its details could not be loaded.")
			return
		}
		respond.JSON(w, http.StatusCreated, view)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil || len(data) > maxUploadBytes {
		respond.Error(w, http.StatusBadRequest, "Upload is too large. The limit is 10MB.")
		return
	}
	contentType := storage.SniffContentType(data)
	if !storage.AllowedContentTypes[contentType] {
		respond.Error(w, http.StatusUnsupportedMediaType, "Only PNG, JPEG, and WebP images are accepted.")
		return
	}

	downscaled, err := storage.DownscaleLongEdge(data, 1500)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "That file doesn't look like a valid image.")
		return
	}

	baseKey := fmt.Sprintf("templates/%s", uuid.NewString())
	objectKey := baseKey + "/full"
	thumbKey := baseKey + "/thumb"
	if err := h.Store.Put(r.Context(), objectKey, downscaled, contentType); err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not store the template image right now.")
		return
	}
	thumb, err := storage.DownscaleLongEdge(downscaled, 400)
	if err != nil {
		thumb = downscaled
	}
	if err := h.Store.Put(r.Context(), thumbKey, thumb, contentType); err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not store the template thumbnail right now.")
		return
	}

	var styleProfileJSON json.RawMessage
	if h.Agents.Enabled() {
		profile, _, err := h.Agents.AnalyzeStyle(r.Context(), []agents.ImageInput{{Data: downscaled, MediaType: contentType}})
		if err != nil {
			respond.Error(w, http.StatusBadGateway, "The image was saved, but style analysis failed. You can fill in the style profile by hand.")
			return
		}
		styleProfileJSON, err = json.Marshal(profile)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not save the analysed style profile.")
			return
		}
	}

	created, err := h.Templates.Create(r.Context(), title, objectKey, thumbKey, occasionTags, season, styleProfileJSON, user.ID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not save the template right now.")
		return
	}

	view, err := h.toResponse(r.Context(), *created)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Template was created but its details could not be loaded.")
		return
	}
	respond.JSON(w, http.StatusCreated, view)
}

// ImportStarterLibrary bulk-creates the bundled starter template library
// (spec: prompt-only templates, no reference image). Idempotent on repeat
// calls — a title that already exists is skipped rather than duplicated.
func (h *TemplatesHandler) ImportStarterLibrary(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	imported, skipped := 0, 0
	for _, tmpl := range templates.StarterLibrary() {
		exists, err := h.Templates.ExistsByTitle(r.Context(), tmpl.Title)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not check existing templates right now.")
			return
		}
		if exists {
			skipped++
			continue
		}
		if _, err := h.createPromptOnlyTemplate(r.Context(), tmpl.Title, []string{tmpl.Category}, "", tmpl.Prompt, user.ID); err != nil {
			respond.Error(w, http.StatusInternalServerError, fmt.Sprintf("Could not import %q. %d templates were imported before this failure.", tmpl.Title, imported))
			return
		}
		imported++
	}

	respond.JSON(w, http.StatusOK, map[string]int{"imported": imported, "skipped": skipped})
}

// createPromptOnlyTemplate builds a template directly from a composition
// prompt, with no image and no Style Analyst call — used by both the
// prompt-only path in Create and the starter-library bulk import.
func (h *TemplatesHandler) createPromptOnlyTemplate(ctx context.Context, title string, occasionTags []string, season, compositionPrompt, createdBy string) (*mysql.Template, error) {
	profile := agents.StyleProfile{
		Summary:       compositionPrompt,
		Palette:       []agents.PaletteHex{},
		FixedElements: []string{},
	}
	styleProfileJSON, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("marshal composition prompt profile: %w", err)
	}
	return h.Templates.Create(ctx, title, "", "", occasionTags, season, styleProfileJSON, createdBy)
}

// Patch lets an admin edit the extracted style profile by hand, adjust tags,
// or deactivate a template (spec §9).
func (h *TemplatesHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.PatchTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Request body must be valid JSON.")
		return
	}

	if _, err := h.Templates.GetByID(r.Context(), id); errors.Is(err, mysql.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "Template not found.")
		return
	} else if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load this template right now.")
		return
	}

	if len(req.StyleProfile) > 0 {
		if err := h.Templates.UpdateStyleProfile(r.Context(), id, req.StyleProfile); err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not save the style profile.")
			return
		}
	}
	if req.OccasionTags != nil || req.Season != nil {
		season := ""
		if req.Season != nil {
			season = *req.Season
		}
		tags := []string{}
		if req.OccasionTags != nil {
			tags = *req.OccasionTags
		}
		if err := h.Templates.UpdateTags(r.Context(), id, tags, season); err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not save the template tags.")
			return
		}
	}
	if req.IsActive != nil {
		if err := h.Templates.SetActive(r.Context(), id, *req.IsActive); err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not update the template's active state.")
			return
		}
	}

	updated, err := h.Templates.GetByID(r.Context(), id)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Template was updated but its details could not be loaded.")
		return
	}
	view, err := h.toResponse(r.Context(), *updated)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load this template's details right now.")
		return
	}
	respond.JSON(w, http.StatusOK, view)
}

func (h *TemplatesHandler) toResponse(ctx context.Context, t mysql.Template) (dto.TemplateResponse, error) {
	return templateToResponse(ctx, h.Store, t)
}

// templateToResponse is package-level so recommendations.go can build the
// same DTO shape without needing a TemplatesHandler.
func templateToResponse(ctx context.Context, store *storage.Store, t mysql.Template) (dto.TemplateResponse, error) {
	// Prompt-only templates (spec: a text-prompt template needs no reference
	// image at all — the file it came from treats a preview as optional and
	// generated later) have no object_key; leave ImageURL/ThumbURL empty
	// rather than signing a URL for a key that doesn't exist.
	var imageURL, thumbURL string
	var err error
	if t.ObjectKey.Valid {
		imageURL, err = store.SignedURL(ctx, t.ObjectKey.String, 15*time.Minute)
		if err != nil {
			return dto.TemplateResponse{}, fmt.Errorf("sign template image url: %w", err)
		}
	}
	if t.ThumbKey.Valid {
		thumbURL, err = store.SignedURL(ctx, t.ThumbKey.String, 15*time.Minute)
		if err != nil {
			return dto.TemplateResponse{}, fmt.Errorf("sign template thumb url: %w", err)
		}
	}

	var tags []string
	if len(t.OccasionTags) > 0 {
		if err := json.Unmarshal(t.OccasionTags, &tags); err != nil {
			return dto.TemplateResponse{}, fmt.Errorf("unmarshal occasion tags: %w", err)
		}
	}

	status := templateStyleProfileStatusReady
	if len(t.StyleProfile) == 0 {
		status = templateStyleProfileStatusPendingManual
	}

	view := dto.TemplateResponse{
		ID:                 t.ID,
		Title:              t.Title,
		ImageURL:           imageURL,
		ThumbURL:           thumbURL,
		OccasionTags:       tags,
		StyleProfile:       t.StyleProfile,
		StyleProfileStatus: status,
		IsActive:           t.IsActive,
		CreatedAt:          t.CreatedAt,
	}
	if t.Season.Valid {
		view.Season = t.Season.String
	}
	return view, nil
}
