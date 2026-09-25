# ADR-0001: Go, compiled to a single static binary

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K1

## Context

Forgelore runs as a hook. A coding agent invokes it on events that happen many
times per session, and the user waits for it every time. It also has to be
installable by someone who wants a memory layer, not a new toolchain.

A runtime dependency would be fatal to both. A Python or Node implementation
means the user's machine needs that runtime at a compatible version, and process
startup costs tens to hundreds of milliseconds before any work begins.

## Decision

Go, built with CGO_ENABLED=0 into a single static binary per platform. Release
targets: linux, darwin and windows on amd64 and arm64.

The Go 1 compatibility promise means a build that works today keeps working.
go.mod declares go 1.26, one release behind the current toolchain, so a
contributor slightly behind on updates can still build.

## Consequences

- Installation is: download one file, verify its checksum, run it. No runtime,
  no C compiler, no package manager.
- Hook startup is a process exec and nothing more.
- CGO_ENABLED=0 is a hard constraint, not a preference. It rules out any
  dependency that needs cgo (see ADR-0002), and scripts/crosscheck.sh builds
  every target on every pull request so a cgo dependency cannot creep in
  unnoticed.
- Binaries are larger than a script, and every release means shipping six of
  them with checksums.
