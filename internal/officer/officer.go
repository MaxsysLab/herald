// Package officer is Herald's orchestrator — the "virtual front-office officer" that
// turns a customer Query into a grounded Reply or a safe escalation. It composes the
// four business-awareness layers:
//
//  1. Screen  (guard)      — reject empty/abusive input.
//  2. Structured lookup    — exact facts (price/hours/policy) answered deterministically
//     from the profile, never paraphrased by an LLM.
//  3. Retrieve + ground    — otherwise search the knowledge base; if nothing clears the
//     grounding threshold, escalate (cite-or-handoff).
//  4. Draft                — phrase a grounded answer via the (swappable) LLM, strictly
//     constrained to the retrieved passages.
package officer

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/MaxsysLab/herald/internal/core"
	"github.com/MaxsysLab/herald/internal/guard"
	"github.com/MaxsysLab/herald/internal/knowledge"
	"github.com/MaxsysLab/herald/internal/llm"
)

// DefaultOfficerName is the persona the officer uses when the profile doesn't set one.
// Unisex by design so the same default fits any business.
const DefaultOfficerName = "Herald"

// Officer answers customer queries for one business.
type Officer struct {
	Profile   *knowledge.BusinessProfile
	Retriever *knowledge.Retriever // may be nil (structured-only officer)
	LLM       llm.Client           // may be nil (structured-only officer)
	MinScore  float64              // grounding threshold; defaults to 0.35 when <= 0
	TopK      int                  // passages to retrieve; defaults to 4 when <= 0
}

// Name is the officer's persona name (defaults to DefaultOfficerName).
func (o *Officer) Name() string {
	if o.Profile != nil && o.Profile.OfficerName != "" {
		return o.Profile.OfficerName
	}
	return DefaultOfficerName
}

// Greeting is the officer's opening line for a new conversation.
func (o *Officer) Greeting() string {
	return fmt.Sprintf("Hi, I'm %s, the assistant for %s. How can I help you today?", o.Name(), o.Profile.Name)
}

func (o *Officer) minScore() float64 {
	if o.MinScore <= 0 {
		return 0.35
	}
	return o.MinScore
}

func (o *Officer) topK() int {
	if o.TopK <= 0 {
		return 4
	}
	return o.TopK
}

// Answer runs the full pipeline for one query.
func (o *Officer) Answer(ctx context.Context, q core.Query) (core.Reply, error) {
	// 1. Input supervisor.
	if s := guard.Screen(q.Text); !s.OK {
		if s.Reason == "empty" {
			return core.Reply{Verdict: core.VerdictRefuse, Text: "Sorry, I didn't catch that — could you rephrase?"}, nil
		}
		return core.Reply{Verdict: core.VerdictRefuse, Text: "I can only help with questions about " + o.Profile.Name + "."}, nil
	}

	// 2. Structured facts — deterministic, exact, never invented.
	if reply, ok := o.structuredAnswer(q.Text); ok {
		return reply, nil
	}

	// 3. Retrieve + ground.
	var top []knowledge.Scored
	if o.Retriever != nil {
		var err error
		top, err = o.Retriever.Retrieve(ctx, q.Text, o.topK())
		if err != nil {
			return core.Reply{}, fmt.Errorf("officer: retrieve: %w", err)
		}
	}
	topScore := 0.0
	if len(top) > 0 {
		topScore = top[0].Score
	}
	verdict, reason := guard.Ground(guard.Evidence{TopScore: topScore, MinScore: o.minScore()})
	if verdict != core.VerdictAnswer {
		return o.escalate(reason, q.Text), nil
	}

	// 4. Draft a grounded answer from the passages that cleared the threshold.
	used := passagesAbove(top, o.minScore())
	if o.LLM == nil {
		// No generator wired: return the top passage verbatim rather than nothing.
		return core.Reply{
			Verdict:   core.VerdictAnswer,
			Text:      used[0].Passage.Text,
			Citations: citations(used),
		}, nil
	}
	draft, err := o.LLM.Complete(ctx, llm.Request{
		System:    o.systemPrompt(used),
		Messages:  []llm.Message{{Role: llm.RoleUser, Content: q.Text}},
		MaxTokens: 400,
	})
	if err != nil {
		return core.Reply{}, fmt.Errorf("officer: draft: %w", err)
	}
	return core.Reply{
		Verdict:   core.VerdictAnswer,
		Text:      strings.TrimSpace(draft.Text),
		Citations: citations(used),
	}, nil
}

// structuredAnswer handles exact-fact intents (price, hours, policy) from the profile.
func (o *Officer) structuredAnswer(query string) (core.Reply, bool) {
	low := strings.ToLower(query)

	if containsAny(low, "price", "prices", "cost", "costs", "how much", "charge", "charges", "fee", "fees", "rate", "rates") {
		if svc, ok := o.Profile.PriceOf(query); ok {
			return grounded(
				fmt.Sprintf("%s is %s.", svc.Name, svc.Price.String()),
				core.Citation{Source: "profile:pricing", Quote: svc.Name + ": " + svc.Price.String(), Score: 1.0},
			), true
		}
		// Price asked but not on record → defer rather than invent a number.
		return o.escalate("price_not_listed", query), true
	}

	if containsAny(low, "hour", "hours", "open", "opens", "opened", "opening", "close", "closes", "closed", "when are you") && len(o.Profile.Hours) > 0 {
		var b strings.Builder
		for i, h := range o.Profile.Hours {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(h.Day + " " + h.Open + "–" + h.Close)
			if h.Note != "" {
				b.WriteString(" (" + h.Note + ")")
			}
		}
		return grounded(
			o.Profile.Name+" hours: "+b.String()+".",
			core.Citation{Source: "profile:hours", Quote: b.String(), Score: 1.0},
		), true
	}

	for i := range o.Profile.Policies {
		p := &o.Profile.Policies[i]
		if strings.Contains(low, strings.ToLower(p.Topic)) {
			return grounded(p.Text, core.Citation{Source: "profile:policy:" + p.Topic, Quote: p.Text, Score: 1.0}), true
		}
	}

	return core.Reply{}, false
}

func (o *Officer) escalate(reason, query string) core.Reply {
	return core.Reply{
		Verdict: core.VerdictEscalate,
		Text:    "That's a great question — let me connect you with someone from " + o.Profile.Name + " who can help.",
		Escalation: &core.Escalation{
			Reason:  reason,
			Contact: o.Profile.EscalationContact(),
			Summary: "Customer asked: " + query,
		},
	}
}

// systemPrompt builds the strictly-grounded instruction for the draft step.
func (o *Officer) systemPrompt(used []knowledge.Scored) string {
	var b strings.Builder
	b.WriteString("You are ")
	b.WriteString(o.Name())
	b.WriteString(", the front-office assistant for ")
	b.WriteString(o.Profile.Name)
	if o.Profile.BrandVoice != "" {
		b.WriteString(". Voice: " + o.Profile.BrandVoice)
	}
	b.WriteString(".\nAnswer ONLY using the context below. If it does not contain the answer, say you'll connect the customer with a team member — never guess.\n\nContext:\n")
	for _, s := range used {
		b.WriteString("- " + s.Passage.Text + "\n")
	}
	return b.String()
}

func grounded(text string, c core.Citation) core.Reply {
	return core.Reply{Verdict: core.VerdictAnswer, Text: text, Citations: []core.Citation{c}}
}

func passagesAbove(scored []knowledge.Scored, min float64) []knowledge.Scored {
	var out []knowledge.Scored
	for _, s := range scored {
		if s.Score >= min {
			out = append(out, s)
		}
	}
	return out
}

func citations(scored []knowledge.Scored) []core.Citation {
	out := make([]core.Citation, 0, len(scored))
	for _, s := range scored {
		out = append(out, core.Citation{Source: s.Passage.Source, Quote: s.Passage.Text, Score: s.Score})
	}
	return out
}

// containsAny matches whole words or consecutive word phrases, ignoring case and
// treating punctuation and whitespace as separators rather than parts of a phrase.
func containsAny(s string, subs ...string) bool {
	words := intentWords(s)
	for _, sub := range subs {
		phrase := intentWords(sub)
		if len(phrase) == 0 {
			continue
		}
		for i := 0; i+len(phrase) <= len(words); i++ {
			if slices.Equal(words[i:i+len(phrase)], phrase) {
				return true
			}
		}
	}
	return false
}

func intentWords(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsMark(r) && r != '_'
	})
}
