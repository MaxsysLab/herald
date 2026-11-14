# Herald

[![CI](https://github.com/MaxsysLab/herald/actions/workflows/ci.yml/badge.svg)](https://github.com/MaxsysLab/herald/actions/workflows/ci.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/MaxsysLab/herald/badge)](https://securityscorecards.dev/viewer/?uri=github.com/MaxsysLab/herald)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

**A configurable AI virtual front-office officer for small businesses.** Herald
represents *one* business at its digital front door. It knows that business's facts,
stays in scope, and hands off to a person rather than guess. It is not a generic
chatbot.

> **Open source, Apache-2.0.** Herald is published by Maxsys International
> Incorporated, a nonprofit, through [MaxsysLab](https://github.com/MaxsysLab). It is
> built and contributed by CSEAM Technology Corporation together with students,
> interns, and volunteers. MaxsysLab's mission is commercial-grade applications that
> communities and nonprofits can run in production. Contributions are welcome; see
> [CONTRIBUTING.md](CONTRIBUTING.md).

## Status

`v0.1.0-dev`, an early release. The full answer pipeline and the HTTP server (chat API,
embeddable widget, management console) are built and tested. The core is lean Go with
zero external dependencies and needs no GPU.

The quick start below runs it on your own machine, which is the best place to get to
know it. One piece is still to come before Herald belongs on the open internet: a
sign-in for the management console. Until it is built, keep the console on `localhost`
or behind a protected proxy. [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) shows how, and
[docs/SECURITY-MODEL.md](docs/SECURITY-MODEL.md) explains what Herald protects today.

## Quick start

You need Go 1.26 or newer. Nothing else.

```bash
git clone https://github.com/MaxsysLab/herald && cd herald
make demo                                        # the answer pipeline, offline, no API keys
go run ./cmd/herald serve -addr 127.0.0.1:8080   # the server and console, on this machine only
```

Then open:

- `http://localhost:8080/admin` for the management console (edit the configuration, add knowledge, test the chat)
- `http://localhost:8080/widget` for a standalone chat test page

To put the chat on a website, add one line before `</body>`:

```html
<script src="https://YOUR-HERALD-HOST/widget.js" defer></script>
```

`serve` runs with **no keys** out of the box, answering from structured facts and
quoting retrieved knowledge word for word. Set an LLM key (for example `GROQ_API_KEY`)
for natural drafting, and an embeddings provider for production-quality retrieval. The
full walk-through is in [docs/QUICKSTART.md](docs/QUICKSTART.md).

## How it answers

The demo answers exact facts deterministically, grounds open questions in a knowledge
base, and escalates anything it cannot support:

```
Q: How much is a cappuccino?      → [answer]   Cappuccino is USD 4.50.        (profile:pricing)
Q: Do you have wifi?              → [answer]   Yes! Free guest wifi ...        (kb:amenities.md)
Q: How much does catering cost?   → [escalate] ... connect you with someone    (price_not_listed)
Q: What's the meaning of life?    → [escalate] ... connect you with someone    (ungrounded)
```

### The apple/orange guarantee

A receptionist that is confidently wrong is worse than none. Herald prevents that two
ways. **Exact facts (prices, hours, policies) are deterministic profile lookups, never
paraphrased by an LLM.** And any answer not backed by a structured fact or a
high-confidence retrieved passage **escalates to a human instead of guessing**.

## Why it exists

Anyone can glue a free voice bot out of GitHub AI projects. It has no idea what the
business actually charges, promises things that are not true, and no one stands behind
it. Herald is the opposite: its core is **business awareness and guardrails**, so it
can be trusted to speak *for* a business.

## Documentation

| Document | What it covers |
|---|---|
| [docs/QUICKSTART.md](docs/QUICKSTART.md) | Build, run the demo, start the server, ask your first questions |
| [docs/CONFIGURATION.md](docs/CONFIGURATION.md) | The configuration file, field by field, and how each kind of question is answered |
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) | Running it as a service behind a proxy, and keeping the console private |
| [docs/SECURITY-MODEL.md](docs/SECURITY-MODEL.md) | What Herald protects, and its known limits today |
| [docs/DATA.md](docs/DATA.md) | What data Herald handles and what leaves the machine |
| [docs/HERALD-V1-DESIGN.md](docs/HERALD-V1-DESIGN.md) | The design: principles, architecture, cost tiers, feature cut |
| [docs/OPEN-QUESTIONS.md](docs/OPEN-QUESTIONS.md) | Decisions still open, starting with the project's name |
| [CHANGELOG.md](CHANGELOG.md) | What changed, release by release |

## LLM providers

Groq is the pre-selected, unbiased default. Point Herald elsewhere with a
configuration change and no code edit:

| Provider | Default endpoint | Env var |
|---|---|---|
| `groq` (default) | `api.groq.com/openai/v1` | `GROQ_API_KEY` |
| `openai` | `api.openai.com/v1` | `OPENAI_API_KEY` |
| `grok` | `api.x.ai/v1` | `XAI_API_KEY` |
| `ollama` | `localhost:11434/v1` | none |
| `custom` | your endpoint | (config) |

## Embeddings and retrieval

Retrieval quality depends on the embedder. The offline default is a crude bag-of-words
stand-in (`FakeEmbedder`). It is fine for structured facts and offline tests, but it
cannot stem ("guests" is not "guest") or resolve synonyms, so it **errs toward
escalating** weak matches. That is the safe failure (defer to a human, never answer
wrong), but for a live site configure a real embedder:

```json
"embed": { "provider": "openai" }
```

`"voyage"` works the same way. The key comes from `OPENAI_API_KEY` or `VOYAGE_API_KEY`.

## Architecture (v1)

A **lean core** with **web chat as the default channel**. Everything heavy (voice,
email, third-party messaging, integrations) is an **opt-in module**. See
[docs/HERALD-V1-DESIGN.md](docs/HERALD-V1-DESIGN.md) for the full design.

```
core/       shared vocabulary (Query, Reply, Citation, Verdict, Escalation)
llm/        provider-agnostic generation: Groq default, OpenAI, Grok, Ollama, your own; Claude seam
embed/      embedding seam: offline stand-in, or a hosted embedder (OpenAI, Voyage)
knowledge/  business awareness: structured profile (exact facts) and vector knowledge base
guard/      input screen and grounding gate (cite or hand off)
officer/    the orchestrator: screen, structured lookup, retrieve, ground, answer or escalate
config/     one business's configuration as a single JSON file
server/     HTTP: chat API, embeddable widget, management console
voice/      opt-in voice seam (interfaces only in v1)
channel/    channel seam (web chat today; email and messaging later)
cmd/herald/ the command: serve, demo, version
```

## The name

"Herald" is a working name. The design document says so, and the final name is an open
decision. If you have a better one, tell us why in an issue. What a good name has to do
is written in [docs/OPEN-QUESTIONS.md](docs/OPEN-QUESTIONS.md). Each business can
already give its own officer any name it likes in the configuration.

## Contributing

Students, interns, and first-time contributors are the people this project is for.
Start with an issue labelled **good first issue**, read
[CONTRIBUTING.md](CONTRIBUTING.md), and note that every commit is signed off
(`git commit -s`, the Developer Certificate of Origin). We aim to give a first response
within 3 business days.

## Community

- Code of Conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- Governance and roles: [GOVERNANCE.md](GOVERNANCE.md), [MAINTAINERS.md](MAINTAINERS.md)
- Security: [SECURITY.md](SECURITY.md). Please do not file public issues for vulnerabilities
- Help: [SUPPORT.md](SUPPORT.md)

Herald is managed on an interim basis by the people named in
[MAINTAINERS.md](MAINTAINERS.md) until a dedicated maintainer is appointed. If you
would like to grow into that role, [GOVERNANCE.md](GOVERNANCE.md) describes the path.

## License

Apache License 2.0. See [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE). Use it as you
please, commercially or otherwise, with minimal strings attached.
