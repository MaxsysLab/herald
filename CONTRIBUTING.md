# Contributing to Herald

Welcome — this is a learning-friendly project run by MaxsysLab. Students and interns
are the intended contributors, so questions are encouraged and small PRs are great.

## Prerequisites

- Go 1.26+ (`go version`)
- That's it — Herald has **zero external dependencies** (standard library only).

## Build, test, run

```bash
make check      # gofmt + go vet + go test — the full quality gate (run before every PR)
make demo       # run the answer pipeline offline (no API keys)
make build      # build ./bin/herald
go run ./cmd/herald serve   # start the server + console on :8080
```

Open `http://localhost:8080/admin` for the management console and
`http://localhost:8080/widget` for the chat widget.

## How the code is laid out

Read [`docs/HERALD-V1-DESIGN.md`](docs/HERALD-V1-DESIGN.md) first — it explains the
architecture and the "business awareness" design. Then:

```
internal/core       shared types (Query, Reply, Citation, Verdict, Escalation)
internal/llm        provider-agnostic generation (Groq default; OpenAI/Grok/Ollama/own)
internal/embed      embedding seam (fake offline default; OpenAI/Voyage option)
internal/knowledge  business awareness: structured profile + vector knowledge base
internal/guard      input screen + grounding gate (cite-or-handoff)
internal/officer    the orchestrator that composes it all
internal/server     HTTP: chat API, embeddable widget, management console
cmd/herald          CLI entry point (serve, demo, version)
```

## The one rule that matters most

Herald's whole point is to be **trustworthy**: it answers exact facts deterministically
and **escalates to a human when it isn't sure, rather than guessing**. When you change
answering behavior, preserve that. A change that makes Herald confidently wrong is worse
than no change. Add a test that proves the safe behavior still holds.

## Good first areas

- A `sqlite-vec` vector-store backend (persist the knowledge index to disk).
- A bundled local CPU embedder so quality retrieval needs no API key.
- A friendlier per-field profile editor in the console (instead of raw JSON).
- New channel adapters (email, messaging) behind `internal/channel`.

## Standards

- `gofmt` clean, `go vet` clean, `go test ./...` green — `make check` runs all three.
- Keep the dependency count at zero unless there's a strong reason; discuss first.
- Small, focused PRs with a clear description and a test.

## Reporting issues

Open a GitHub issue describing what you expected and what happened. For a wrong answer,
include the question asked and the reply received — that's the most useful bug report.
