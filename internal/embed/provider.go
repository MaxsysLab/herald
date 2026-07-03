package embed

import (
	"fmt"
	"os"
)

// ProviderConfig selects an embedder. The zero value (or Provider "fake") resolves to
// the dependency-free FakeEmbedder — so the product runs offline out of the box, with
// a real hosted embedder available whenever a client configures one.
type ProviderConfig struct {
	Provider string `json:"provider"` // "fake" (default), "openai", "voyage", "custom"
	BaseURL  string `json:"base_url"` // required for "custom"
	Model    string `json:"model"`
	APIKey   string `json:"api_key"` // if empty, read from the provider's env var
}

type embedDefault struct {
	baseURL string
	model   string
	envKey  string
}

var embedDefaults = map[string]embedDefault{
	"openai": {"https://api.openai.com/v1", "text-embedding-3-small", "OPENAI_API_KEY"},
	"voyage": {"https://api.voyageai.com/v1", "voyage-3-lite", "VOYAGE_API_KEY"},
}

// Resolve builds an Embedder from a ProviderConfig, applying defaults and reading the
// API key from the environment when not supplied.
func Resolve(cfg ProviderConfig) (Embedder, error) {
	switch cfg.Provider {
	case "", "fake":
		return FakeEmbedder{}, nil
	}
	def, known := embedDefaults[cfg.Provider]
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = def.baseURL
	}
	model := cfg.Model
	if model == "" {
		model = def.model
	}
	if baseURL == "" {
		return nil, fmt.Errorf("embed: provider %q needs a base URL", cfg.Provider)
	}
	apiKey := cfg.APIKey
	if apiKey == "" && known && def.envKey != "" {
		apiKey = os.Getenv(def.envKey)
	}
	return NewOpenAICompatEmbedder(cfg.Provider, baseURL, apiKey, model), nil
}
