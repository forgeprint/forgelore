# Forgelore

Local-first, team-shared memory for coding agents. Forgelore remembers fixes and
dead ends, injects them only when the same error comes back, and measures
whether it actually saves tokens.

A single Go binary with no runtime dependencies. It works with any agent that
can run a shell command.

> **Status: early development.** The repository skeleton and CI are in place;
> the record store, fingerprinting and CLI are not implemented yet. There is no
> release to install. Follow `docs/plan.md` for the roadmap.

## What it does

- **Event-triggered, just-in-time recall.** When a command fails, Forgelore
  fingerprints the error. If that fingerprint has been seen and solved before,
  it injects a one-line hint — the fix, and the approaches already known not to
  work. No match means no tokens spent.
- **Remembers dead ends.** Approaches that were tried and failed leave no trace
  in a codebase, yet repeated error loops burn the most tokens.
- **Staged access.** A small, budgeted index at session start; detail only on
  request. The tool design enforces it: you cannot fetch detail without an id
  from the index.
- **Measured, not claimed.** Every injected byte is recorded in a local ledger.
  An optional A/B mode disables injection for a share of sessions and compares
  token usage, so the savings figure is an observation rather than a promise.

## What it does not do

- No background model calls. Capture is deterministic.
- No network access, no telemetry.
- Never blocks the agent. Hooks are fail-open with a hard timeout.
- Does not duplicate an agent's built-in memory features.

## Design constraints

The decisions that shape this project — single static binary, pure-Go SQLite,
plain markdown as the source of truth, adapters as data rather than code — are
recorded as ADRs in [`docs/adr/`](docs/adr/) and in [`docs/plan.md`](docs/plan.md).

## Development

Every check CI runs is a script in [`scripts/`](scripts/), runnable locally:

```sh
./scripts/test.sh        # gofmt, go vet, go test
./scripts/crosscheck.sh  # cross-compile all targets with CGO_ENABLED=0
./scripts/gitleaks.sh    # secret scan with a pinned, checksum-verified binary
./scripts/ci.sh          # all of the above, exactly what CI calls
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Commits must be signed off under the
[Developer Certificate of Origin](DCO) (`git commit -s`).

## License

[Apache License 2.0](LICENSE).
