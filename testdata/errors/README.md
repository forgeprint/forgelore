# Error samples

Real output from real failing commands, used to calibrate and test error
fingerprinting (phase 2 of `docs/plan.md`).

Every file here was produced by running the command it names, on a machine with
the toolchain version it records. None of it was copied from a web page: a
sample from a search result is an assertion about what a tool prints, and the
whole point of this corpus is to observe what tools actually print.

Regenerate the whole corpus with:

```sh
./scripts/capture-errors.sh          # everything
./scripts/capture-errors.sh python   # one language
```

## Layout

```
testdata/errors/<tool>/<family>/<variant>.txt
```

A **family** is one error. A **variant** is the same error seen differently:
`a` and `b` were captured in workspaces with different names, with the failing
line in a different place, so they differ in exactly the ways a fingerprint has
to ignore.

That gives the tests two claims to check:

- both variants of a family produce the **same** fingerprint (a miss means
  memory stays silent when it should not);
- two different families produce **different** fingerprints (a collision means
  an irrelevant hint gets injected).

One family does not hold up its end. `go/unknown-import` was captured twice and
the two files are byte for byte identical: the message names neither the
package nor a line that moved, so the two captures differ in nothing. It stays
in the corpus, because it is a real error worth extracting, but
`internal/fingerprint` excludes it from the variant comparison — a test that
passed on it would be claiming evidence that is not there. Recapturing it with
a different import path, on a line that moves, is open work and needs the
machine the corpus was made on.

## File format

The restricted YAML dialect of ADR-0017, then the raw output verbatim:

```
---
command: "go build ./..."
exit: 1
family: go/undefined-identifier
tool: "go version go1.27.0 windows/amd64"
platform: windows/amd64
captured: 2026-09-29
---
# alpha
.\main.go:5:14: undefined: greet
```

## Scrubbing

Paths are rewritten so that no account name survives, while the shape of the
path is kept: a Windows user directory stays a Windows user directory, because
that is precisely what normalisation has to deal with. Node prints paths inside
JavaScript string literals with doubled separators, which needs its own rule —
the first version of the script missed it and leaked a real account name into
two files.

Anything added here by hand must be scrubbed the same way.

## What is not here yet

- **Errors from real projects.** These samples are small and deliberate. Real
  failures are messier: CI logs, container builds, dependency resolution, a
  stack trace forty frames deep. They are worth more than anything in this
  directory and have to be collected by hand.
- **Languages with no toolchain on the capture machine**, such as Java and
  Rust. They are outside the plan's stated scope for now. If they are added
  from documentation rather than from a run, they must say so, because an
  unverified sample is a different kind of evidence.
