# Design input: gateway patterns mined from hermes-agent (2026-07-16)

Source: nousresearch/hermes-agent (MIT) — its gateway serves ~20 chat platforms from one process and is the most battle-hardened part of that codebase. Full triage with file:line citations: `~/dev/research-corpus/reports/2026-07-16-hermes-agent-triage.md`. These are **candidate patterns for Herald's channel layer**, not commitments — review before any implementation follows. Same input applies to Alex.

## Adopt-candidates

### 1. Capability-descriptor delivery engine (their `gateway/relay/descriptor.py`)
One generic adapter driven by a frozen, versioned per-platform capability descriptor: max message length, length-unit (`chars` | `utf16`), draft-streaming / edit / thread support, markdown dialect, emoji support. The core delivery engine never branches on platform identity; per-platform code shrinks to a connector that fills in the descriptor plus wire I/O. Hermes arrived at this only after years of per-adapter drift (their built-in path still touches ~16 files per new platform) — we get to start with the end state. In Go: a `ChannelDescriptor` struct + one delivery engine; new channel = descriptor + transport, zero core changes.

### 2. Busy-input as three primitives: interrupt / queue / steer
When a message arrives mid-turn: `interrupt` aborts and restarts with the new text; `queue` holds it for the next turn; **`steer` injects it into the running turn** (hermes appends it to the last tool result so the model sees it on the next iteration — no abort, no role-alternation break, in-flight work preserved). Steer auto-demotes to queue when unsafe (subagents active, attachments present). Internal/synthetic events never trigger interrupt semantics — a background job finishing must never abort a live customer exchange. For a front office, steer is the difference between "Alex restarts the answer" and "Alex smoothly absorbs the customer's follow-up".

### 3. Cross-channel continuity: handoff vs mirror are different primitives
- **Handoff** = re-bind the session key to the existing transcript (full continuity when a conversation moves channels), then synthesize an internal event so the agent opens on the new channel.
- **Mirror** = one-way context injection into another session's transcript, role forced to `user` — never `assistant`→`assistant`, which violates strict-alternation LLM APIs.
Conflating these two is the trap; hermes keeps them separate and so should we.

### 4. Scheduler rules (for Herald's follow-ups/digests)
- Fresh agent per scheduled job — no conversation history bleed.
- **Idle-based timeout**, not wall-clock: kill after N minutes of *inactivity*; a productive turn may run long.
- Dead-target registry: a delivery target that hard-fails (chat deleted, bot blocked) is marked dead and skipped — never re-spammed into flood control — self-clearing on next success.
- Scheduled output delivered as a continuable thread: a plain reply resumes the job's context.

### 5. Long-turn chat UX
Typing-indicator heartbeat (~2s refresh; platform indicators expire ~5s) plus ONE progress message edited in place (throttled ~1.5s for flood control), skipped entirely on channels without message editing rather than posting a bubble per step.

### 6. Chunking correctness
Length-unit-agnostic message splitter: pluggable length function (chars vs UTF-16 code units — Telegram/Signal count surrogate pairs), and code-fence-aware splits that close and reopen a fence (with language tag) across message boundaries.

## Anti-patterns (their scar tissue — we design these out)
- 21,808-line gateway run loop; module boundaries by accretion. Enforce package/responsibility budgets from day one.
- Per-platform regex markdown transpilers held upright by hundreds of tests. **Rule: parse markdown to AST once, render N channel backends.**
- Two config-resolution paths with divergent behavior. One config path, ever.

## Invariants worth adopting verbatim
1. Never write assistant→assistant into any transcript (mirror as `user`).
2. A background completion never interrupts a live turn.
3. Self-heal dead delivery targets; never retry into flood control.
4. Every capability flag defaults to the safe/degraded behavior — an unsupported feature degrades, never silently drops.
5. Never mutate past context or rebuild the system prompt mid-conversation (prompt cache is sacred).
