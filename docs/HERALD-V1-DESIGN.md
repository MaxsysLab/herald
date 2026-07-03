# Herald — v1 Design

_Working name: **Herald** (placeholder — rename freely). Last updated: 2026-07-03._

## Mission

**Represent a business at its digital frontend as an intelligent, authoritative,
effective communicator.** One job, done well. Herald is a configurable AI virtual
front-office officer (receptionist / front-office manager) that a small business
"onboards" with its own knowledge, then deploys to answer customers on its behalf.

It is **not** a generic chatbot and not a free GitHub mashup. The whole value is that
it can be *trusted to speak for a business*: it knows the business's real facts, stays
in scope, and hands off to a human rather than guess.

## Origin & competitive framing

The inquiry came from a business whose staffer demoed a free voice chatbot glued from
open-source AI projects (WhatsApp + an open agent). Those demos talk, but have no
business awareness, no guardrails, and no accountability. Herald's differentiator is
exactly what they lack: **business awareness + guardrails**, in a real owned product.

## Principles

1. **Single purpose.** Front-office representation only. Resist feature-creep.
2. **Lean core, pluggable modules.** Web chat is the always-on core channel; voice,
   email, third-party messaging, and integrations are opt-in modules. Resource cost
   scales with what a client enables.
3. **Lightweight via cloud API keys.** The model is an API call, so the deployable box
   stays small — no GPU. A Go single binary, self-hostable locally or in the cloud.
4. **Provider-agnostic, unbiased default.** The LLM backend is user-selectable; **Groq
   is the pre-selected default** as a deliberate no-favorite stance. Users can switch
   to Claude, OpenAI, Grok, Ollama, or their own endpoint.
5. **Cost-tiering is first-class.** The product must scale down to a tiny business
   without a stack of API bills. Every paid dependency is optional with a cheaper/free
   fallback.
6. **Verifier ≠ generator.** The grounding gate and guardrails are the product's core
   IP and live in *our* core — never delegated to a provider's built-in agent.

## Why Go

Herald is a customer-deployed, install-per-client, must-stay-lite web service — unlike
the internal Python AI tooling elsewhere in the portfolio. Go wins here on the axes
that matter for that shape: a **single static binary** (scp one file, no runtime/venv
drift on client boxes), **low footprint** (~10–25 MB RAM idle), and **goroutine
concurrency** for many simultaneous chat/webhook/stream connections. The usual
"Python has the AI ecosystem" objection collapses because the intelligence is remote
(an API call), the vector store is embeddable, and we avoid heavy frameworks by
design. v1 has **zero external Go dependencies** — stdlib only.

## Architecture

```
Customer ──▶ channel ──▶ officer ──▶ [ guard.Screen ]
                                      [ knowledge: structured lookup (exact facts) ]
                                      [ knowledge: retrieve (RAG) → guard.Ground ]
                                      [ llm: draft (constrained to evidence) ]
                                      └──▶ Reply (answer + citations) | Escalation
```

- **`core`** — shared vocabulary (`Query`, `Reply`, `Citation`, `Verdict`,
  `Escalation`). Depends on nothing; no import cycles.
- **`llm`** — provider-agnostic generation. `OpenAICompatClient` (stdlib HTTP) serves
  Groq/OpenAI/Grok/Ollama/own; `Resolve(ProviderConfig)` applies defaults and reads
  the key from the environment. Groq is the default. Claude is a documented adapter
  seam (different wire format) — it returns an explicit error rather than mis-routing.
- **`embed`** — embedding seam. Anthropic has **no embeddings endpoint**, so embeddings
  come from a small local CPU model (private, GPU-free) or a hosted provider (Voyage).
  `FakeEmbedder` (deterministic, hash-based) powers offline tests and the demo.
- **`knowledge`** — the business-awareness layer (see below).
- **`guard`** — `Screen` (input supervisor) and `Ground` (output gate).
- **`officer`** — the orchestrator composing the four layers.
- **`config`** — one business's whole setup as a single JSON file (profile + provider
  + grounding threshold). API keys are never stored here — they come from the env.
- **`voice`**, **`channel`** — module seams (interfaces only in v1).

## Business awareness — "onboarding a new hire"

The core mechanism, in four layers, each defending a different failure mode:

1. **Structured business profile (facts of record).** Services, **prices**, hours,
   policies, contacts — stored as typed data. Volatile must-be-exact facts are answered
   by *deterministic lookup*, never paraphrased by an LLM. Prices are exact strings,
   never floats. This is the primary defense against the "price of apples → color of
   oranges" failure.
2. **Knowledge base / wiki (training manual).** Unstructured docs/FAQ/site content,
   chunked, embedded into a vector store, retrieved by similarity, and always cited.
3. **Onboarding pipeline (the training).** Ingest → chunk → embed → index, ending in a
   **gap review** where the owner sees what the officer does and doesn't know and fills
   holes — like quizzing a new hire before their first shift. _(Dashboard-driven;
   ingestion/index built, gap-review UI is later.)_
4. **Grounding guardrail (don't make it up).** An answer may be sent only if it traces
   to a structured fact or a passage above the grounding threshold. Otherwise the
   officer **says so and hands off**. Grounding reduces but never eliminates
   hallucination, so this gate sits *under* escalation, deliberately conservative.

### The apple/orange guarantee (implemented)

- Exact facts → deterministic profile lookup, LLM never consulted. _(officer_test:
  `TestPriceAnsweredFromStructuredFact` also asserts the LLM was **not** called.)_
- Unlisted price → escalate, never invent a number.
  _(`TestPriceNotListedEscalates`.)_
- Ungrounded question → escalate. _(`TestUngroundedEscalates`; the demo's "meaning of
  life" query.)_
- The grounding threshold is the key tuning knob; a real embedder has cleaner
  score separation than the crude fake used in tests/demo.

## Cost tiers

Every paid dependency is optional. The economy path can run the whole product on a
**single API relationship** (Groq) plus a local CPU embedder.

| Concern | Free | Economy | Premium |
|---|---|---|---|
| LLM | Ollama (local) | **Groq** (default) | Claude / OpenAI / Grok |
| Embeddings | local CPU model | local CPU model | Voyage |
| Voice STT+TTS | browser Web Speech API | **Groq** Whisper + Orpheus (one key) | ElevenLabs voice persona |
| Vector store | in-memory / sqlite-vec | sqlite-vec | sqlite-vec |

## Voice (deferred module)

Voice is opt-in and cost-tiered. Providers do **I/O only** — the reasoning stays in
Herald's core; we never hand the conversation to a provider's built-in voice agent.
Order: browser voice first, telephony later. Groq covers the economy voice loop
(Whisper STT ~$0.04/hr + Orpheus TTS) on the same key as the LLM; ElevenLabs is the
premium branded-voice upgrade.

## Feature cut

- **v1 core (BUILT & live-verified):** Go orchestrator · web-chat channel over HTTP
  (JSON `/api/chat`, CORS) · **embeddable widget** (`/widget.js`) · **management
  console** (`/admin`: edit config, add knowledge, live test-chat, embed snippet) ·
  business awareness (structured profile + RAG) · input + grounding guardrails ·
  configurable escalation · per-business JSON config (hot-reload on save) ·
  provider-agnostic LLM (Groq default) with live wiring · real embedder option
  (OpenAI/Voyage) with a fake offline default · module seams. Runs keyless out of the
  box (structured facts + verbatim retrieval); lights up drafting/embeddings with keys.
- **v1 next:** sqlite-vec persistence (vs in-memory); a bundled local CPU embedder;
  per-field profile editor (vs JSON) + onboarding gap-review; auth on the console.
- **v1.1 modules:** email channel; browser voice.
- **v2 modules:** telephony voice; WhatsApp/SMS; per-client integrations
  (calendar/CRM/bookings) — each behind the workflow engine.

## Reference patterns (from research)

- Separate language understanding from business logic (Rasa CALM): the LLM interprets,
  deterministic flows/lookups decide — the core anti-hallucination pattern.
- Grounding over a content library (Intercom Fin) with cite-or-handoff.
- Supervisor guardrails (Sierra): input + output checks around the core.
- Thin own-abstractions over heavy frameworks (Botpress interfaces) — matches our
  no-heavy-framework / no-MCP stance.
- Embedded vector store (sqlite-vec) — no separate DB service, GPU-free.

## Open decisions

- Final product name.
- Dashboard stack (server-rendered Go templates vs. a small SPA).
- Workflow/orchestration engine shape for multi-step tasks (booking/lead-capture) —
  handoff-graph vs. sequential, introduced when the first real workflow lands.
- Which embedder ships as the default local CPU model.
