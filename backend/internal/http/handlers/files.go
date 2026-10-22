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
	cleanKey := filepath.Clean(strings.TrimLeft(key, "/\\"))
	absRoot, err := filepath.Abs(h.RootDir)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Storage error.")
		return
	}
	path := filepath.Join(absRoot, cleanKey)
	if !strings.HasPrefix(path, absRoot+string(filepath.Separator)) && path != absRoot {
		respond.Error(w, http.StatusBadRequest, "Invalid file path.")
		return
	}
	http.ServeFile(w, r, path)
}
