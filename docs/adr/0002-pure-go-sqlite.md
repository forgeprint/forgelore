# ADR-0002: modernc.org/sqlite for the derived index

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K2

## Context

Recall has to be fast enough to run inside a hook: a fingerprint lookup on every
failed command, and full-text search over records. That needs an index, and
SQLite with FTS5 is the obvious fit. It is a single file, needs no server, and
its full-text support is mature.

The usual Go binding, mattn/go-sqlite3, wraps the C library through cgo. That
contradicts ADR-0001: it needs a C compiler to build, and it turns
cross-compiling to six targets into an exercise in toolchain management.

## Decision

Use modernc.org/sqlite, a pure-Go translation of SQLite that includes FTS5. It
is the only dependency outside the standard library (ADR-0003).

## Consequences

- Cross-compilation stays a plain GOOS/GOARCH go build.
- It is slower than the C library. That is acceptable here: the index is derived
  data over one project's records, measured in thousands of rows rather than
  millions. Phase 1 measures index build and query time on a synthetic
  10,000-record repository and puts the numbers in the report. If they
  disappoint, this ADR is revisited with evidence rather than intuition.
- The package is large and tracks upstream SQLite on its own schedule. It is
  vendored, so an upstream change cannot break a build without a deliberate
  update.
- A corrupt index is deleted and rebuilt from the markdown files. Nothing is
  lost, because nothing authoritative lives in it (ADR-0009).
