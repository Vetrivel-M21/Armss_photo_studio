package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// AgentSettings is the single-row admin-controlled override for which LLM
// vendor/model answers the agent pipeline — lets an admin switch providers
// from the UI (e.g. when one vendor's quota runs out) without editing env
// vars or restarting the server. The three API key columns are a separate,
// independent override: an admin can paste a replacement key for any
// provider without necessarily switching the active one.
type AgentSettings struct {
	Provider             string         `db:"provider"`
	Model                string         `db:"model"`
	DefaultImageProvider sql.NullString `db:"default_image_provider"`
	AnthropicAPIKey      sql.NullString `db:"anthropic_api_key"`
	GeminiAPIKey         sql.NullString `db:"gemini_api_key"`
	OpenAIAPIKey         sql.NullString `db:"openai_api_key"`
	FluxAPIKey           sql.NullString `db:"flux_api_key"`
	IdeogramAPIKey       sql.NullString `db:"ideogram_api_key"`
}

type AgentSettingsRepo struct {
	db *sqlx.DB
}

func NewAgentSettingsRepo(db *sqlx.DB) *AgentSettingsRepo {
	return &AgentSettingsRepo{db: db}
}

// Get returns the saved override, or ErrNotFound if the admin has never set
// one (callers should fall back to the env-configured default provider).
func (r *AgentSettingsRepo) Get(ctx context.Context) (*AgentSettings, error) {
	var s AgentSettings
	err := r.db.GetContext(ctx, &s,
		`SELECT provider, model, default_image_provider, anthropic_api_key, gemini_api_key, openai_api_key, flux_api_key, ideogram_api_key FROM agent_settings WHERE id = 1`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get agent settings: %w", err)
	}
	return &s, nil
}

// Set upserts the provider/model row, leaving any previously-saved API keys
// untouched (a provider switch and a key update are independent actions).
func (r *AgentSettingsRepo) Set(ctx context.Context, provider, model string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agent_settings (id, provider, model) VALUES (1, ?, ?)
		 ON DUPLICATE KEY UPDATE provider = VALUES(provider), model = VALUES(model)`,
		provider, model)
	if err != nil {
		return fmt.Errorf("set agent settings: %w", err)
	}
	return nil
}

// SetAnthropicKey upserts just the Anthropic API key, defaulting provider/model
// to the given values only if this is the first time the row is created.
func (r *AgentSettingsRepo) SetAnthropicKey(ctx context.Context, defaultProvider, defaultModel, key string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agent_settings (id, provider, model, anthropic_api_key) VALUES (1, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE anthropic_api_key = VALUES(anthropic_api_key)`,
		defaultProvider, defaultModel, key)
	if err != nil {
		return fmt.Errorf("set anthropic api key: %w", err)
	}
	return nil
}

// SetGeminiKey upserts just the Gemini API key (see SetAnthropicKey).
func (r *AgentSettingsRepo) SetGeminiKey(ctx context.Context, defaultProvider, defaultModel, key string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agent_settings (id, provider, model, gemini_api_key) VALUES (1, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE gemini_api_key = VALUES(gemini_api_key)`,
		defaultProvider, defaultModel, key)
	if err != nil {
		return fmt.Errorf("set gemini api key: %w", err)
	}
	return nil
}

// SetOpenAIKey upserts just the OpenAI (image generation) API key — this key
// is unrelated to the agents.Client provider/model above; it feeds the
// separate imagegen.Registry.
func (r *AgentSettingsRepo) SetOpenAIKey(ctx context.Context, defaultProvider, defaultModel, key string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agent_settings (id, provider, model, openai_api_key) VALUES (1, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE openai_api_key = VALUES(openai_api_key)`,
		defaultProvider, defaultModel, key)
	if err != nil {
		return fmt.Errorf("set openai api key: %w", err)
	}
	return nil
}

// SetFluxKey upserts the Flux API key.
func (r *AgentSettingsRepo) SetFluxKey(ctx context.Context, defaultProvider, defaultModel, key string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agent_settings (id, provider, model, flux_api_key) VALUES (1, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE flux_api_key = VALUES(flux_api_key)`,
		defaultProvider, defaultModel, key)
	if err != nil {
		return fmt.Errorf("set flux api key: %w", err)
	}
	return nil
}

// SetIdeogramKey upserts the Ideogram API key.
func (r *AgentSettingsRepo) SetIdeogramKey(ctx context.Context, defaultProvider, defaultModel, key string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agent_settings (id, provider, model, ideogram_api_key) VALUES (1, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE ideogram_api_key = VALUES(ideogram_api_key)`,
		defaultProvider, defaultModel, key)
	if err != nil {
		return fmt.Errorf("set ideogram api key: %w", err)
	}
	return nil
}

// SetDefaultImageProvider upserts the default image provider selection ("flux" | "ideogram" | "openai").
func (r *AgentSettingsRepo) SetDefaultImageProvider(ctx context.Context, defaultProvider, defaultModel, imageProvider string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agent_settings (id, provider, model, default_image_provider) VALUES (1, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE default_image_provider = VALUES(default_image_provider)`,
		defaultProvider, defaultModel, imageProvider)
	if err != nil {
		return fmt.Errorf("set default image provider: %w", err)
	}
	return nil
}
