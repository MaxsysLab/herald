package config

import (
	"path/filepath"
	"testing"

	"github.com/MaxsysLab/herald/internal/knowledge"
	"github.com/MaxsysLab/herald/internal/llm"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	cfg := &ProjectConfig{
		Name: "bloom-cafe",
		Profile: knowledge.BusinessProfile{
			Name: "Bloom Café",
			Services: []knowledge.Service{
				{Name: "Cappuccino", Price: &knowledge.Money{Amount: "4.50", Currency: "USD"}},
			},
		},
		LLM:      llm.ProviderConfig{Provider: "groq", Model: "llama-3.3-70b-versatile"},
		MinScore: 0.4,
	}

	path := filepath.Join(t.TempDir(), "project.json")
	if err := cfg.Save(path); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got.Name != cfg.Name || got.MinScore != cfg.MinScore {
		t.Fatalf("scalar mismatch: %+v", got)
	}
	if got.LLM.Provider != "groq" {
		t.Fatalf("llm provider not preserved: %+v", got.LLM)
	}
	if len(got.Profile.Services) != 1 || got.Profile.Services[0].Price.String() != "USD 4.50" {
		t.Fatalf("profile not preserved: %+v", got.Profile)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("expected error loading missing file")
	}
}
