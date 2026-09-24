package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"armss-poster-studio/internal/http/respond"
)

// FilesHandler serves objects from the local-disk storage backend
// (storage.NewLocal) back over HTTP, standing in for the real presigned URLs
// MinIO/S3 would issue. Only wired up when running without Docker.
type FilesHandler struct {
	RootDir string
}

func NewFilesHandler(rootDir string) *FilesHandler {
	return &FilesHandler{RootDir: rootDir}
}

func (h *FilesHandler) Serve(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "*")
	cleaned := filepath.Clean("/" + key)
	path := filepath.Join(h.RootDir, cleaned)
	if !strings.HasPrefix(path, filepath.Clean(h.RootDir)+string(filepath.Separator)) {
		respond.Error(w, http.StatusBadRequest, "Invalid file path.")
		return
	}
	http.ServeFile(w, r, path)
}
