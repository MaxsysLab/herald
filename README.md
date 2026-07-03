# Herald

**A configurable AI virtual front-office officer for small businesses** — an
intelligent, authoritative, effective communicator that represents *one* business at
its digital frontend. Not a generic chatbot: Herald *knows* the business it fronts,
stays in scope, and hands off to a human rather than guess.

> **Open source, Apache-2.0.** Herald is built by [MaxsysLab](https://github.com/MaxsysLab)
> — whose mission is commercial-grade applications that communities and nonprofits can
> run in production — with students, interns, and volunteers. Matured versions are
> released for anyone to use as they please. Contributions welcome; see
> [CONTRIBUTING.md](CONTRIBUTING.md).

## Status

`v0.1.0-dev` — the full answer pipeline **and** the HTTP server (chat API +
embeddable widget + management console) are built and tested. Lean Go core, zero
external dependencies, GPU-free.

```bash
make demo     # run the full pipeline offline — no API keys needed
make check    # gofmt + go vet + go test  (the quality gate)

go run ./cmd/herald serve        # start the server + console on :8080
#   http://localhost:8080/admin    management console (edit config, add knowledge, test chat)
#   http://localhost:8080/widget   standalone chat test page
#   <script src="http://HOST:8080/widget.js" defer></script>   embed on any website
```

`serve` runs with **no keys** out of the box (structured facts + verbatim retrieval).
Set an LLM key (e.g. `GROQ_API_KEY`) to enable natural drafting; set an embeddings key
(`OPENAI_API_KEY` or `VOYAGE_API_KEY`) and an `embed.provider` to enable
production-quality retrieval. See "Embeddings & retrieval" below.

## Embeddings & retrieval

Retrieval quality depends on the embedder. The offline default is a crude bag-of-words
stand-in (`FakeEmbedder`) — fine for structured facts and offline tests, but it can't
stem ("guests" ≠ "guest") or resolve synonyms, so it **errs toward escalating** weak
matches. That's the safe failure mode (defer to a human, never answer wrong), but for a
live site configure a real embedder:

```json
"embed": { "provider": "openai" }   // or "voyage"; key from OPENAI_API_KEY / VOYAGE_API_KEY
```

The demo answers exact facts deterministically, grounds open questions in a knowledge
base, and escalates anything it can't support:

```
Q: How much is a cappuccino?      → [answer]   Cappuccino is USD 4.50.        (profile:pricing)
Q: Do you have wifi?              → [answer]   Yes! Free guest wifi ...        (kb:amenities.md)
Q: How much does catering cost?   → [escalate] ... connect you with someone    (price_not_listed)
Q: What's the meaning of life?    → [escalate] ... connect you with someone    (ungrounded)
```

## Why it exists

Anyone can glue a free voice bot out of GitHub AI projects. It has no idea what the
business actually charges, promises things that aren't true, and no one stands behind
it. Herald is the opposite: a real, owned product whose core IP is **business
awareness + guardrails** — so it can be trusted to speak *for* a business.

## Architecture (v1)

A **lean core** with **web chat as the default channel**; everything heavy (voice,
email, third-party messaging, integrations) is an **opt-in module**. See
[`docs/HERALD-V1-DESIGN.md`](docs/HERALD-V1-DESIGN.md) for the full design.

```
core/       shared vocabulary (Query, Reply, Citation, Verdict, Escalation)
llm/        provider-agnostic generation — Groq default, OpenAI/Grok/Ollama/own; Claude seam
embed/      embedding seam (local CPU model or Voyage) — Anthropic has no embeddings API
knowledge/  business awareness: structured profile (exact facts) + vector KB (RAG)
guard/      input screen + grounding gate (cite-or-handoff)
officer/    the orchestrator: screen → structured lookup → retrieve → ground → answer/escalate
config/     per-business config (profile + provider + threshold) as one JSON file
voice/      opt-in voice seam (free browser / economy Groq / premium ElevenLabs)
channel/    channel seam (web chat default; email/messaging later)
cmd/herald/ CLI entry point (offline demo today; serve later)
```

### The apple/orange guarantee

A receptionist that's confidently wrong is worse than none. Herald prevents that two
ways: **exact facts (prices, hours, policies) are deterministic profile lookups, never
paraphrased by an LLM**; and any answer not backed by a structured fact or a
high-confidence retrieved passage **escalates to a human instead of guessing**.

## LLM providers

Groq is the pre-selected, unbiased default. Point Herald elsewhere with a config
change — no code edit:

| Provider | Default endpoint | Env var |
|---|---|---|
| `groq` (default) | `api.groq.com/openai/v1` | `GROQ_API_KEY` |
| `openai` | `api.openai.com/v1` | `OPENAI_API_KEY` |
| `grok` | `api.x.ai/v1` | `XAI_API_KEY` |
| `ollama` | `localhost:11434/v1` | — |
| `custom` | your endpoint | (config) |

## Requirements

Go 1.26+. No other dependencies.

## Governance

Maintained by MaxsysLab. Maryanne (maxsys.org) manages and moderates the project on an
interim basis until a dedicated maintainer is appointed. Issues and pull requests are
welcome from everyone.

## License

Apache License 2.0 — see [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE). Use it as you
please, commercially or otherwise, with minimal strings attached.
