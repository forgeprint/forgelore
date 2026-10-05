# ADR-0020: Four canonical events, and mappings in JSON

- Status: Accepted
- Date: 2026-10-05
- Phase: 4

## Context

K5 says agent adapters are data rather than code. Phase 4 had to decide what
that data describes, and the plan's own sketch answered it one way: a
canonical event per agent lifecycle hook, named after Claude Code's —
SessionStart, PromptSubmit, PreTool, PostTool, Stop, PreCompact, SessionEnd.

Checking the official reference on 2026-10-05 made that sketch look wrong in
two ways.

Claude Code fires more than thirty hook events, not seven. Mirroring a
lifecycle means the canonical model grows whenever any agent grows, and every
other adapter has to answer for events its agent does not have. Forgelore
only ever asks four questions of a session.

Separately, a mapping file is nested by nature: an event holds field paths,
and a field path holds alternatives. The restricted YAML of ADR-0017 has no
nested maps, on purpose.

## Decision

The canonical model is four events, named for what Forgelore does about them
and not for when an agent fires them: `session_started`, `command_failed`,
`command_succeeded`, `session_ended`.

Mapping files are JSON, not the ADR-0017 dialect. Each carries its own
`mapping_version` and a `verified_against` naming the agent version whose
real payloads it was tested against.

A field path is a list of dotted paths, and the first that resolves wins.

## Consequences

- An agent with no concept of a session, or one that reports failures through
  a channel Claude Code does not have, maps onto the same four events. The
  adapter absorbs the difference, which is what K5 asked for.
- The project now has two configuration dialects: the restricted YAML for
  records and settings, which humans write, and JSON for mappings, which are
  closer to code. Two is worse than one, and the alternative was nesting in a
  dialect that forbids it.
- Field paths as lists are the hedge against the part of an agent's format
  that documentation does not pin down. A tool result that is a plain string
  today and an object tomorrow is handled by listing both, with no release.
- `verified_against` names the agent version whose captured payloads the
  mapping was checked against, and a test fails if it names one with no
  payloads. The Claude Code mapping shipped with it empty, and the mapping
  trusted only the dedicated `PostToolUseFailure` event rather than guessing
  how a failed command reports itself.

## Update, 2026-10-05: what the captured payloads changed

Holding the mapping against real payloads from Claude Code 2.1.289 found the
documentation-derived version wrong in the one place that mattered.

A failed Bash command has **no `tool_response` field at all**. Its output is
in a top-level `error`, exit code included:

```
"error": "Exit code 1\n# example.com/broken/cmd/app\ncmd/app/main.go:4:2: undefined: greet"
```

The mapping read `tool_response.stderr`, `tool_response.stdout` and
`tool_response`, found none of them, and produced an event with no output.
Every fingerprint would have come back empty and the injection path would
have been silent forever, with no error anywhere to show for it.

A *succeeding* command does have `tool_response`, as an object with `stdout`
and `stderr`. So the two events genuinely differ in shape, which is what the
list-of-paths field format was for; adding `error` to the front of the list
was the whole fix, in data.

Two more things the payloads settled: `is_interrupt` marks a command the user
stopped, which is not a failure worth remembering, so the mapping gained a
`skip_when`; and `scratchpad_dir` is present on all four events, so session
state lives where the agent already cleans up.

The decision this vindicates is not the mapping's content but its format.
Being wrong cost a JSON edit.
