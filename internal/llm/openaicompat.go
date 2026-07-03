package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAICompatClient talks to any OpenAI-compatible /chat/completions endpoint —
// Groq, OpenAI, Grok (xAI), Ollama, or a customer's own gateway. It uses only the
// standard library (no vendor SDK), keeping the deployable binary lean.
type OpenAICompatClient struct {
	BaseURL  string // e.g. https://api.groq.com/openai/v1 (no trailing slash)
	APIKey   string // bearer token; may be empty for keyless local endpoints
	Model    string // provider-specific model id
	Provider string // label for Name()
	HTTP     *http.Client
}

// NewOpenAICompatClient builds a client with a sane default HTTP timeout.
func NewOpenAICompatClient(provider, baseURL, apiKey, model string) *OpenAICompatClient {
	return &OpenAICompatClient{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
		Provider: provider,
		HTTP:     &http.Client{Timeout: 60 * time.Second},
	}
}

// Name returns the provider label.
func (c *OpenAICompatClient) Name() string { return c.Provider }

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete runs one chat completion against the configured endpoint.
func (c *OpenAICompatClient) Complete(ctx context.Context, req Request) (Response, error) {
	msgs := make([]chatMessage, 0, len(req.Messages)+1)
	if req.System != "" {
		msgs = append(msgs, chatMessage{Role: string(RoleSystem), Content: req.System})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, chatMessage{Role: string(m.Role), Content: m.Content})
	}

	body, err := json.Marshal(chatRequest{
		Model:       c.Model,
		Messages:    msgs,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	})
	if err != nil {
		return Response{}, fmt.Errorf("llm: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("llm: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("llm: request %s: %w", c.Provider, err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return Response{}, fmt.Errorf("llm: read response: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Response{}, fmt.Errorf("llm: decode response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return Response{}, fmt.Errorf("llm: %s returned %d: %s", c.Provider, resp.StatusCode, parsed.Error.Message)
		}
		return Response{}, fmt.Errorf("llm: %s returned status %d", c.Provider, resp.StatusCode)
	}
	if len(parsed.Choices) == 0 {
		return Response{}, fmt.Errorf("llm: %s returned no choices", c.Provider)
	}

	model := parsed.Model
	if model == "" {
		model = c.Model
	}
	return Response{Text: parsed.Choices[0].Message.Content, Model: model}, nil
}
