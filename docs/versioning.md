# Versioning

Three numbers move independently in this project, and conflating them is the
mistake this document exists to prevent.

| What | Where it lives | Now |
|---|---|---|
| The binary | git tag, `forgelore version` | pre-release |
| The record schema | `schema:` in every record file | 1 |
| The mapping format | `mapping_version` in every mapping file | 1 |

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

Agent mappings carry `mapping_version: 1`. A mapping written for a later
format is refused with a sentence naming the problem, never half-read.

This number moves more often than the schema, because it changes whenever an
agent's format needs something the format cannot express. It has already
happened twice:

- a reply shape, when Copilot CLI turned out to read context at the top level
  where Claude Code reads it inside a wrapper;
- a caller-supplied event name, when Copilot CLI turned out not to name its
  events in the payload at all.

Both were additions, so the number stayed at 1: an older Forgelore reading a
mapping that uses neither still works. It moves to 2 the first time an
existing field changes meaning.

Mappings are data and ship inside the binary, but a file on disk overrides
them. A user whose agent has changed can fix their own install with a JSON
edit and no release, which is the point of K5.

## What a release note has to say

- Anything that changed in a command's flags or output.
- Whether the record schema moved, and what to run if it did.
- Whether the mapping format moved, and which agents are affected.
- Which agent versions the mappings are verified against, from
  `docs/compatibility.md`.
