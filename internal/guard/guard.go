// Package guard holds Herald's safety and grounding checks — the discipline that
// keeps an "authoritative communicator" from being confidently wrong. Two gates:
//
//   - Screen: an input supervisor run before the officer answers (scope/abuse). The
//     P0 version is a cheap heuristic; an LLM classifier is the documented upgrade.
//   - Ground: the output gate. An answer may be sent only when it is backed by a
//     structured fact or a high-confidence retrieved passage; otherwise the officer
//     must escalate. Grounding reduces but never eliminates hallucination, so this
//     gate is deliberately conservative — when unsure, defer to a human.
package guard

import (
	"strings"

	"github.com/MaxsysLab/herald/internal/core"
)

// ScreenResult is the outcome of the input supervisor.
type ScreenResult struct {
	OK     bool
	Reason string // machine-readable when !OK: "empty", "prompt_injection"
}

// Screen performs a cheap pre-answer check on raw customer input. It is intentionally
// minimal — a seam, not a security boundary — with an LLM-classifier upgrade planned.
func Screen(text string) ScreenResult {
	t := strings.TrimSpace(text)
	if t == "" {
		return ScreenResult{OK: false, Reason: "empty"}
	}
	low := strings.ToLower(t)
	for _, marker := range injectionMarkers {
		if strings.Contains(low, marker) {
			return ScreenResult{OK: false, Reason: "prompt_injection"}
		}
	}
	return ScreenResult{OK: true}
}

var injectionMarkers = []string{
	"ignore previous instructions",
	"ignore all previous",
	"disregard your instructions",
	"you are now",
	"system prompt",
}

// Evidence is what the grounding gate weighs for a proposed answer.
type Evidence struct {
	HasStructuredFact bool    // a deterministic profile lookup matched (price, hours, policy)
	TopScore          float64 // best retrieval score among candidate passages
	MinScore          float64 // grounding threshold; below it, retrieval is not trusted
}

// Ground decides whether a proposed answer is grounded. A structured fact is always
// sufficient; otherwise the top passage must clear MinScore. Anything else escalates.
// It returns the verdict and a machine-readable reason.
func Ground(ev Evidence) (core.Verdict, string) {
	if ev.HasStructuredFact {
		return core.VerdictAnswer, "structured_fact"
	}
	if ev.TopScore >= ev.MinScore && ev.MinScore > 0 {
		return core.VerdictAnswer, "retrieved_passage"
	}
	return core.VerdictEscalate, "ungrounded"
}
