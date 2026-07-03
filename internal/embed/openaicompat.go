package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

// OpenAICompatEmbedder calls any OpenAI-compatible /embeddings endpoint — OpenAI,
// Voyage, or a customer's own — using only the standard library. It is the real
// (higher-quality) embedder for production RAG; FakeEmbedder remains the offline
// default. Anthropic has no embeddings endpoint, so Claude is never used here.
type OpenAICompatEmbedder struct {
	BaseURL  string
	APIKey   string
	Model    string
	Provider string
	D        int // known dimensionality; learned from the first response when 0
	HTTP     *http.Client
}

// NewOpenAICompatEmbedder builds an embedder with a default HTTP timeout.
func NewOpenAICompatEmbedder(provider, baseURL, apiKey, model string) *OpenAICompatEmbedder {
	return &OpenAICompatEmbedder{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
		Provider: provider,
		HTTP:     &http.Client{Timeout: 60 * time.Second},
	}
}

// Name identifies the provider.
func (e *OpenAICompatEmbedder) Name() string { return e.Provider }

// Dim returns the dimensionality (0 until the first Embed call learns it).
func (e *OpenAICompatEmbedder) Dim() int { return e.D }

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Embed returns one vector per input, preserving input order.
func (e *OpenAICompatEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	body, err := json.Marshal(embedRequest{Model: e.Model, Input: texts})
	if err != nil {
		return nil, fmt.Errorf("embed: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.BaseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embed: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if e.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.APIKey)
	}
	resp, err := e.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embed: request %s: %w", e.Provider, err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, fmt.Errorf("embed: read response: %w", err)
	}
	var parsed embedResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("embed: decode (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return nil, fmt.Errorf("embed: %s returned %d: %s", e.Provider, resp.StatusCode, parsed.Error.Message)
		}
		return nil, fmt.Errorf("embed: %s returned status %d", e.Provider, resp.StatusCode)
	}
	if len(parsed.Data) != len(texts) {
		return nil, fmt.Errorf("embed: expected %d vectors, got %d", len(texts), len(parsed.Data))
	}
	sort.Slice(parsed.Data, func(i, j int) bool { return parsed.Data[i].Index < parsed.Data[j].Index })
	out := make([][]float32, len(parsed.Data))
	for i, d := range parsed.Data {
		out[i] = d.Embedding
		if e.D == 0 {
			e.D = len(d.Embedding)
		}
	}
	return out, nil
}
