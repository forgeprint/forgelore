# Agent compatibility

What Forgelore supports, and how much of it has been checked rather than
read about.

## Tiers

- **A** — hooks, MCP and the CLI. Memory arrives on its own when a command
  fails, and the agent can also search it.
- **B** — MCP and the CLI. The agent has to ask.
- **C** — the CLI only, pointed at from `AGENTS.md`.

A tier describes what has been **checked against a running agent**, not what
its documentation promises. Every mapping in this project started as a
reading of official docs, and every one of them was wrong somewhere that
would have left Forgelore silent — so a capability nobody has exercised is
listed as not measured, and does not count towards a tier.

## Supported

| Agent | Tier | Verified against | Hooks | MCP | Usage reader |
|---|---|---|---|---|---|
| Claude Code | **A** | 2.1.289 | ✅ 4 captured payloads | ✅ both protocol eras measured | ✅ status line |
| Copilot CLI | **B**, hooks verified | 1.0.91 | ✅ 3 captured payloads | never connected | none known |
| Codex CLI | **unverified** | — | mapping written from docs | never connected | none known |
| Gemini CLI | researched, not started | — | — | — | — |
| Cursor | researched, not started | — | — | — | — |

Copilot CLI is listed as B rather than A on a technicality worth keeping:
its hooks are verified against captured payloads, but nobody has pointed its
MCP client at `forgelore mcp`, so the A in "hooks, MCP and the CLI" is not
earned. Claiming it would be the same mistake this table exists to prevent.

Any agent that can run a shell command is tier C today without doing
anything: `forgelore init` prints the line to add to `AGENTS.md`.

## The MCP server is not agent-specific

"Never connected" in the table above is a statement about that agent's
client, not about Forgelore. `forgelore mcp` serves two protocol revisions
and is exercised by two real clients — Claude Code and the official MCP
inspector — so a client that speaks either revision should work. What has
not been established for Copilot CLI or Codex CLI is whether theirs does,
and nobody should assume it from this table.

One thing is worth knowing before wiring any agent through MCP:
`recall_error` only counts towards `forgelore report` when the caller passes
a `session` argument. Hooks supply the session themselves; an MCP client has
to be persuaded to, and whether a given model reliably does is its own
question.

## Claude Code

Verified against 2.1.289 on 2026-10-05.

Hooks: `SessionStart`, `PostToolUse`, `PostToolUseFailure`, `SessionEnd`,
configured by the plugin in [`plugin/`](../plugin/). A failed Bash command
puts its output in a top-level `error`, not in `tool_response`; a successful
one uses `tool_response` as an object. `is_interrupt` marks a command the
user stopped, and it is ignored.

Measured end to end on 2026-10-05: the plugin was installed from the
Forgeprint marketplace, a recorded fix was injected into five real sessions,
and `forgelore report` counted them. Hook latency was 10 ms at p95.

### A piped command hides its failure

This is the one gap worth knowing about, and no mapping can close it.

A Bash command's exit status reaches Forgelore only through which event
fires: `PostToolUseFailure` for a command that failed, `PostToolUse` for one
that did not. Neither payload carries an exit code. So when the agent writes

```
go build ./... 2>&1 | head -40
```

the pipeline exits with `head`'s status, which is zero. Claude Code is right
to send `PostToolUse`, and Forgelore is right to treat it as a success — but
the build failed, and nothing is recorded or recalled. It happens silently
and looks exactly like an agent that met no errors.

Whether the agent pipes a build is its own habit, not something this project
controls. If it does it often, the fix would be to decide failure from the
output text rather than from the event, the way the Copilot CLI mapping
already does with `failure.output_matches`. That is a change to the mapping
and to what "failed" means for this agent, so it is not made here on the
strength of one observation.

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

A second guess sits beside it. The mapping skips an event carrying
`interrupted` or `is_interrupt`, copied from how Claude Code marks a command
the user stopped — but Codex documents `Interrupt` as an *event*, not a
field, so the skip may never fire and a cancelled command may be remembered
as a failure.

Three attempts to sign the CLI in failed on 2026-10-05 — two browser flows
and a device code — so nothing here has been checked. Assume it is wrong
until a payload says otherwise.

## Copilot CLI

Verified against 1.0.91 on 2026-10-05. Both guesses made from the
documentation were wrong, and in ways that would have left Forgelore silent.

**The payload does not name its event.** There is no `hookEventName` field of
any spelling: a payload carries `sessionId`, `timestamp`, `cwd` and whatever
the event itself adds. The event is known only from which hook entry fired,
so the hook command carries it: `forgelore hook --adapter copilot-cli --event
postToolUse`. The mapping format gained an empty `event_name` meaning "the
caller supplies it".

**A failing command is not a failing tool.** `go build` exiting 1 arrives as
`postToolUse` with `toolResult.resultType: "success"`, and the shell's exit
code appears at the end of the result text:

```
# example.com/broken/cmd/app
cmd/app/main.go:4:2: undefined: greet
<shellId: 0 completed with exit code 1>
```

So the mapping decides failure with `output_matches` on `completed with exit
code [1-9]`. `postToolUseFailure` exists but means the tool itself failed,
which a broken build does not do; no payload for it has been captured.

Context goes back as a top-level `additionalContext` rather than inside a
wrapper — which is why the mapping format gained a `response` section.

### Wiring it up

One entry per event, each passing its own name, in a hooks file:

```json
{
  "version": 1,
  "hooks": {
    "sessionStart":       [{ "type": "command", "bash": "forgelore hook --adapter copilot-cli --event sessionStart",       "timeoutSec": 10 }],
    "postToolUse":        [{ "type": "command", "bash": "forgelore hook --adapter copilot-cli --event postToolUse",        "timeoutSec": 10 }],
    "postToolUseFailure": [{ "type": "command", "bash": "forgelore hook --adapter copilot-cli --event postToolUseFailure", "timeoutSec": 10 }],
    "sessionEnd":         [{ "type": "command", "bash": "forgelore hook --adapter copilot-cli --event sessionEnd",         "timeoutSec": 1 }]
  }
}
```

The documentation gives three places for it: `.github/hooks/*.json` in the
repository, `~/.copilot/hooks/` (or `$COPILOT_HOME/hooks/`), or inline in a
settings file. The capture for this corpus used `$COPILOT_HOME/hooks/`,
because the repository location was not read in the throwaway directory the
script works in — which is not a git repository, so that may be the reason
rather than a general rule. If `.github/hooks` does nothing for you, try the
home location before concluding anything is broken, and
`forgelore report --days 1` is how you tell which happened.

## Gemini CLI and Cursor

Researched on 2026-10-05, not started.

Gemini CLI configures hooks in `.gemini/settings.json` and names its tool
events `BeforeTool` and `AfterTool`. Cursor configures them in
`.cursor/hooks.json` and hangs them off shell execution and file edits —
`beforeShellExecution`, `afterFileEdit`, `afterMCPExecution` — rather than
off a tool result, which is a different enough model to deserve its own
round.

## Keeping this table true

An agent that releases a new version does not announce it to this file.

```sh
./scripts/drift-versions.sh        # has anything shipped past what is verified?
./scripts/drift-payloads.sh <agent>  # did a field actually move? needs the agent signed in
forgelore doctor                   # what is installed here, against what is verified
```

The first runs weekly in CI and opens an issue. The second is the only one
that can tell whether anything is broken, and it has to run on a machine
with the agent signed in.

## Adding an agent

1. `./scripts/capture-agent-events.sh <agent>` in a throwaway directory,
   with that agent signed in. Payloads land in `testdata/agents/`.
2. Write `internal/agent/mappings/<agent>.json`. It is data: event names,
   field paths, failure test, reply shape. No Go changes should be needed,
   and if they are, say so — that is a gap in the mapping format.
3. Leave `verified_against` empty until the captured payloads pass, then set
   it to the version they came from. A test refuses a claim with no corpus
   behind it.
4. Measure which MCP protocol revision the agent's client speaks, and
   whether it passes `session` to `recall_error`. Do not assume either:
   Claude Code speaks two revisions, and which one depends on a flag.
5. Update the table, including the columns you did **not** check.
