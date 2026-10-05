# Versioning

Three numbers move independently in this project, and conflating them is the
mistake this document exists to prevent.

| What | Where it lives | Now |
|---|---|---|
| The binary | git tag, `forgelore version` | pre-release |
| The record schema | `schema:` in every record file | 1 |
| The mapping format | `mapping_version` in every mapping file | 2 |

## The binary

Semantic versioning, from `v0.1.0`.

While the major version is `0`, a minor bump may break a command's flags or
its output. That is what `0` is for, and the compatibility this project does
promise in the meantime is about data, not about the command line: a record
written today is readable by every later version.

Reaching `1.0.0` needs three things that are not true yet:

- the command set settled, with no flag renamed in the previous two releases;
- at least two agents verified against captured payloads, not documentation;
- the measurement having run long enough on real work to say whether the tool
  pays for itself, either way.

A release is cut with `./scripts/release.sh vX.Y.Z`. It refuses a dirty tree,
runs every check, builds all six targets and writes `SHA256SUMS`.

### Which Go builds a release

The published binaries are built by CI with the **minimum** Go this project
claims to support — the version in `go.mod`, at its latest patch — and not
with the newest Go available. Shipping what the README promises is the only
way that promise is ever tested.

The consequence is worth knowing before it confuses somebody: running
`./scripts/release.sh` on a machine with a newer toolchain produces different
bytes from the release, and the checksums will not match. That is the
compiler, not a tampered artifact. To reproduce a release exactly, build with
the Go version in `go.mod`.

CI also runs every check a second time on the newest Go, so a toolchain
release that breaks the build is found here rather than by somebody running
`go install`.

## The record schema

Records carry `schema: 1`. This number changes only when a field changes
meaning or disappears — not when one is added, because an unknown field is
preserved and written back untouched (ADR-0011).

A reader that meets a schema newer than it understands skips the file and
says so through `doctor`, rather than guessing. That is why the number exists
at all: a newer Forgelore in a team is normal, and the older one has to fail
visibly instead of silently misreading.

There is no automatic migration, and there is not meant to be. Records are
markdown files in git (ADR-0009): a schema change comes with a note in the
release saying what to run, and the run is an ordinary commit somebody
reviews.

## The mapping format

Agent mappings carry a `mapping_version`, and the current format is **2**. A
mapping written for a later format is refused with a sentence naming the
problem, never half-read.

This number moves more often than the schema, because it changes whenever an
agent's format needs something the format cannot express. It has happened
three times:

- a reply shape, when Copilot CLI turned out to read context at the top level
  where Claude Code reads it inside a wrapper;
- a caller-supplied event name, when Copilot CLI turned out not to name its
  events in the payload at all;
- `output_has_diagnostic`, when a build piped through `head` turned out to
  exit zero and leave nothing in the payload to read (ADR-0022).

The first two left the number at 1. The third moved it to 2, and the
difference is not that one was an addition and the others were not — all
three were. The question is narrower:

> Can an older Forgelore read this file and still be **right**?

For a reply shape or an event name, yes: a mapping that does not use them
behaves identically on an old binary. For `output_has_diagnostic`, no. An
older binary does not know the field, drops it, and calls every failure a
success — silently, which is the one outcome this project spends its effort
avoiding. So the file says 2, and that binary refuses it out loud instead.

A mapping only carries the version it needs: `copilot-cli.json` and
`codex-cli.json` are still 1, and a Forgelore from before this change reads
them.

Setting `output_has_diagnostic` in a file that claims version 1 is refused
too, for the same reason — otherwise the number would be a label rather than
a guarantee.

The npm packages carry the binary's version with the leading `v` removed,
and the wrapper and the six platform packages always move together: the
wrapper pins its dependencies exactly, so it can never be republished alone
(ADR-0024).

Mappings are data and ship inside the binary, but a file on disk overrides
them. A user whose agent has changed can fix their own install with a JSON
edit and no release, which is the point of K5.

## What a release note has to say

- Anything that changed in a command's flags or output.
- Whether the record schema moved, and what to run if it did.
- Whether the mapping format moved, and which agents are affected.
- Which agent versions the mappings are verified against, from
  `docs/compatibility.md`.
