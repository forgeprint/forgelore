# Agent compatibility

What Forgelore supports, and how much of it has been checked rather than
read about.

## Tiers

- **A** — hooks, MCP and the CLI. Memory arrives on its own when a command
  fails, and the agent can also search it.
- **B** — MCP and the CLI. The agent has to ask.
- **C** — the CLI only, pointed at from `AGENTS.md`.

A tier is only claimed once captured payloads from that agent back it. An
agent whose mapping was written from documentation alone is listed as
**unverified**, whatever the documentation promises, because the Claude Code
mapping was written that way and was wrong in the one place that mattered:
a failed command has no `tool_response` field at all, and the injection path
would have been silent forever.

## Supported

| Agent | Tier | Verified against | Hooks | MCP | Usage reader |
|---|---|---|---|---|---|
| Claude Code | **A** | 2.1.289 | ✅ 4 captured payloads | ✅ both protocol eras measured | ✅ status line |
| Codex CLI | A, **unverified** | — | mapping written from docs | not measured | none known |
| Copilot CLI | A, **unverified** | — | mapping written from docs | not measured | none known |
| Gemini CLI | researched, not started | — | — | — | — |
| Cursor | researched, not started | — | — | — | — |

Every other agent is tier C today: `forgelore` is a shell command, and
`forgelore init` prints the line to add to `AGENTS.md`.

## Claude Code

Verified against 2.1.289 on 2026-10-05.

Hooks: `SessionStart`, `PostToolUse`, `PostToolUseFailure`, `SessionEnd`,
configured by the plugin in [`plugin/`](../plugin/). A failed Bash command
puts its output in a top-level `error`, not in `tool_response`; a successful
one uses `tool_response` as an object. `is_interrupt` marks a command the
user stopped, and it is ignored.

MCP: both protocol eras are served, because the client speaks either one
depending on a feature flag. Its v2 runtime opens with `server/discover` at
`2026-07-28`; its v1 runtime opens with `initialize` at `2025-11-25`. See
[ADR-0021](adr/0021-mcp-target-versions.md).

Usage: read from the status line payload, which is the only documented place
a cumulative per-session figure appears. See
[ADR-0019](adr/0019-cost-in-dollars-is-the-comparison.md).

## Codex CLI

Mapping written from the official configuration reference on 2026-10-05, and
**not verified against a single real payload**.

Hooks exist and cover twelve events including `PreToolUse`, `PostToolUse`,
`SessionStart` and `SessionEnd`, in `~/.codex/hooks.json` or inline in
`config.toml`. Two things make it unlike Claude Code: hooks are **off by
default** behind `features.hooks`, and project-scoped configuration is only
read for a project the user has trusted.

No dedicated tool-failure event is documented, so the mapping treats
`PostToolUse` as the failure when an `error` field is present and as a
success otherwise. That guess is the first thing a captured payload should
settle.

## Copilot CLI

Mapping written from the official hooks reference on 2026-10-05, and **not
verified against a single real payload**.

Events are camelCase: `sessionStart`, `postToolUse`, `postToolUseFailure`,
`sessionEnd`. The documentation also describes a snake_case "VS Code
compatible" form, so the mapping lists both spellings for every field and
takes whichever resolves.

A failing tool reports `error` as a string, the same shape Claude Code turned
out to use. Context goes back as a top-level `additionalContext` rather than
inside a wrapper — which is why the mapping format gained a `response`
section in this phase.

## Gemini CLI and Cursor

Researched on 2026-10-05, not started.

Gemini CLI configures hooks in `.gemini/settings.json` and names its tool
events `BeforeTool` and `AfterTool`. Cursor configures them in
`.cursor/hooks.json` and hangs them off shell execution and file edits —
`beforeShellExecution`, `afterFileEdit`, `afterMCPExecution` — rather than
off a tool result, which is a different enough model to deserve its own
round.

## Adding an agent

1. `./scripts/capture-agent-events.sh <agent>` in a throwaway directory,
   with that agent signed in. Payloads land in `testdata/agents/`.
2. Write `internal/agent/mappings/<agent>.json`. It is data: event names,
   field paths, failure test, reply shape. No Go changes should be needed,
   and if they are, say so — that is a gap in the mapping format.
3. Leave `verified_against` empty until the captured payloads pass, then set
   it to the version they came from. A test refuses a claim with no corpus
   behind it.
4. Measure which MCP protocol revision the agent's client speaks. Do not
   assume: Claude Code speaks two, and which one depends on a flag.
