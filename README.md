# Forgelore

[![CI](https://github.com/forgeprint/forgelore/actions/workflows/ci.yml/badge.svg)](https://github.com/forgeprint/forgelore/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/forgeprint/forgelore)](https://github.com/forgeprint/forgelore/releases/latest)

Local-first, team-shared memory for coding agents. Forgelore remembers fixes and
dead ends, injects them only when the same error comes back, and measures
whether it actually saves tokens.

A single Go binary with no runtime dependencies. It works with any agent that
can run a shell command.

> **Status: early development.** The record store, fingerprinting, the
> command line, the measurement ledger, Claude Code hooks and the MCP server
> all work and are tested. `v0.1.0` is published; the record format and the
> command line may still change before `1.0`. Follow `docs/plan.md` for the
> roadmap and `docs/compatibility.md` for which agents are verified.

## What it looks like

A fix recorded while running `go build` is found later by `go vet`, in a
different file, on a different line. Real output, from `v0.1.1`:

```console
$ go build ./...
# example.com/app/cmd/app
cmd/app/main.go:4:2: undefined: greet

$ forgelore recall --command "go build ./..." --error-file err.txt
1 error(s), 0 with something recorded

cd023fb609411574  undefined: greet
  nothing recorded

$ forgelore record --type fix \
    --title "greet lives in internal/greeter; import it" \
    --command "go build ./..." --error-file err.txt
01M45XT2HS8G3PXDBRTWWV3449  .forgelore/records/01M45XT2HS8G3PXDBRTWWV3449.md
fingerprint cd023fb609411574
```

Later, a different file, a different command:

```console
$ go vet ./...
# example.com/app/internal/svc
vet: internal/svc/svc.go:4:2: undefined: greet

$ forgelore recall --command "go vet ./..." --error-file err2.txt
1 error(s), 1 with something recorded

cd023fb609411574  undefined: greet
  fix       01M45XT2HS8G3PXDBRTWWV3449  greet lives in internal/greeter; import it
```

The fingerprint is the same because the file, the line and the subcommand are
not part of it. The error is.

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

## Installing

```sh
curl -fsSL https://raw.githubusercontent.com/forgeprint/forgelore/main/scripts/install.sh | bash
```

It verifies the download against the release's `SHA256SUMS` and refuses to
install anything that does not match. No sudo, nothing outside your home
directory, no runtime to install alongside it.

To build it yourself instead, `./scripts/build.sh` puts a binary in `dist/`.

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

## With Claude Code

[`plugin/`](plugin/) is a Claude Code plugin that wires the hooks up: a
budgeted index at session start, a hint when a command fails with an error it
has seen, and a proposed fix when a command that was failing starts working.
Nothing is recorded without you accepting it with `forgelore review`.

Every hook exits successfully whatever happens. A broken store, a corrupt
index or a missing directory makes Forgelore silent, never your session
slower or louder.

## With any MCP client

`forgelore mcp` serves the Model Context Protocol on stdin and stdout, so a
client that is not Claude Code can search the same memory:

```sh
claude mcp add --transport stdio forgelore -- forgelore mcp --dir .
```

Four tools: `search`, `get`, `recall_error` and `propose`. `get` only takes an
id that `search` or `recall_error` returned, so no single call can hand over
the whole store, and `propose` writes a candidate for `forgelore review`
rather than a memory.

## Sharing with a team

A memory starts local. Moving one into the team scope is a step somebody
takes on purpose:

```sh
forgelore promote <id>     # local record becomes a team record
forgelore check            # refuses to let a secret be committed
forgelore dedupe           # two records for one error; retire the old one
```

`forgelore init --with-git-hook` installs a pre-commit hook that runs the
check, and refuses to overwrite a hook you already have. A record derived
from content outside the project is marked tainted and needs
`--force-tainted` to be shared.

Records are one file each, named by a ULID, so two people adding memories at
the same time produce two files and no conflict. The index, the ledger and
your local scope are never committed.

[`docs/team-trial.md`](docs/team-trial.md) is a protocol for trying this with
two people for a week.

## Which agents

[`docs/compatibility.md`](docs/compatibility.md) lists what is supported and,
more usefully, what has actually been checked against a running agent rather
than read in its documentation. Claude Code is verified; mappings for Codex
CLI and Copilot CLI are written but unverified. Everything else works through
the CLI today.

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
