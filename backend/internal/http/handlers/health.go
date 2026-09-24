package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"

	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
)

type HealthHandler struct {
	DB    *sqlx.DB
	Store *storage.Store
}

func NewHealthHandler(db *sqlx.DB, store *storage.Store) *HealthHandler {
	return &HealthHandler{DB: db, Store: store}
}

// Healthz reports whether the process itself is up, with no dependency checks.
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readyz reports whether MySQL and the object store are both reachable.
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if err := mysql.Ping(r.Context(), h.DB); err != nil {
		respond.Error(w, http.StatusServiceUnavailable, "Database is not reachable.")
		return
	}
	if err := h.Store.Ping(r.Context()); err != nil {
		respond.Error(w, http.StatusServiceUnavailable, "Object storage is not reachable.")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
