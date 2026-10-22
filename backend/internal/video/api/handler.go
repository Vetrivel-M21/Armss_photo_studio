package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"armss-poster-studio/internal/http/middleware"
	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/video/render"
	"armss-poster-studio/internal/video/store"
	"armss-poster-studio/internal/video/validate"
)

type JobEnqueuer interface {
	EnqueueVideoJob(jobID, userID string) error
}

type Handler struct {
	Store    *store.Store
	Storage  *storage.Store
	Limits   validate.Limits
	Enqueuer JobEnqueuer // optional: nil in local dev / sync test mode
}

func NewHandler(st *store.Store, storage *storage.Store, enq JobEnqueuer) *Handler {
	return &Handler{
		Store:    st,
		Storage:  storage,
		Limits:   validate.DefaultLimits(),
		Enqueuer: enq,
	}
}

type CreateProjectRequest struct {
	Title string            `json:"title"`
	Spec  *render.VideoSpec `json:"spec,omitempty"`
}

type UpdateProjectRequest struct {
	Title    string            `json:"title"`
	Spec     *render.VideoSpec `json:"spec"`
	Revision int               `json:"revision"`
}

type CreateJobRequest struct {
	ProjectID  string            `json:"project_id,omitempty"`
	Resolution string            `json:"resolution"`
	Spec       *render.VideoSpec `json:"spec"`
}

func (h *Handler) userPrefix(userID string) string {
	return fmt.Sprintf("video/%s/", userID)
}

// CreateProject handles POST /video/projects
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respond.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		respond.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "Untitled Video"
	}

	spec := req.Spec
	if spec != nil {
		if err := validate.Spec(*spec, h.Limits, h.userPrefix(user.ID)); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
	}

	proj, err := h.Store.CreateProject(r.Context(), user.ID, title, spec)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Failed to create video project")
		return
	}

	respond.JSON(w, http.StatusCreated, proj)
}

// GetProject handles GET /video/projects/{id}
func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respond.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	id := chi.URLParam(r, "id")
	proj, err := h.Store.GetProject(r.Context(), id, user.ID)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "Video project not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Failed to load video project")
		return
	}

	respond.JSON(w, http.StatusOK, proj)
}

// ListProjects handles GET /video/projects
func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respond.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}
	offset := 0
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if parsed, err := strconv.Atoi(oStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	projects, err := h.Store.ListProjects(r.Context(), user.ID, limit, offset)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Failed to list video projects")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]any{
		"projects": projects,
		"limit":    limit,
		"offset":   offset,
	})
}

// UpdateProject handles PUT /video/projects/{id}
func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respond.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	id := chi.URLParam(r, "id")
	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.Spec == nil {
		respond.Error(w, http.StatusBadRequest, "spec is required")
		return
	}

	if err := validate.Spec(*req.Spec, h.Limits, h.userPrefix(user.ID)); err != nil {
		respond.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "Untitled Video"
	}

	proj, err := h.Store.UpdateProject(r.Context(), id, user.ID, title, req.Spec, req.Revision)
	if errors.Is(err, store.ErrConflict) {
		respond.Error(w, http.StatusConflict, "The project was modified by another session. Please reload.")
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "Video project not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Failed to update video project")
		return
	}

	respond.JSON(w, http.StatusOK, proj)
}

// DeleteProject handles DELETE /video/projects/{id}
func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respond.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.Store.DeleteProject(r.Context(), id, user.ID); errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "Video project not found")
		return
	} else if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Failed to delete video project")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// CreateJob handles POST /video/jobs
func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respond.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.Spec == nil {
		respond.Error(w, http.StatusBadRequest, "spec is required")
		return
	}

	if err := validate.Resolution(req.Resolution); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validate.Spec(*req.Spec, h.Limits, h.userPrefix(user.ID)); err != nil {
		respond.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	job, err := h.Store.CreateJob(r.Context(), user.ID, req.ProjectID, req.Resolution, req.Spec)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Failed to create render job")
		return
	}

	if h.Enqueuer != nil {
		if err := h.Enqueuer.EnqueueVideoJob(job.ID, user.ID); err != nil {
			_ = h.Store.UpdateJobStatus(r.Context(), job.ID, "failed", 0, "", "Job enqueuing failed: "+err.Error())
			respond.Error(w, http.StatusInternalServerError, "Failed to enqueue render job")
			return
		}
	}

	respond.JSON(w, http.StatusAccepted, job)
}

// GetJob handles GET /video/jobs/{id}
func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respond.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	id := chi.URLParam(r, "id")
	job, err := h.Store.GetJob(r.Context(), id, user.ID)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "Render job not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Failed to load render job")
		return
	}

	respond.JSON(w, http.StatusOK, job)
}

// ListAudioPresets handles GET /video/audio-presets
func (h *Handler) ListAudioPresets(w http.ResponseWriter, r *http.Request) {
	presets, err := h.Store.ListAudioPresets(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Failed to list audio presets")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]any{
		"presets": presets,
	})
}

const maxAudioBytes = 25 << 20 // 25MB limit

// UploadAudio handles POST /video/audio-uploads
func (h *Handler) UploadAudio(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		respond.Error(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := r.ParseMultipartForm(maxAudioBytes); err != nil {
		respond.Error(w, http.StatusBadRequest, "Audio upload is too large. Limit is 25MB.")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Attach audio file in 'file' form field")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxAudioBytes+1))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Failed to read audio file")
		return
	}
	if len(data) > maxAudioBytes {
		respond.Error(w, http.StatusBadRequest, "Audio upload is too large. Limit is 25MB.")
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".mp3"
	}

	switch ext {
	case ".mp3", ".wav", ".aac", ".m4a", ".ogg":
		// Allowed
	default:
		respond.Error(w, http.StatusBadRequest, "Only MP3, WAV, AAC, M4A, or OGG audio files are supported")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "audio/mpeg"
	}

	objectKey := fmt.Sprintf("video/%s/audio/%s%s", user.ID, uuid.NewString(), ext)
	if err := h.Storage.Put(r.Context(), objectKey, data, contentType); err != nil {
		respond.Error(w, http.StatusInternalServerError, "Failed to store audio file")
		return
	}

	respond.JSON(w, http.StatusCreated, map[string]any{
		"key":         objectKey,
		"url":         "/files/" + objectKey,
		"filename":    header.Filename,
		"sizeBytes":   len(data),
		"contentType": contentType,
	})
}
