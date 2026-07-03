// Package core holds Herald's shared vocabulary: the types that flow between the
// channel layer, the officer (orchestrator), and the guardrails. It deliberately
// depends on nothing else in the tree so every other package can import it without
// creating cycles.
package core

// Channel identifies where a conversation is taking place. Web chat is the default
// and only always-on channel; email and voice are opt-in modules (see docs).
type Channel string

const (
	ChannelWebChat Channel = "web_chat"
	ChannelEmail   Channel = "email"
	ChannelVoice   Channel = "voice"
)

// Query is a single inbound customer message the officer must answer.
type Query struct {
	Text    string
	Channel Channel
	Locale  string // BCP-47, optional (e.g. "en", "en-US")
}

// Citation records a source that grounds part of an answer. Every non-escalated
// reply must carry at least one — an answer with no citation is, by construction,
// not grounded and must not be sent (this is the anti-hallucination invariant).
type Citation struct {
	Source string  // "profile:pricing", "kb:refund-policy.md", ...
	Quote  string  // the exact supporting text
	Score  float64 // retrieval score for KB sources; 1.0 for structured facts
}

// Verdict is the guardrail outcome for a proposed reply.
type Verdict string

const (
	// VerdictAnswer: grounded in a structured fact or a high-confidence passage.
	VerdictAnswer Verdict = "answer"
	// VerdictEscalate: in scope but not confidently grounded — hand off to a human
	// rather than guess. This is the "price of apples → color of oranges" guard.
	VerdictEscalate Verdict = "escalate"
	// VerdictRefuse: out of scope or unsafe input — decline politely.
	VerdictRefuse Verdict = "refuse"
)

// Escalation captures a safe hand-off to a human.
type Escalation struct {
	Reason  string // machine-readable: "ungrounded", "high_risk", "explicit_request"
	Contact string // resolved from the business profile's escalation contact
	Summary string // short summary of what the customer wanted, for the human
}

// Reply is what the officer returns for a Query. Exactly one of the answer path
// (Verdict == VerdictAnswer, Text + Citations populated) or the escalation path
// (Escalation populated) is meaningful for a given verdict.
type Reply struct {
	Verdict    Verdict
	Text       string
	Citations  []Citation
	Escalation *Escalation
}
