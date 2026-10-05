# Forgelore

Local-first, team-shared memory for coding agents. Forgelore remembers fixes and
dead ends, injects them only when the same error comes back, and measures
whether it actually saves tokens.

A single Go binary with no runtime dependencies. It works with any agent that
can run a shell command.

> **Status: early development.** The record store, error fingerprinting, the
> command line and the measurement ledger work and are tested. Agent
> integrations are not built yet, and there is no release to install — you
> build it yourself. Follow `docs/plan.md` for the roadmap.

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

## Using it

```sh
forgelore init                 # create .forgelore/ in your project
go build ./... 2>err.txt       # something fails

# ask before trying a fix
forgelore recall --command "go build ./..." --error-file err.txt

# once you have solved it, say so
forgelore record --type fix \
  --title "Import the package that defines greet" \
  --command "go build ./..." --error-file err.txt

# and the next time that error appears, in any file, under any go subcommand,
# recall finds it
```

`search <query>` returns ids and titles only; `show <id>` returns one record in
full. Every command takes `--json`. `init` prints the one line to add to your
`AGENTS.md` and does not write the file itself.

## Measuring whether it helps

Every lookup is written to a local ledger, and `forgelore report` shows what
was spent against what it bought:

```sh
forgelore report
```

Two things are compared. **Repeated errors per session** comes from
Forgelore's own ledger and needs nothing from your agent. **Cost per session**
needs the agent to report what it spent; `forgelore usage --help-wiring`
prints how to wire that up for Claude Code. Without it the report shows the
spending and claims no saving.

Turn on the A/B mode to get a control group — the same share of sessions get
no injections at all, and see exactly what a session with no memory sees:

```yaml
# .forgelore/config.yaml
measure.ab.control_percent: 20
```

Differences that are not statistically meaningful are printed as "not
meaningful", with the interval, rather than as a win. The ledger is local and
never leaves the machine.

## Design constraints

The decisions that shape this project — single static binary, pure-Go SQLite,
plain markdown as the source of truth, adapters as data rather than code — are
recorded as ADRs in [`docs/adr/`](docs/adr/) and in [`docs/plan.md`](docs/plan.md).
The record file format is specified in [`docs/record-format.md`](docs/record-format.md).

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

If you work with Claude Code, [CLAUDE.md](CLAUDE.md) has the rules this project
is built under.

## License

[Apache License 2.0](LICENSE).
