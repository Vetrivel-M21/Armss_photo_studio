package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"armss-poster-studio/internal/agents"
	"armss-poster-studio/internal/http/dto"
	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/imagegen"
	"armss-poster-studio/internal/imagegen/flux"
	"armss-poster-studio/internal/imagegen/ideogram"
	"armss-poster-studio/internal/imagegen/openai"
	"armss-poster-studio/internal/store/mysql"
)

// AgentSettingsHandler lets an admin switch the agent pipeline's LLM
// provider/model or the default image generation provider at runtime,
// and separately paste in replacement API keys for any provider
// (Anthropic, Gemini, OpenAI, Flux, Ideogram) — all without editing .env
// or restarting the server. A saved key is never echoed back.
type AgentSettingsHandler struct {
	Client   *agents.Client
	Repo     *mysql.AgentSettingsRepo
	Registry *imagegen.Registry

	OpenAIBaseURL   string
	OpenAITimeout   time.Duration
	OpenAICostCents int

	FluxBaseURL   string
	FluxModel     string
	FluxCostCents int

	IdeogramBaseURL   string
	IdeogramModel     string
	IdeogramCostCents int

	mu          sync.Mutex
	Config      agents.BackendConfig
	openaiKey   string
	fluxKey     string
	ideogramKey string
}

type AgentSettingsHandlerOpts struct {
	Client          *agents.Client
	Repo            *mysql.AgentSettingsRepo
	Config          agents.BackendConfig
	Registry        *imagegen.Registry
	OpenAIKey       string
	OpenAIBaseURL   string
	OpenAITimeout   time.Duration
	OpenAICostCents int
	FluxKey         string
	FluxBaseURL     string
	FluxModel       string
	FluxCostCents   int
	IdeogramKey     string
	IdeogramBaseURL string
	IdeogramModel   string
	IdeogramCostCents int
}

func NewAgentSettingsHandler(opts AgentSettingsHandlerOpts) *AgentSettingsHandler {
	return &AgentSettingsHandler{
		Client:            opts.Client,
		Repo:              opts.Repo,
		Config:            opts.Config,
		Registry:          opts.Registry,
		openaiKey:         opts.OpenAIKey,
		OpenAIBaseURL:     opts.OpenAIBaseURL,
		OpenAITimeout:     opts.OpenAITimeout,
		OpenAICostCents:   opts.OpenAICostCents,
		fluxKey:           opts.FluxKey,
		FluxBaseURL:       opts.FluxBaseURL,
		FluxModel:         opts.FluxModel,
		FluxCostCents:     opts.FluxCostCents,
		ideogramKey:       opts.IdeogramKey,
		IdeogramBaseURL:   opts.IdeogramBaseURL,
		IdeogramModel:     opts.IdeogramModel,
		IdeogramCostCents: opts.IdeogramCostCents,
	}
}

func (h *AgentSettingsHandler) snapshot() dto.AgentSettingsResponse {
	h.mu.Lock()
	defer h.mu.Unlock()
	return dto.AgentSettingsResponse{
		Provider:            h.Client.Provider(),
		Model:               h.Client.ModelName(),
		AnthropicConfigured: h.Config.AnthropicAPIKey != "",
		GeminiConfigured:    h.Config.GeminiAPIKey != "",
		OpenAIConfigured:    h.openaiKey != "",
		FluxConfigured:      h.fluxKey != "",
		IdeogramConfigured:  h.ideogramKey != "",
		ImageProvider:       h.Registry.DefaultName(),
	}
}

func (h *AgentSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, h.snapshot())
}

func (h *AgentSettingsHandler) Put(w http.ResponseWriter, r *http.Request) {
	var req dto.AgentSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Request body must be valid JSON.")
		return
	}

	defaultProvider, defaultModel := h.Client.Provider(), h.Client.ModelName()

	// Update Image Provider if requested
	if req.ImageProvider != "" {
		if req.ImageProvider != "flux" && req.ImageProvider != "ideogram" && req.ImageProvider != "openai" {
			respond.Error(w, http.StatusBadRequest, "imageProvider must be \"flux\", \"ideogram\", or \"openai\".")
			return
		}
		if err := h.Repo.SetDefaultImageProvider(r.Context(), defaultProvider, defaultModel, req.ImageProvider); err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not save the image provider setting.")
			return
		}
		h.Registry.SetDefault(req.ImageProvider)
	}

	// Update LLM Agent Provider/Model if requested
	if req.Provider != "" || req.Model != "" {
		if req.Provider != "anthropic" && req.Provider != "gemini" {
			respond.Error(w, http.StatusBadRequest, "provider must be \"anthropic\" or \"gemini\".")
			return
		}
		if req.Model == "" {
			respond.Error(w, http.StatusBadRequest, "model is required when updating provider.")
			return
		}

		h.mu.Lock()
		cfg := h.Config.WithModel(req.Provider, req.Model)
		h.mu.Unlock()

		backend, resolvedProvider, err := agents.BuildBackend(req.Provider, cfg)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := h.Repo.Set(r.Context(), req.Provider, req.Model); err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not save the setting right now.")
			return
		}
		h.Client.SwapBackend(resolvedProvider, backend)
	}

	respond.JSON(w, http.StatusOK, h.snapshot())
}

// SetAPIKey saves a replacement API key for one provider.
func (h *AgentSettingsHandler) SetAPIKey(w http.ResponseWriter, r *http.Request) {
	var req dto.SetAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Request body must be valid JSON with \"provider\" and \"apiKey\".")
		return
	}
	if req.APIKey == "" {
		respond.Error(w, http.StatusBadRequest, "apiKey is required.")
		return
	}

	defaultProvider, defaultModel := h.Client.Provider(), h.Client.ModelName()

	switch req.Provider {
	case "anthropic", "gemini":
		h.mu.Lock()
		h.Config = h.Config.WithKey(req.Provider, req.APIKey)
		cfg := h.Config
		h.mu.Unlock()

		var saveErr error
		if req.Provider == "anthropic" {
			saveErr = h.Repo.SetAnthropicKey(r.Context(), defaultProvider, defaultModel, req.APIKey)
		} else {
			saveErr = h.Repo.SetGeminiKey(r.Context(), defaultProvider, defaultModel, req.APIKey)
		}
		if saveErr != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not save this key right now.")
			return
		}

		if h.Client.Provider() == req.Provider {
			backend, resolvedProvider, err := agents.BuildBackend(req.Provider, cfg)
			if err != nil {
				respond.Error(w, http.StatusBadRequest, err.Error())
				return
			}
			h.Client.SwapBackend(resolvedProvider, backend)
		}

	case "openai":
		h.mu.Lock()
		h.openaiKey = req.APIKey
		h.mu.Unlock()

		if err := h.Repo.SetOpenAIKey(r.Context(), defaultProvider, defaultModel, req.APIKey); err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not save this key right now.")
			return
		}
		h.Registry.Swap("openai", openai.NewClient(req.APIKey, h.OpenAIBaseURL, h.OpenAITimeout, h.OpenAICostCents))

	case "flux":
		h.mu.Lock()
		h.fluxKey = req.APIKey
		h.mu.Unlock()

		if err := h.Repo.SetFluxKey(r.Context(), defaultProvider, defaultModel, req.APIKey); err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not save this key right now.")
			return
		}
		h.Registry.Swap("flux", flux.NewClient(req.APIKey, h.FluxBaseURL, h.FluxModel, h.OpenAITimeout, h.FluxCostCents))

	case "ideogram":
		h.mu.Lock()
		h.ideogramKey = req.APIKey
		h.mu.Unlock()

		if err := h.Repo.SetIdeogramKey(r.Context(), defaultProvider, defaultModel, req.APIKey); err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not save this key right now.")
			return
		}
		h.Registry.Swap("ideogram", ideogram.NewClient(req.APIKey, h.IdeogramBaseURL, h.IdeogramModel, h.OpenAITimeout, h.IdeogramCostCents))

	default:
		respond.Error(w, http.StatusBadRequest, "provider must be \"anthropic\", \"gemini\", \"openai\", \"flux\", or \"ideogram\".")
		return
	}

	respond.JSON(w, http.StatusOK, h.snapshot())
}
