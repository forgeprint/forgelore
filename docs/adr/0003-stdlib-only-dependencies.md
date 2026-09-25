# ADR-0003: The standard library, one dependency, all vendored

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K3

## Context

Forgelore is security-relevant: it reads a project's error output, writes files
into the repository, and feeds content back into an agent's context. Every
dependency is code with that same access, from an author the user never chose.

It also has to still build in five years. A module that is deleted, renamed, or
has its tags moved should not be able to break the build.

## Decision

The standard library plus modernc.org/sqlite (ADR-0002), and nothing else.
Everything is vendored, so the repository contains every line of code that goes
into the binary.

Adding a dependency is a decision rather than a convenience: it needs an issue
first, then an ADR explaining why the standard library is insufficient.

## Consequences

- We write things other projects would import: CLI argument handling, a
  restricted frontmatter parser, ULID generation, a JSON-RPC loop, a config
  reader. Each is small, testable and well understood.
- The full supply chain can be reviewed, and go mod vendor keeps it visible in
  the diff.
- The cost is real: more of our own code to maintain and test, and the
  occasional rough edge where a popular library would be smoother.
- Test-only dependencies get no exemption. The standard testing package is
  enough.
