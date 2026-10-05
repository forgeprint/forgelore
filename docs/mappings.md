# Updating an agent mapping

An agent changes its hook format and Forgelore goes quiet. Fixing that is a
JSON edit, not a code change, and this page is how.

If you find yourself needing to change Go code to support an agent, that is a
gap in the mapping format and worth saying so in an issue — the whole point
of keeping adapters as data (ADR-0005) is that this page is enough.

## Where things are

```
internal/agent/mappings/<agent>.json   the mapping, compiled into the binary
testdata/agents/<agent>/<version>/     real payloads that agent sent
```

A mapping file on disk overrides the built-in one:

```sh
forgelore hook --adapter claude-code --mapping ./my-fix.json
```

So you can fix your own install today and send the change afterwards.

## Finding out what actually changed

Never guess from documentation. Every mapping in this repository was first
written from an agent's official docs, and so far **every one of them has
been wrong** in at least one way that would have left Forgelore silent:

- Claude Code puts a failed command's output in a top-level `error`, not in
  the documented `tool_response`.
- Copilot CLI sends no event name in its payload at all, and reports a
  command that exited non-zero as a *successful* tool call with the exit code
  buried in the result text.

With the agent installed and signed in:

```sh
./scripts/capture-agent-events.sh <agent>   # writes testdata/agents/<agent>/<version>/
./scripts/drift-payloads.sh <agent>         # what moved since the last capture
```

`drift-payloads.sh` compares field paths rather than values, because a
session id differs on every run and a missing field is the whole story.

## The parts of a mapping

```json
{
  "mapping_version": 1,
  "agent": "example-cli",
  "verified_against": "",
  "common": { "event_name": "", "session": ["sessionId", "session_id"],
              "cwd": "cwd", "scratchpad": "" },
  "events": [ ... ],
  "response": { "context_path": "additionalContext" }
}
```

**`verified_against`** is the agent version whose captured payloads this
mapping passes against. Leave it empty until that is true; a test fails if it
names a version with no payloads in `testdata/agents`.

**Field paths are lists, and the first that resolves wins.** This is the
hedge against a format that changes: list the old spelling and the new one
and the mapping keeps working across the change.

```json
"output": ["error", "tool_response.stderr", "tool_response"]
```

**`common.event_name` may be empty**, meaning the payload does not say which
event it is and the caller will. The hook then carries it:
`forgelore hook --adapter copilot-cli --event postToolUse`.

**Deciding that a command failed** has three forms, and the right one depends
on what the agent gives you:

```json
"failure": { "always": true }                                  // a dedicated failure event
"failure": { "field_present": "error" }                        // a field appears
"failure": { "output_matches": "completed with exit code [1-9]" }  // the output says so
```

**`skip_when`** names a field that, when present and not `false`, means the
event is not worth acting on — `is_interrupt` for a command somebody stopped
with Ctrl-C.

**`response.context_path`** is where the agent reads context back, as dotted
keys. Claude Code wants `hookSpecificOutput.additionalContext`; Copilot CLI
wants `additionalContext`. `event_name_path` echoes the agent's own event
name back when it insists on seeing it.

## Checking your change

```sh
go test ./internal/agent
```

The contract test replays every captured payload through the mapping that
ships for it. Break a field path and it fails by name, which is the point:

```
PostToolUse-1.json: the command did not resolve
```

## Sending it

A mapping change is a small pull request:

1. the edited `internal/agent/mappings/<agent>.json`;
2. the captured payloads under `testdata/agents/<agent>/<version>/`, scrubbed
   — the capture script replaces your account name and refuses to write a
   file it survived in;
3. `verified_against` set to the version those payloads came from;
4. a line in `docs/compatibility.md`.

Say in the pull request what the agent actually sent, not what its
documentation says. The next person reading the mapping needs to know which
of the two they are looking at.
