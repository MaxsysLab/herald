package llm

import (
	"fmt"
	"os"
)

// ProviderConfig selects and configures an LLM backend. The zero value resolves to
// the Groq default. Provider is user-facing: a client can point Herald at OpenAI,
// Grok, a local Ollama, or their own gateway without a code change.
type ProviderConfig struct {
	Provider string // "groq" (default), "openai", "grok", "ollama", "custom"
	BaseURL  string // required for "custom"; defaults filled in for known providers
	Model    string // defaults filled in for known providers
	APIKey   string // if empty, read from the provider's conventional env var
}

type providerDefault struct {
	baseURL string
	model   string
	envKey  string
}

// Groq is the deliberate default: fast, cheap, and provider-neutral. Model ids and
// endpoints below are defaults only — override any of them via ProviderConfig.
var providerDefaults = map[string]providerDefault{
	"groq":   {"https://api.groq.com/openai/v1", "llama-3.3-70b-versatile", "GROQ_API_KEY"},
	"openai": {"https://api.openai.com/v1", "gpt-4o-mini", "OPENAI_API_KEY"},
	"grok":   {"https://api.x.ai/v1", "grok-2-latest", "XAI_API_KEY"},
	"ollama": {"http://localhost:11434/v1", "llama3.1", ""},
}

// DefaultProvider is the pre-selected, unbiased default.
const DefaultProvider = "groq"

// Resolve builds a Client from a ProviderConfig, applying defaults and reading the
// API key from the environment when not supplied.
//
// Claude (Anthropic) uses a different wire format than the OpenAI-compatible shape
// used here; wiring it is a documented adapter seam and returns an explicit error
// for now so callers fail loudly rather than silently mis-routing.
func Resolve(cfg ProviderConfig) (Client, error) {
	name := cfg.Provider
	if name == "" {
		name = DefaultProvider
	}

	if name == "claude" || name == "anthropic" {
		return nil, fmt.Errorf("llm: provider %q not wired yet (Anthropic uses a non-OpenAI wire format — adapter seam)", name)
	}

	def, known := providerDefaults[name]
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = def.baseURL
	}
	model := cfg.Model
	if model == "" {
		model = def.model
	}
	if baseURL == "" {
		return nil, fmt.Errorf("llm: provider %q needs a base URL", name)
	}

	apiKey := cfg.APIKey
	if apiKey == "" && known && def.envKey != "" {
		apiKey = os.Getenv(def.envKey)
	}

	return NewOpenAICompatClient(name, baseURL, apiKey, model), nil
}
