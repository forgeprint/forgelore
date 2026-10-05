# Team trial protocol

A week of ordinary work by two people, to find out whether shared memory
survives contact with a real repository. It is not a benchmark. The question
is whether anything breaks, annoys, or quietly does nothing.

## Before you start

Both people:

```sh
./scripts/build.sh                 # or install a release when one exists
forgelore init --with-git-hook     # once, by one person, then committed
```

The person who runs `init` commits `.forgelore/records/` and
`.forgelore/.gitignore`. Everyone else pulls. The index, the ledger and the
local scope are gitignored and stay on each machine.

Turn the measurement on, in `.forgelore/config.yaml`, committed:

```yaml
measure.ab.control_percent: 20
measure.ab.salt: "trial-1"
```

Twenty percent of sessions get no injections. That is what makes the
difference at the end a measurement rather than an impression.

Wire the agent up. For Claude Code, enable the plugin in `plugin/`; for
anything that speaks MCP, add `forgelore mcp --dir .`. Both people should use
the same wiring, or the two halves of the week are not comparable.

## During the week

Work normally. Specifically, do not go looking for errors to record.

When a fix is worth keeping:

```sh
forgelore review                                    # what the session proposed
forgelore review --accept <fingerprint> --title "…" # becomes a local record
forgelore promote <id>                              # share it with the team
git add .forgelore/records && git commit
```

A title states the answer, not the question: "greet lives in
internal/greeter; import it", never "greet is undefined".

Keep a running note of anything that annoyed you. That list is the most
valuable thing the week produces, and it will not survive to Friday unless it
is written down as it happens.

## What to record each day

One line per person, in a shared file:

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

## What would count as a failure

- A merge conflict in `.forgelore/`.
- A secret reaching a commit.
- A hint that was wrong often enough to be distrusted.
- Nobody promoting anything, because the flow was too much work.

The last one is the likeliest and the least visible. If at the end of the week
the team scope is empty, that is the finding.
