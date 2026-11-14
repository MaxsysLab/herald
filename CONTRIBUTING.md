# Contributing to Herald

Welcome. This is a learning-friendly project published by Maxsys International
Incorporated through MaxsysLab. Students and interns are the intended contributors, so
questions are encouraged and small pull requests are great.

## New to open source? Start here

1. Run the project on your own machine: [docs/QUICKSTART.md](docs/QUICKSTART.md) takes
   a few minutes and needs no API keys.
2. Pick an issue labelled **good first issue**. Comment on it so others know you are
   working on it. If nothing fits, ask in a new issue and a maintainer will suggest one.
3. Read "The one rule that matters most" below.
4. Open a small pull request. A maintainer will review it and help you get it merged.

You do not have to write code to contribute. Clearer documentation, a bug report with
the exact question and reply, and help answering other people's issues all count.

## Ground rules

- Be kind and follow the [Code of Conduct](CODE_OF_CONDUCT.md).
- **File an issue before a large pull request.** Small fixes (typos, obvious bugs) can
  go straight to a pull request. For anything substantial, open an issue first so we
  can agree on the approach.
- Security problems are never reported in a public issue. Follow [SECURITY.md](SECURITY.md).

## Prerequisites

- Go 1.26+ (`go version`)
- That is all. Herald has **zero external dependencies** (standard library only).

## Build, test, run

```bash
make check      # gofmt + go vet + go test: the full quality gate (run before every PR)
make demo       # run the answer pipeline offline (no API keys)
make build      # build ./bin/herald
go run ./cmd/herald serve -addr 127.0.0.1:8080   # start the server and console
```

Open `http://localhost:8080/admin` for the management console and
`http://localhost:8080/widget` for the chat widget.

## How the code is laid out

Read [`docs/HERALD-V1-DESIGN.md`](docs/HERALD-V1-DESIGN.md) first. It explains the
architecture and the "business awareness" design. Then:

```
internal/core       shared types (Query, Reply, Citation, Verdict, Escalation)
internal/llm        provider-agnostic generation (Groq default; OpenAI, Grok, Ollama, your own)
internal/embed      embedding seam (offline stand-in by default; OpenAI or Voyage as options)
internal/knowledge  business awareness: structured profile and vector knowledge base
internal/guard      input screen and grounding gate (cite or hand off)
internal/officer    the orchestrator that composes it all
internal/server     HTTP: chat API, embeddable widget, management console
cmd/herald          the command (serve, demo, version)
```

## The one rule that matters most

Herald's whole point is to be **trustworthy**: it answers exact facts deterministically
and **escalates to a human when it is not sure, rather than guessing**. When you change
answering behavior, preserve that. A change that makes Herald confidently wrong is worse
than no change. Add a test that proves the safe behavior still holds.

## Making a change

1. Fork the repository and create a branch from `main`.
2. Make your change. Keep pull requests focused and small.
3. Add or update tests, and run `make check`.
4. Write commit messages in the [Conventional Commits](https://www.conventionalcommits.org/)
   style, for example `fix: handle empty input`.
5. **Sign off every commit** (see below).
6. Open a pull request, fill in the template, and link the issue it closes.

## Developer Certificate of Origin (DCO)

This project uses the **DCO**, not a contributor licence agreement. By signing off, you
certify that you wrote the contribution or have the right to submit it under the
project's licence. The full text is at <https://developercertificate.org/>. Sign off by
adding a `Signed-off-by` line to each commit:

```bash
git commit -s -m "fix: handle empty input"
```

This appends `Signed-off-by: Your Name <your@email>`, matching your git configuration.
A check on every pull request enforces it. Fix a missing sign-off with
`git commit --amend -s`, or `git rebase --signoff` for a series.

## Good first areas

- A `sqlite-vec` vector-store backend (persist the knowledge index to disk).
- A bundled local CPU embedder so quality retrieval needs no API key.
- A friendlier per-field profile editor in the console (instead of raw JSON).
- Sign-in for the management console.
- New channel adapters (email, messaging) behind `internal/channel`.

## Standards

- `gofmt` clean, `go vet` clean, `go test ./...` green. `make check` runs all three.
- Keep the dependency count at zero unless there is a strong reason; discuss first.
- Small, focused pull requests with a clear description and a test.

## Review and merge

We aim to give a first response within 3 business days. Reviews are about the code,
never the person. Once a change is approved and every required check passes, a
maintainer merges it.

## Reporting issues

Open a GitHub issue describing what you expected and what happened. For a wrong answer,
include the question asked and the reply received. That is the most useful bug report.

## Growing here

Sustained, good contribution earns responsibility. [GOVERNANCE.md](GOVERNANCE.md)
describes the path from first contribution to maintainer. Herald is looking for a
dedicated maintainer, so that path is real.
