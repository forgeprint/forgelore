# Working on Forgelore

Local-first, team-shared memory for coding agents. One static Go binary, no
runtime dependencies. See `README.md` for what it does.

## Read these first, in this order

1. **`docs/ilerleme.md`** — where the work stands: steps completed, decisions
   taken, open items, what is next. Turkish. This is the notebook between
   sessions; start every session here.
2. **`docs/plan.md`** — the plan and the locked decisions K1 to K15. The source
   of truth for anything architectural. Turkish.
3. **`docs/adr/`** — one record per decision, with the costs named. English.
4. **`docs/record-format.md`** — the normative record format, schema version 1.

## How this project is worked on

These rules come from `docs/plan.md` and are not optional.

- **Ask before each step.** Summarise what you will do, which files you will
  touch and which decisions are needed. Do not write code, create files or run
  commands until the user agrees.
- **Report after each step.** What you did, which tests passed, the state of the
  acceptance criteria, what is still open. Do not start the next step on your
  own.
- **`[SEN]` items belong to the user.** Do not do them. Say plainly what needs
  doing and wait for confirmation. `[CLAUDE]` items are yours, and still subject
  to the first rule.
- **Locked decisions are not re-argued.** If one looks unworkable, say why and
  wait for an answer. Never change it on your own. Changing one means changing
  `docs/plan.md` and its ADR first, then the code.
- **No new dependencies.** The standard library and `modernc.org/sqlite`,
  nothing else — including for tests. If something seems to need one, ask
  first; if it is agreed, it comes with an ADR and is vendored.
- **Never write about an agent's hooks from memory.** Before writing code
  against any coding agent's hook format, event names or config paths, check
  that agent's current official documentation and cite the source in your
  report. Mark anything you could not verify as unverified.
- **Do not guess at an ambiguity.** Ask the user. Do not bury the question in a
  code comment.
- **Update `docs/ilerleme.md`** at the end of every step: what was done, what
  was decided, what is still open. A session that does not update it has lost
  the work for the next one.

The user writes Turkish; answer in Turkish. Public documents and code comments
are English; internal notes (`docs/ilerleme.md`, `docs/plan.md`) are Turkish.

## Forgelore is installed here

This repository uses its own product. `.forgelore/records/` is committed and
holds what earlier sessions learned the hard way; the index, the ledger and
the local scope are gitignored.

```sh
forgelore search <query>      # what is already known
forgelore recall --command "<the command>" --error-file -   # with its output on stdin
forgelore record --type decision --scope team --title "…" --body "…"
forgelore review              # what a session proposed, before it becomes a record
```

When a command here fails, look before fixing. When something turns out to
be worth knowing next time, record it — a title states the answer, not the
question.

One limit worth knowing: the fingerprinter recognises compiler and runtime
diagnostics, not the prose errors CLI tools write. `npm error code E403` and
`Failed to authenticate: …` produce no fingerprint, so they cannot be `fix`
records and are kept as `decision` instead.

## Commands

Every check CI runs is a script, and CI runs nothing else:

```sh
./scripts/ci.sh              # everything below, exactly what CI calls
./scripts/test.sh            # gofmt, go vet, go test
./scripts/crosscheck.sh      # every release target, CGO_ENABLED=0
./scripts/gitleaks.sh        # secret scan, pinned binary
./scripts/build.sh           # host binary into dist/
./scripts/capture-errors.sh  # regenerate testdata/errors
./scripts/capture-agent-events.sh  # regenerate testdata/agents
./scripts/release.sh vX.Y.Z  # build every artifact and SHA256SUMS
./scripts/npm-pack.sh vX.Y.Z  # assemble the npm packages from dist/
./scripts/install.sh         # what a user runs; not needed to develop
./scripts/drift-versions.sh  # has an agent released a version nobody checked?
./scripts/drift-payloads.sh <agent>  # recapture and diff against the corpus
```

`drift-versions.sh` and `drift-payloads.sh` are deliberately **not** in
`ci.sh`: the first needs the network and the second needs a signed-in agent,
and `ci.sh` has to stay runnable offline on any machine. The weekly
`.github/workflows/drift.yml` runs the first one.

`scripts/test.sh` includes a ten thousand record measurement that takes about
twenty seconds. `go test -short ./...` skips it.

## What a fresh machine needs

- **Go 1.26 or newer.** Built with 1.27 so far.
- Nothing else to build or test: dependencies are vendored, so `go build` works
  with no network and no module cache.
- `scripts/gitleaks.sh` downloads a pinned, checksum-verified gitleaks binary
  into `.tools/` on first run. That one needs network.
- `scripts/capture-errors.sh` additionally needs `dotnet`, `node` and `python`.
  It is only needed to regenerate the error corpus, which is already committed.
- `scripts/capture-agent-events.sh` needs the `claude` CLI, logged in. The CLI
  authenticates separately from the desktop app, so `claude` then `/login`
  once. Only needed to add to `testdata/agents`, which is already committed.

## Things that are easy to get wrong here

- `vendor/` is committed on purpose (135 MB on disk, 22 MiB packed). It is
  exempt from line-ending normalisation in `.gitattributes` so the tree stays
  byte-identical to what upstream published. Do not reformat it.
- **Never run `gofmt -w .` from the repository root.** It walks into
  `vendor/` and rewrites doc-comment indentation across the tree. CI will not
  catch it: `scripts/test.sh` excludes `vendor/` from the format check by
  design, so a reformatted vendor tree passes every check and only shows up
  in `git status`. Format a directory you own instead: `gofmt -w ./internal`,
  `gofmt -w ./cmd`. If it has already happened, `git checkout -- vendor/`.
- `.gitattributes` forces LF everywhere else. A shell script checked out with
  CRLF fails on its shebang line, and this project is developed on Windows.
- Shell scripts need their executable bit set explicitly on Windows:
  `git add --chmod=+x scripts/foo.sh`.
- Commits are signed off: `git commit -s`. The DCO application on GitHub checks
  this; CI does not.
- The ledger under `.forgelore/ledger/` is not in the SQLite index and must
  not be moved there. The index is derived and is deleted and rebuilt whenever
  it cannot be read; the ledger is the only copy of what was measured.
- The host binary is 7.1 MB. `cmd/forgelore` links SQLite through
  `internal/store`, which is most of that; `scripts/build.sh` strips symbols
  already, so there is no easy win left to look for.
