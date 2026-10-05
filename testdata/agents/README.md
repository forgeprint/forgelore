# Captured agent events

Real hook payloads, captured from a running agent by
`scripts/capture-agent-events.sh`. Documentation says what an agent intends to
send; these files are what it sent.

```
testdata/agents/<agent>/<version>/<Event>-<n>.json
```

`internal/agent` translates every file here with the mapping that ships and
fails if the two have drifted, so a mapping cannot quietly stop matching the
agent it names.

## Scrubbing

The account name is replaced with `dev` and the capture directory with
`/work`, while the shape of each path is kept — a home directory stays a home
directory, because the shape is what a mapping has to deal with. The capture
script checks its own work and refuses to write a file the account name
survived in.

## What is not here yet

`PostToolUse` and `PostToolUseFailure`. Capturing either needs a model turn,
which needs a logged-in `claude` CLI; the session events here came from a run
that had neither. They are the two events the injection path depends on, and
until a payload exists the mapping's treatment of them is an inference from
documentation. `internal/agent` lists the gap in `withoutSamples` so that it
stays visible rather than being forgotten.
