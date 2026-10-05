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
- `verified_against` is empty for the Claude Code mapping as shipped. The
  shape of a failed Bash result is not in the documentation, so the mapping
  trusts only the dedicated `PostToolUseFailure` event and treats every other
  finished command as a success. The cost is a missed injection wherever that
  event does not fire; the alternative was guessing a pattern and injecting a
  hint after a command that worked. Silence is the cheaper mistake.
