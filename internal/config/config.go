// Package config persists one business's Herald setup — the structured profile plus
// which LLM provider to use and how conservative to be about grounding. It is stored
// as plain JSON so a business's whole configuration is a single reviewable file.
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MaxsysLab/herald/internal/embed"
	"github.com/MaxsysLab/herald/internal/knowledge"
	"github.com/MaxsysLab/herald/internal/llm"
)

// ProjectConfig is the complete, portable configuration for one deployed officer:
// the business identity, its knowledge, the chosen providers, and grounding strictness.
type ProjectConfig struct {
	Name      string                    `json:"name"`
	Profile   knowledge.BusinessProfile `json:"profile"`
	Knowledge []knowledge.Passage       `json:"knowledge"` // KB source text (vectors recomputed on load)
	LLM       llm.ProviderConfig        `json:"llm"`
	Embed     embed.ProviderConfig      `json:"embed"`
	MinScore  float64                   `json:"min_score"` // grounding threshold; 0 → officer default
}

// Load reads a ProjectConfig from a JSON file.
func Load(path string) (*ProjectConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	var c ProjectConfig
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	return &c, nil
}

// Save writes the config to a JSON file (0600 — it may name a provider, though never
// a secret: API keys come from the environment, not this file).
func (c *ProjectConfig) Save(path string) error {
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("config: write %s: %w", path, err)
	}
	return nil
}
