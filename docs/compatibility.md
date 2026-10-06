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
| Claude Code | **A** | 2.1.290 | ✅ two corpora, 2.1.289 and 2.1.290 | ✅ both protocol eras measured | ✅ status line |
| Copilot CLI | **B**, hooks verified | 1.0.91 | ✅ 3 captured payloads | never connected | none known |
| Codex CLI | **unverified** | — | mapping written from docs | never connected | none known |
| Gemini CLI | **B**, hooks verified | 0.62.0 | ✅ 4 captured payloads | never connected | none known |
| Cursor | **B**, hooks verified | 2026.10.01 | ✅ 6 captured payloads | never connected | none known |

Gemini CLI and Cursor are B for the same reason as Copilot CLI: their
hooks are verified, but nobody has pointed their MCP clients at
`forgelore mcp`.

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

Verified against 2.1.289 on 2026-10-05 and against 2.1.290 on 2026-10-06.
Both corpora are kept and both are replayed, because the difference between
them is the point.

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

**This is handled.** The `PostToolUse` mapping carries
`failure.output_has_diagnostic`, which asks the fingerprinter whether the
output contains an error it recognises and treats the event as a failure
when it does — so a piped build is recalled like any other. See
[ADR-0022](adr/0022-a-failure-test-that-is-not-data.md) for what that costs:
a command that succeeds while printing error-shaped text, `cat build.log`
for instance, is now read as a failure. Nothing is written to memory either
way; the visible effect is a lookup that probably misses.

### What changed in 2.1.290

`scratchpad_dir` is **gone from every event**. In 2.1.289 all four carried
it; in 2.1.290 none do. Nothing breaks — session state falls back to the
store's own cache, which was always the documented behaviour for a session
without a scratchpad — but the fallback is now the ordinary path rather
than the corner case this project twice assumed it was not.

The failure payload is otherwise unchanged: a failed Bash command still
arrives as `PostToolUseFailure` with a top-level `error` reading
`Exit code 1\n…`, and `is_interrupt` still marks a stopped command.

`drift-payloads.sh` found this, on its first real use, by comparing field
paths rather than values. It also reported `PostToolUseFailure: no longer
captured`, which was **not** a regression: the model had written
`go build ./... 2>&1 | head -40`, so the command genuinely succeeded. The
capture prompt now forbids pipes and redirection, because that mistake has
cost three investigations.

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

Signing in succeeded on 2026-10-06 and a session ran, but **no hook payload
was produced**, so nothing here is checked against one. Assume it is wrong
until a payload says otherwise.

What that attempt did settle, by reading the 0.160.0 binary rather than any
documentation — `openai/codex` has no hooks page in its `docs/`:

- the event names are right: `PreToolUse`, `PermissionRequest`,
  `PostToolUse`, `PreCompact`, `PostCompact`, `SessionStart`, `SessionEnd`,
  `UserPromptSubmit`, `SubagentStart`, `SubagentStop`, `Stop`, `Interrupt`;
- the payload is Claude-shaped: `session_id`, `transcript_path`, `cwd`,
  `hook_event_name`, `permission_mode`, `turn_id`, `model`, `reason`,
  `tool_input`, `stop_hook_active`, replying through `hookSpecificOutput`;
- the `hooks` feature flag is on by default, so forcing it is unnecessary.

The mapping was rewritten against those schemas, and they contradicted
three things it had been asserting from documentation:

- **there is no top-level `error`**, so the failure test it used could
  never have fired. Failure is now decided with `output_has_diagnostic`
  ([ADR-0022](adr/0022-a-failure-test-that-is-not-data.md)), because
  nothing in the schema constrains the inside of `tool_response`;
- **there is no `tool_output`**; the result is `tool_response`;
- **there is no `interrupted` field**, and `Interrupt` is an event of its
  own, so the skip the mapping carried would never have fired either. A
  cancelled command whose output holds no diagnostic produces nothing
  anyway.

And one mistake that would have been silent: the mapping replied at a
top-level `additionalContext`. Codex reads it at
`hookSpecificOutput.additionalContext`, beside `hookSpecificOutput.hookEventName`,
exactly where Claude Code does. Every injected hint would have gone
somewhere Codex never looks, and nothing would have errored.

What is still unknown is where the configuration goes and in what shape.
`$CODEX_HOME/hooks.json` with top-level event keys fires nothing, and a
`{"hooks": {…}}` wrapper made the session hang until it was killed — which
suggests the file is read and something in it blocks. Until that is
settled, nothing here has met a real payload.

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

## Gemini CLI

Verified against 0.62.0 on 2026-10-06. The mapping started as a reading of
the official hooks reference at that tag, and the captured payloads
contradicted it in the usual place.

Hooks live in `settings.json` under `hooks`, one array per event, each entry
carrying an optional `matcher` — a **regex** for tool events — and a list of
`{type, command, timeout}`. `timeout` is **milliseconds** here, where Claude
Code counts seconds.

Every payload carries `session_id`, `transcript_path`, `cwd`,
`hook_event_name` and `timestamp`. There is no scratchpad directory, so
session state falls back to the store's own cache.

A shell command arrives as `AfterTool` with `tool_name: "run_shell_command"`
and the command in `tool_input.command`. The result is in `tool_response`,
which holds `llmContent`, `returnDisplay` and an **optional** `error`.

**The documented `error` field does not fire for a failing command.** A
`go build` that exits 1 arrives with `tool_response` holding only
`llmContent` and `returnDisplay`. A mapping that tested `error` alone would
have treated every failing build as a success.

What the payload does carry is the exit code, as a line inside `llmContent`:

```
<untrusted_context>
Output: # example.com/broken/cmd/app
cmd/app/main.go:4:2: undefined: greet
Exit Code: 1
Process Group PGID: 72958
</untrusted_context>
```

A command that succeeds has no `Exit Code` line at all. So failure is
decided by `output_matches` on `Exit Code: [1-9]`, the same shape Copilot
CLI needed, with `field_present: tool_response.error` kept beside it for
whatever does set that field.

`llmContent` is used for the output rather than the cleaner
`returnDisplay`, because it is the one that carries the exit code. The
wrapper, the `Output:` prefix and the per-run process group id do not reach
the fingerprint — the same error reported by Gemini and by Claude Code
hashes to the same value, and a test holds that.

One thing is unresolved: a command the user interrupts exits 130, which
matches `[1-9]`, and nothing in the payload distinguishes it. Claude Code
marks those with `is_interrupt`; Gemini documents no equivalent. A
cancelled command may be remembered as a failure.

Context goes back as `hookSpecificOutput.additionalContext`, appended to the
tool result. Exit code 2 blocks; this project never uses it.

### Getting a capture to run at all

Two obstacles, neither of them about hooks, both solved inside
`scripts/capture-agent-events.sh` without touching the user's own
configuration.

**Signing in with Google is not enough.** The session dies before any tool
runs:

```
IneligibleTierError: This client is no longer supported for Gemini Code
Assist for individuals.
```

A capture needs `GEMINI_API_KEY` (from Google AI Studio) or Vertex AI. And
setting the variable is not enough either, because the CLI reads the chosen
method from `security.auth.selectedType` in its **merged** settings: a
machine that once signed in with Google keeps using that. The capture
writes `"selectedType": "gemini-api-key"` into the throwaway workspace's
own `settings.json` when `GEMINI_API_KEY` is set, which overrides it for
that session only.

### Two things the documentation gets wrong

Both were found by running it, which is the only reason they are here.

**Folder trust is on by default.** `docs/cli/trusted-folders.md` says the
feature is "disabled by default"; the code reads
`settings.security?.folderTrust?.enabled ?? true`. An untrusted folder makes
the CLI ignore the workspace `settings.json` outright — so the hooks never
load — and quietly downgrades `--yolo` back to prompting. A throwaway
capture directory is never trusted, so the capture profile sets
`GEMINI_CLI_TRUST_WORKSPACE=true` rather than writing to the user's own
`~/.gemini/trustedFolders.json`.

**`run_shell_command` requires confirmation**, which a non-interactive
capture cannot give, hence `--yolo`.

## Cursor

Verified against `cursor-agent` 2026.10.01 on 2026-10-06.

An earlier note here said Cursor's event model was "different enough to
deserve its own round". That was wrong: it fires `sessionStart`,
`postToolUse`, `postToolUseFailure` and `sessionEnd`, which are the four
canonical events exactly.

Hooks go in `.cursor/hooks.json` — project, or `~/.cursor/hooks.json` —
shaped `{"version": 1, "hooks": {…}}`, the same family as Copilot CLI's.

**One command fires two events.** A failing `go build` produces
`afterShellExecution`, carrying the command and its output, *and*
`postToolUseFailure`, carrying the same failure with `tool_name`, `cwd`,
`error_message`, `failure_type` and `is_interrupt`. Only the tool events
are mapped; mapping both would look one error up twice. The duplicate
payloads stay in the corpus because the duplication is the finding.

`postToolUse` returns the result as a **JSON string**:

```json
"tool_output": "{\"output\":\"go version go1.27.1 darwin/arm64\\n\",\"exitCode\":0}"
```

So the exit code is readable, and failure is decided with `output_matches`
on `"exitCode":[1-9]` rather than by guessing from the output's shape. The
cost of that string is real though: a masked failure's text arrives with
its newlines escaped, as one line, so the fingerprinter finds nothing in
it. For Cursor a piped build is invisible, and no mapping can fix that.

**Context goes back as a top-level `additional_context`** — snake_case,
unlike every other agent here. This one was checked against a running
agent rather than read: a hook returned a probe token and the model
repeated it verbatim. Codex's documented reply path turned out to be
wrong, and would have been silent, so the documented path is no longer
taken on trust.

Session events carry no `cwd`, only `workspace_roots`, which is a list.
Nothing needs it — the event's working directory is informational — but it
is why the contract test asks for one on command events only.

Every payload carries `user_email`. The capture scrubs email addresses by
shape, and a test refuses any address in the corpus that is not the
placeholder.

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
