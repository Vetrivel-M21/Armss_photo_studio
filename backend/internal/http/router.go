// Package http assembles the chi router, middleware chain, and every HTTP handler.
package http

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	authpkg "armss-poster-studio/internal/auth"
	"armss-poster-studio/internal/http/handlers"
	"armss-poster-studio/internal/http/middleware"
)

// Deps carries every handler the router wires up. Fields are added as each
// build step introduces new handlers (templates, uploads, render, ...).
type Deps struct {
	Logger          *slog.Logger
	AuthManager     *authpkg.Manager
	CORSOrigin      string
	Health          *handlers.HealthHandler
	Auth            *handlers.AuthHandler
	Projects        *handlers.ProjectsHandler
	PipelineStages  *handlers.PipelineStagesHandler
	SSE             *handlers.SSEHandler
	Templates       *handlers.TemplatesHandler
	BrandAssets     *handlers.BrandAssetsHandler
	Uploads         *handlers.UploadsHandler
	Recommendations *handlers.RecommendationsHandler
	Render          *handlers.RenderHandler
	Composite       *handlers.CompositeHandler
	AgentSettings   *handlers.AgentSettingsHandler
	// Files is only set when running against the local-disk storage backend
	// (no Docker/MinIO available) — it serves objects back over plain HTTP.
	Files *handlers.FilesHandler
}

func NewRouter(d Deps) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recover(d.Logger))
	r.Use(middleware.Logging(d.Logger))
	r.Use(middleware.CORS(d.CORSOrigin))

	r.Get("/healthz", d.Health.Healthz)
	r.Get("/readyz", d.Health.Readyz)
	if d.Files != nil {
		r.Get("/files/*", d.Files.Serve)
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", d.Auth.Login)
		r.Post("/auth/logout", d.Auth.Logout)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(d.AuthManager))

			r.Get("/auth/me", d.Auth.Me)

			r.Get("/templates", d.Templates.List)
			r.Get("/templates/{id}", d.Templates.Get)

			r.Get("/brand-assets", d.BrandAssets.List)

			r.Post("/uploads", d.Uploads.Create)
			r.Post("/recommendations", d.Recommendations.Recommend)

			r.Post("/projects", d.Projects.Create)
			r.Get("/projects", d.Projects.List)
			r.Get("/projects/{id}", d.Projects.Get)
			r.Get("/projects/{id}/events", d.SSE.Stream)
			r.Get("/projects/{id}/download", d.Projects.Download)

			r.Post("/projects/{id}/brief/rerun", d.PipelineStages.RerunBrief)
			r.Post("/projects/{id}/concepts/{n}/select", d.PipelineStages.SelectConcept)
			r.Patch("/projects/{id}/copy", d.PipelineStages.PatchCopy)
			r.Patch("/projects/{id}/art-direction", d.PipelineStages.PatchArtDirection)
			r.Post("/projects/{id}/art-direction/rerun", d.PipelineStages.RerunArtDirection)

			r.Post("/projects/{id}/render", d.Render.Render)
			r.Post("/projects/{id}/composite", d.Composite.Composite)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin)
				r.Post("/templates", d.Templates.Create)
				r.Patch("/templates/{id}", d.Templates.Patch)
				r.Post("/templates/import-starter-library", d.Templates.ImportStarterLibrary)
				r.Put("/brand-assets", d.BrandAssets.Put)
				r.Get("/agent-settings", d.AgentSettings.Get)
				r.Put("/agent-settings", d.AgentSettings.Put)
				r.Put("/agent-settings/keys", d.AgentSettings.SetAPIKey)
			})
		})
	})

	return r
}
