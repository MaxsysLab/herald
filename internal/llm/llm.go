// Package llm is Herald's provider-agnostic language-model seam. The default keyed
// provider is Groq (fast, cheap open-model inference), but the same interface serves
// OpenAI, Grok (xAI), a local Ollama endpoint, or a customer's own OpenAI-compatible
// endpoint — the choice is a user-facing setting, not a hardcode. Groq is pre-selected
// as a deliberate unbiased stance. Claude uses a different wire format and is a
// documented adapter seam (see provider.go), not wired in this scaffold.
package llm

import "context"

// Role is a chat message role.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is one turn in a chat request.
type Message struct {
	Role    Role
	Content string
}

// Request is a single completion request. System is sent as a leading system message.
type Request struct {
	System      string
	Messages    []Message
	MaxTokens   int
	Temperature float64
}

// Response is a completion result.
type Response struct {
	Text  string
	Model string
}

// Client is the minimal generation interface every provider implements.
type Client interface {
	// Complete runs one chat completion.
	Complete(ctx context.Context, req Request) (Response, error)
	// Name identifies the provider (e.g. "groq", "openai", "fake").
	Name() string
}
