# Team trial protocol

A week of ordinary work by two people, to find out whether shared memory
survives contact with a real repository. It is not a benchmark. The question
is whether anything breaks, annoys, or quietly does nothing.

## Before you start

**Both people**, on their own machine:

```sh
curl -fsSL https://raw.githubusercontent.com/forgeprint/forgelore/main/scripts/install.sh | bash
```

**One person**, once, in the repository:

```sh
forgelore init
git add .forgelore && git commit -m "add forgelore"
```

That commits `.forgelore/records/` and `.forgelore/.gitignore`. The index,
the ledger and the local scope are gitignored and stay on each machine.

**Both people again**, after pulling:

```sh
forgelore init --with-git-hook
```

Running it a second time is safe — it reports what is already there. The
`--with-git-hook` part matters and cannot be shared: a git hook lives in
`.git/hooks/`, which git never commits. Whoever skips this has no check
stopping a secret reaching a commit.

Turn the measurement on, in `.forgelore/config.yaml`, committed:

```yaml
measure.ab.control_percent: 20
measure.ab.salt: "trial-1"
```

One session in five gets no injections. Twenty rather than fifty because
this week cannot be powered to measure anything either way — a bigger
control arm would cost usefulness and buy no significance.

### Wire the agent up

| Agent | Wiring | Measured |
|---|---|---|
| Claude Code | enable the plugin in `plugin/` | yes |
| Copilot CLI | hooks per `docs/compatibility.md` | yes |
| anything via MCP | `forgelore mcp --dir .` | only if the caller passes `session` |
| `forgelore recall` by hand | — | only with `--session` |

A lookup with no session identifier still answers; it just cannot be
attributed to a run, so it counts towards what was spent and drops out of
the per-session comparisons. Through hooks the identifier comes from the
agent and there is nothing to do. Through MCP the model has to pass it, and
whether it reliably does is one of the things this week finds out — check on
day zero rather than on Friday.

Both people should use the same wiring, or the two halves of the week are
not comparable.

### Day zero: prove the wiring works

Do this before any real work. Without it, "nobody promoted anything" at the
end of the week is indistinguishable from "the hooks never fired", and the
trial would have measured nothing while looking like a finding.

```sh
# 1. Break something so the compiler emits a diagnostic — a call to a
#    function that does not exist is enough — and let the agent build it.
# 2. Check the hook saw it:
forgelore report --days 1        # errors looked up should be at least 1
forgelore doctor                 # agents: your agent should say "verified"
```

**Break it in a way that produces a diagnostic.** The first attempt at this
on 2026-10-06 used `go build ./internal/notapackage`, which fails with
`stat …: directory not found` — a sentence with no diagnostic shape, so
Forgelore correctly found nothing and the smoke test looked like a wiring
failure. A missing function, a type error, a syntax error: anything the
compiler reports with a file and a line.

If `report` says nothing was looked up, check that before suspecting the
wiring: run the same command yourself and pipe its output into
`forgelore recall --command "<the command>" --error-file -`. If that finds
no error either, the problem is the error you chose, not the hooks.

If `doctor` says your agent is not verified against a captured payload, the
hooks may run and still match nothing; say so in the write-up rather than
treating the week as evidence about Forgelore. It also reports a version
drift — "verified against 2.1.290, you have 2.1.291" — which is worth
writing down at the start of the week and checking again at the end.

## During the week

Work normally. Specifically, do not go looking for errors to record.

Pull at least once a day. Two people adding records is the thing this trial
is testing, and it is only tested if your trees actually meet.

When a fix is worth keeping:

```sh
forgelore review                                     # what the session proposed
forgelore review --accept <fingerprint> --title "…"  # prints the new record id
forgelore promote <that id>                          # moves it to the team scope
git add .forgelore/records && git commit
```

A title states the answer, not the question: "greet lives in
internal/greeter; import it", never "greet is undefined".

Keep a running note of anything that annoyed you. That list is the most
valuable thing the week produces, and it will not survive to Friday unless it
is written down as it happens. Keep it outside the repository, or the daily
log becomes part of what you are diffing.

## What to record each day

One line per person:

- memories promoted today
- times a hint appeared, and whether it helped
- times a hint appeared and was wrong or irrelevant
- anything that got in the way

## At the end

Both people, on their own machine:

```sh
forgelore report --days 7
forgelore dedupe
forgelore doctor
```

`report` compares the two arms on cost per session and on repeated errors per
session, and says plainly when a difference is not meaningful. With two people
for one week it very likely will not be: a handful of sessions each is not
enough to separate anything, and the report saying so is the correct outcome
rather than a disappointing one.

The numbers worth looking at regardless:

- how many memories exist, and how many were ever injected
- how many fingerprints have more than one record (`dedupe`)
- whether any injected hint was wrong

Cost per session only appears if your agent reports its usage; for Claude
Code that means wiring the status line, which `forgelore usage --help-wiring`
explains. Without it the report shows what was spent and claims no saving,
which is the intended behaviour and not a fault.

## What would count as a failure

- A merge conflict in `.forgelore/`.
- A secret reaching a commit.
- A hint that was wrong often enough to be distrusted.
- Nobody promoting anything, because the flow was too much work.

The last one is the likeliest and the least visible. If at the end of the week
the team scope is empty, that is the finding — provided day zero proved the
wiring worked, which is the only thing that tells the two apart.
