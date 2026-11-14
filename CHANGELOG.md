# Changelog

All notable changes to Herald are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/).

## [Unreleased]

The code reports its version as `0.1.0-dev`. Nothing has been tagged yet. This section
becomes `0.1.0` at the first release.

### Added

- The answer pipeline: input screen, deterministic lookup of exact facts (prices,
  hours, policies), retrieval from a knowledge base with a grounding threshold, an
  optional LLM draft held to the retrieved passages, and escalation to a human when an
  answer cannot be supported.
- The HTTP server: `POST /api/chat`, `GET /api/greeting`, an embeddable chat widget
  (`/widget.js`), a standalone test page (`/widget`), a management console (`/admin`)
  for editing the configuration, adding knowledge and testing the chat, and `/healthz`.
- One JSON configuration file per business, reloaded when saved from the console.
- LLM providers behind one interface: Groq (default), OpenAI, Grok, Ollama, or any
  OpenAI-compatible endpoint.
- Embedders: an offline stand-in for tests and demos, and hosted OpenAI or Voyage.
- An offline demo (`herald demo`) that needs no API keys.
- Design documents: the v1 design, and a note of channel-layer patterns to consider.

### Project

- Stated who stands behind the project: published by Maxsys International
  Incorporated through MaxsysLab, built and contributed by CSEAM Technology
  Corporation.
- Added the community files: code of conduct, security policy, support, governance,
  maintainers, issue and pull request templates, code owners.
- Added documentation: quick start, configuration reference, deployment, security
  model, data handling, open questions.
- Added checks that run on every pull request: build and test, sign-off (DCO),
  licence metadata (REUSE), dependency review, secret scan, pinned actions, and the
  OpenSSF Scorecard.

### Known limits

- The management console and its write routes have no sign-in.
- The chat API accepts requests from any origin and has no rate limit.
- The knowledge index is held in memory and rebuilt at start.

See [docs/SECURITY-MODEL.md](docs/SECURITY-MODEL.md).
