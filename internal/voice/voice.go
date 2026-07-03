// Package voice is the opt-in voice module seam. Voice is deferred past the lean v1
// core, and every tier is pluggable so a business pays only for what it enables:
//
//   - Free:    browser Web Speech API (STT+TTS client-side, zero API cost).
//   - Economy: the whole loop on one Groq key — Whisper STT + Orpheus TTS.
//   - Premium: ElevenLabs for a bespoke branded/cloned voice persona.
//
// Herald uses voice providers for I/O only; the reasoning (business awareness,
// grounding, guardrails) always stays in Herald's core — we never hand the
// conversation to a provider's built-in voice agent.
package voice

import "context"

// Recognizer converts speech to text (STT).
type Recognizer interface {
	Transcribe(ctx context.Context, audio []byte, mimeType string) (string, error)
	Name() string
}

// Synthesizer converts text to speech (TTS), returning audio bytes and their MIME type.
type Synthesizer interface {
	Speak(ctx context.Context, text string) (audio []byte, mimeType string, err error)
	Name() string
}
