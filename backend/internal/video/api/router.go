package api

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts all video module routes onto the given chi router.
// Call this within an authenticated route group.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/video", func(r chi.Router) {
		r.Post("/projects", h.CreateProject)
		r.Get("/projects", h.ListProjects)
		r.Get("/projects/{id}", h.GetProject)
		r.Put("/projects/{id}", h.UpdateProject)
		r.Delete("/projects/{id}", h.DeleteProject)

		r.Post("/jobs", h.CreateJob)
		r.Get("/jobs/{id}", h.GetJob)

		r.Get("/audio-presets", h.ListAudioPresets)
		r.Post("/audio-uploads", h.UploadAudio)
	})
}
