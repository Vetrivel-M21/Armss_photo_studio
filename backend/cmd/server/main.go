// Command server boots the ARMSS Poster Studio API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"armss-poster-studio/internal/agents"
	"armss-poster-studio/internal/auth"
	"armss-poster-studio/internal/compositor"
	"armss-poster-studio/internal/concurrency"
	"armss-poster-studio/internal/config"
	apihttp "armss-poster-studio/internal/http"
	"armss-poster-studio/internal/http/handlers"
	"armss-poster-studio/internal/imagegen"
	"armss-poster-studio/internal/imagegen/flux"
	"armss-poster-studio/internal/imagegen/ideogram"
	"armss-poster-studio/internal/imagegen/openai"
	"armss-poster-studio/internal/observability"
	"armss-poster-studio/internal/pipeline"
	"armss-poster-studio/internal/ratelimit"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
)

func main() {
	// Loads .env into the process environment for local dev if present;
	// silently a no-op in production where real env vars are set directly.
	_ = godotenv.Load()

	logger := observability.NewLogger()
	slog.SetDefault(logger)

	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		logger.Error("config_load_failed", "error", err)
		os.Exit(1)
	}

	if err := mysql.Migrate(cfg.MySQLDSN); err != nil {
		logger.Error("migration_failed", "error", err)
		os.Exit(1)
	}

	db, err := mysql.Open(cfg.MySQLDSN)
	if err != nil {
		logger.Error("db_connect_failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	var objectStore *storage.Store
	if cfg.StorageDriver == "s3" {
		objectStore, err = storage.New(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3UseSSL)
	} else {
		objectStore, err = storage.NewLocal(cfg.LocalStorageDir, cfg.PublicBaseURL)
	}
	if err != nil {
		logger.Error("object_store_init_failed", "error", err)
		os.Exit(1)
	}
	bootCtx, bootCancel := context.WithTimeout(context.Background(), 15*time.Second)
	if err := objectStore.EnsureBucket(bootCtx); err != nil {
		logger.Error("bucket_ensure_failed", "error", err)
		os.Exit(1)
	}
	bootCancel()

	usersRepo := mysql.NewUsersRepo(db)
	sessionsRepo := mysql.NewSessionsRepo(db)
	authManager := auth.NewManager(usersRepo, sessionsRepo, cfg.SessionCookie, cfg.SessionTTL, cfg.Env == "production")

	projectsRepo := mysql.NewProjectsRepo(db)
	projectRefsRepo := mysql.NewProjectReferencesRepo(db)
	stageOutputsRepo := mysql.NewStageOutputsRepo(db)
	rendersRepo := mysql.NewRendersRepo(db)
	templatesRepo := mysql.NewTemplatesRepo(db)
	uploadsRepo := mysql.NewUploadsRepo(db)
	userLimitsRepo := mysql.NewUserLimitsRepo(db)
	rateRecordsRepo := mysql.NewRateRecordsRepo(db)
	rateLimiter := ratelimit.NewLimiter(userLimitsRepo, rateRecordsRepo, logger)

	backendCfg := agents.BackendConfig{
		AnthropicAPIKey:  cfg.AnthropicAPIKey,
		AnthropicModel:   cfg.AnthropicModel,
		AnthropicBaseURL: cfg.AnthropicBaseURL,
		GeminiAPIKey:     cfg.GeminiAPIKey,
		GeminiModel:      cfg.GeminiModel,
		GeminiBaseURL:    cfg.GeminiBaseURL,
		Timeout:          cfg.AgentTimeout,
	}
	agentsClient, err := agents.NewFromProvider(cfg.AgentProvider, backendCfg)
	if err != nil {
		logger.Error("agent_provider_init_failed", "error", err)
		os.Exit(1)
	}

	agentSettingsRepo := mysql.NewAgentSettingsRepo(db)
	effectiveOpenAIKey := cfg.OpenAIAPIKey
	effectiveFluxKey := cfg.FluxAPIKey
	effectiveIdeogramKey := cfg.IdeogramAPIKey
	defaultImageProvider := cfg.DefaultImageProvider
	if defaultImageProvider == "" {
		defaultImageProvider = "flux"
	}

	// An admin's saved provider/model/API-key overrides (set via the
	// agent-settings UI) take effect immediately on boot, on top of the
	// env-configured defaults — lets switching providers or rotating a key
	// survive a restart.
	if saved, err := agentSettingsRepo.Get(context.Background()); err == nil {
		if saved.AnthropicAPIKey.Valid && saved.AnthropicAPIKey.String != "" {
			backendCfg.AnthropicAPIKey = saved.AnthropicAPIKey.String
		}
		if saved.GeminiAPIKey.Valid && saved.GeminiAPIKey.String != "" {
			backendCfg.GeminiAPIKey = saved.GeminiAPIKey.String
		}
		if saved.OpenAIAPIKey.Valid && saved.OpenAIAPIKey.String != "" {
			effectiveOpenAIKey = saved.OpenAIAPIKey.String
		}
		if saved.FluxAPIKey.Valid && saved.FluxAPIKey.String != "" {
			effectiveFluxKey = saved.FluxAPIKey.String
		}
		if saved.IdeogramAPIKey.Valid && saved.IdeogramAPIKey.String != "" {
			effectiveIdeogramKey = saved.IdeogramAPIKey.String
		}
		if saved.DefaultImageProvider.Valid && saved.DefaultImageProvider.String != "" {
			defaultImageProvider = saved.DefaultImageProvider.String
		}
		backend, resolvedProvider, err := agents.BuildBackend(saved.Provider, backendCfg.WithModel(saved.Provider, saved.Model))
		if err != nil {
			logger.Error("saved_agent_settings_invalid", "error", err)
		} else {
			agentsClient.SwapBackend(resolvedProvider, backend)
		}
	} else if !errors.Is(err, mysql.ErrNotFound) {
		logger.Error("agent_settings_load_failed", "error", err)
	}

	concurrencyLocks := concurrency.NewRegistry()
	bus := pipeline.NewBus()
	brandAssetsRepo := mysql.NewBrandAssetsRepo(db)
	orchestrator := &pipeline.Orchestrator{
		Agents:       agentsClient,
		Store:        objectStore,
		Projects:     projectsRepo,
		StageOutputs: stageOutputsRepo,
		ProjectRefs:  projectRefsRepo,
		Templates:    templatesRepo,
		Uploads:      uploadsRepo,
		BrandAssets:  brandAssetsRepo,
		Bus:                  bus,
		Logger:               logger,
		Locks:                concurrencyLocks,
		DefaultImageProvider: defaultImageProvider,
	}

	idempotencyRepo := mysql.NewIdempotencyRepo(db)

	openaiClient := openai.NewClient(effectiveOpenAIKey, cfg.OpenAIBaseURL, cfg.ProviderTimeout, cfg.OpenAIImageCostCents)
	fluxClient := flux.NewClient(effectiveFluxKey, cfg.FluxBaseURL, cfg.FluxModel, cfg.ProviderTimeout, cfg.FluxCostCents)
	ideogramClient := ideogram.NewClient(effectiveIdeogramKey, cfg.IdeogramBaseURL, cfg.IdeogramModel, cfg.ProviderTimeout, cfg.IdeogramCostCents)
	imageRegistry := imagegen.NewRegistry(defaultImageProvider, openaiClient, fluxClient, ideogramClient)

	comp, err := compositor.New(compositor.Config{
		NavyHex:       cfg.BrandNavy,
		ContactString: cfg.BrandContactString,
		FontPath:      cfg.BrandFontPath,
	})
	if err != nil {
		logger.Error("compositor_init_failed", "error", err)
		os.Exit(1)
	}
	assetLoader := compositor.NewAssetLoader(brandAssetsRepo, objectStore)

	var filesHandler *handlers.FilesHandler
	if cfg.StorageDriver != "s3" {
		filesHandler = handlers.NewFilesHandler(cfg.LocalStorageDir)
	}

	router := apihttp.NewRouter(apihttp.Deps{
		Logger:          logger,
		AuthManager:     authManager,
		CORSOrigin:      "http://localhost:5173",
		Health:          handlers.NewHealthHandler(db, objectStore),
		Auth:            handlers.NewAuthHandler(authManager),
		Projects:        handlers.NewProjectsHandler(orchestrator, projectsRepo, stageOutputsRepo, rendersRepo, objectStore, rateLimiter),
		PipelineStages:  handlers.NewPipelineStagesHandler(orchestrator, projectsRepo, stageOutputsRepo, rendersRepo, objectStore, rateLimiter),
		SSE:             handlers.NewSSEHandler(bus),
		Templates:       handlers.NewTemplatesHandler(templatesRepo, objectStore, agentsClient),
		BrandAssets:     handlers.NewBrandAssetsHandler(brandAssetsRepo, objectStore, assetLoader.Invalidate),
		Uploads:         handlers.NewUploadsHandler(uploadsRepo, objectStore),
		Recommendations: handlers.NewRecommendationsHandler(templatesRepo, objectStore, agentsClient),
		Render:          handlers.NewRenderHandler(imageRegistry, projectsRepo, stageOutputsRepo, rendersRepo, idempotencyRepo, objectStore, bus, concurrencyLocks, rateLimiter),
		Composite:       handlers.NewCompositeHandler(comp, assetLoader, projectsRepo, stageOutputsRepo, rendersRepo, idempotencyRepo, objectStore, bus, concurrencyLocks),
		Usage:           handlers.NewUsageHandler(userLimitsRepo, rateRecordsRepo),
		AgentSettings: handlers.NewAgentSettingsHandler(handlers.AgentSettingsHandlerOpts{
			Client:            agentsClient,
			Repo:              agentSettingsRepo,
			Config:            backendCfg,
			Registry:          imageRegistry,
			OpenAIKey:         effectiveOpenAIKey,
			OpenAIBaseURL:     cfg.OpenAIBaseURL,
			OpenAITimeout:     cfg.ProviderTimeout,
			OpenAICostCents:   cfg.OpenAIImageCostCents,
			FluxKey:           effectiveFluxKey,
			FluxBaseURL:       cfg.FluxBaseURL,
			FluxModel:         cfg.FluxModel,
			FluxCostCents:     cfg.FluxCostCents,
			IdeogramKey:       effectiveIdeogramKey,
			IdeogramBaseURL:   cfg.IdeogramBaseURL,
			IdeogramModel:     cfg.IdeogramModel,
			IdeogramCostCents: cfg.IdeogramCostCents,
		}),
		Files: filesHandler,
	})

	srv := apihttp.NewServer(cfg.Port, router)

	go func() {
		logger.Info("server_starting", "port", cfg.Port, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server_failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("server_stopping")
	if err := apihttp.Shutdown(srv, 15*time.Second); err != nil {
		logger.Error("server_shutdown_failed", "error", err)
	}
}
