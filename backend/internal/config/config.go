// Package config loads all server configuration from environment variables.
package config

import (
	"fmt"
	"strconv"
	"time"
)

type Config struct {
	Port string
	Env  string

	MySQLDSN string

	// StorageDriver selects the object store: "local" (default — plain disk,
	// no Docker needed, files served back via GET /files/*) or "s3" (MinIO/S3,
	// per spec §11, run via docker-compose).
	StorageDriver   string
	LocalStorageDir string
	PublicBaseURL   string

	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool

	SessionSecret string
	SessionTTL    time.Duration
	SessionCookie string

	// AgentProvider selects which LLM answers the 7-agent pipeline:
	// "anthropic" (default, spec §5) or "gemini". Config-driven so it can be
	// changed without a deploy.
	AgentProvider string

	AnthropicAPIKey  string
	AnthropicModel   string
	AnthropicBaseURL string

	GeminiAPIKey  string
	GeminiModel   string
	GeminiBaseURL string

	DefaultImageProvider string

	OpenAIAPIKey         string
	OpenAIBaseURL        string
	OpenAIImageCostCents int

	FluxAPIKey     string
	FluxBaseURL    string
	FluxModel      string
	FluxCostCents  int

	IdeogramAPIKey    string
	IdeogramBaseURL   string
	IdeogramModel     string
	IdeogramCostCents int

	BrandNavy          string
	BrandRed           string
	BrandSaffron       string
	BrandCream         string
	BrandContactString string
	BrandFontPath      string

	AgentTimeout    time.Duration
	ProviderTimeout time.Duration
}

// Load reads configuration from the given environment lookup function (normally os.LookupEnv).
// Only MYSQL_DSN, S3_* and SESSION_SECRET are required; AI provider keys may be blank —
// callers must handle that as "feature disabled", never as a startup error.
func Load(getenv func(string) (string, bool)) (*Config, error) {
	get := func(key, def string) string {
		if v, ok := getenv(key); ok && v != "" {
			return v
		}
		return def
	}
	require := func(key string) (string, error) {
		v, ok := getenv(key)
		if !ok || v == "" {
			return "", fmt.Errorf("required env var %s is not set", key)
		}
		return v, nil
	}

	mysqlDSN, err := require("MYSQL_DSN")
	if err != nil {
		return nil, err
	}
	sessionSecret, err := require("SESSION_SECRET")
	if err != nil {
		return nil, err
	}

	storageDriver := get("STORAGE_DRIVER", "local")
	var s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket string
	if storageDriver == "s3" {
		if s3Endpoint, err = require("S3_ENDPOINT"); err != nil {
			return nil, err
		}
		if s3AccessKey, err = require("S3_ACCESS_KEY"); err != nil {
			return nil, err
		}
		if s3SecretKey, err = require("S3_SECRET_KEY"); err != nil {
			return nil, err
		}
		if s3Bucket, err = require("S3_BUCKET"); err != nil {
			return nil, err
		}
	}

	sessionTTLHours, err := strconv.Atoi(get("SESSION_TTL_HOURS", "168"))
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_TTL_HOURS: %w", err)
	}
	imageCostCents, err := strconv.Atoi(get("OPENAI_IMAGE_COST_CENTS", "8"))
	if err != nil {
		return nil, fmt.Errorf("invalid OPENAI_IMAGE_COST_CENTS: %w", err)
	}
	fluxCostCents, err := strconv.Atoi(get("FLUX_IMAGE_COST_CENTS", "3"))
	if err != nil {
		return nil, fmt.Errorf("invalid FLUX_IMAGE_COST_CENTS: %w", err)
	}
	ideogramCostCents, err := strconv.Atoi(get("IDEOGRAM_IMAGE_COST_CENTS", "8"))
	if err != nil {
		return nil, fmt.Errorf("invalid IDEOGRAM_IMAGE_COST_CENTS: %w", err)
	}
	s3UseSSL, err := strconv.ParseBool(get("S3_USE_SSL", "false"))
	if err != nil {
		return nil, fmt.Errorf("invalid S3_USE_SSL: %w", err)
	}

	port := get("PORT", "8080")

	return &Config{
		Port: port,
		Env:  get("ENV", "development"),

		MySQLDSN: mysqlDSN,

		StorageDriver:   storageDriver,
		LocalStorageDir: get("LOCAL_STORAGE_DIR", "data/storage"),
		PublicBaseURL:   get("PUBLIC_BASE_URL", "http://localhost:"+port),

		S3Endpoint:  s3Endpoint,
		S3AccessKey: s3AccessKey,
		S3SecretKey: s3SecretKey,
		S3Bucket:    s3Bucket,
		S3UseSSL:    s3UseSSL,

		SessionSecret: sessionSecret,
		SessionTTL:    time.Duration(sessionTTLHours) * time.Hour,
		SessionCookie: "armss_session",

		// AI provider keys are intentionally allowed to be blank — see agents.Client
		// and imagegen.Registry, which switch to a "disabled" mode instead of erroring.
		AgentProvider: get("AGENT_PROVIDER", "anthropic"),

		AnthropicAPIKey:  get("ANTHROPIC_API_KEY", ""),
		AnthropicModel:   get("ANTHROPIC_MODEL", "claude-sonnet-5"),
		AnthropicBaseURL: get("ANTHROPIC_BASE_URL", "https://api.anthropic.com"),

		GeminiAPIKey:  get("GEMINI_API_KEY", ""),
		GeminiModel:   get("GEMINI_MODEL", "gemini-3.6-flash"),
		GeminiBaseURL: get("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com"),

		DefaultImageProvider: get("DEFAULT_IMAGE_PROVIDER", "flux"),

		OpenAIAPIKey:         get("OPENAI_API_KEY", ""),
		OpenAIBaseURL:        get("OPENAI_BASE_URL", "https://api.openai.com"),
		OpenAIImageCostCents: imageCostCents,

		FluxAPIKey:     get("FLUX_API_KEY", ""),
		FluxBaseURL:    get("FLUX_BASE_URL", "https://api.together.xyz"),
		FluxModel:      get("FLUX_MODEL", "black-forest-labs/FLUX.1-schnell"),
		FluxCostCents:  fluxCostCents,

		IdeogramAPIKey:    get("IDEOGRAM_API_KEY", ""),
		IdeogramBaseURL:   get("IDEOGRAM_BASE_URL", "https://api.ideogram.ai"),
		IdeogramModel:     get("IDEOGRAM_MODEL", "V_2"),
		IdeogramCostCents: ideogramCostCents,

		BrandNavy:          get("BRAND_NAVY", "#16265A"),
		BrandRed:           get("BRAND_RED", "#D32027"),
		BrandSaffron:       get("BRAND_SAFFRON", "#F5A623"),
		BrandCream:         get("BRAND_CREAM", "#FFF8EE"),
		BrandContactString: get("BRAND_CONTACT_STRING", "ARMSS Employment Solutions"),
		BrandFontPath:      get("BRAND_FONT_PATH", "assets/fonts/NotoSansTamil-Bold.ttf"),

		AgentTimeout:    60 * time.Second,
		ProviderTimeout: 120 * time.Second,
	}, nil
}
