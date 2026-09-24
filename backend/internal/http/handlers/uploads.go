package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"armss-poster-studio/internal/http/dto"
	"armss-poster-studio/internal/http/middleware"
	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
)

const maxUploadBytes = 10 << 20 // 10MB per file, spec §13

type UploadsHandler struct {
	Uploads *mysql.UploadsRepo
	Store   *storage.Store
}

func NewUploadsHandler(uploads *mysql.UploadsRepo, store *storage.Store) *UploadsHandler {
	return &UploadsHandler{Uploads: uploads, Store: store}
}

// Create implements POST /uploads (multipart, spec §9): validates by content
// sniffing (not extension), downscales to 1500px on the long edge, stores the
// file, and returns a stable upload id the client references later.
func (h *UploadsHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		respond.Error(w, http.StatusBadRequest, "Upload is too large. The limit is 10MB per file.")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "Attach the image under the \"file\" field.")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not read the uploaded file.")
		return
	}
	if len(data) > maxUploadBytes {
		respond.Error(w, http.StatusBadRequest, "Upload is too large. The limit is 10MB per file.")
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

	objectKey := fmt.Sprintf("uploads/%s/%s", user.ID, uuid.NewString())
	if err := h.Store.Put(r.Context(), objectKey, downscaled, contentType); err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not store the uploaded file right now.")
		return
	}

	upload, err := h.Uploads.Create(r.Context(), objectKey, contentType, len(downscaled), user.ID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Upload was stored but could not be recorded. Try again.")
		return
	}

	respond.JSON(w, http.StatusCreated, dto.UploadResponse{
		ID:          upload.ID,
		ContentType: upload.ContentType,
		SizeBytes:   upload.SizeBytes,
	})
}
