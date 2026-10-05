# ADR-0024: npm, with one package per platform

- Status: Accepted
- Date: 2026-10-05
- Phase: post-release

## Context

The plan asked for an optional npm wrapper that "only downloads the right
binary and verifies its checksum, nothing else". Phase 8 deferred it:
`install.sh` already covers the need, and npm is a new publishing surface.

What changed is reach rather than need. Every agent this project targets
lives in a JavaScript install story — Claude Code, Copilot CLI and Codex CLI
are all reached through npm — so `npm install -g forgelore` is the idiom the
audience already has in its fingers.

Two shapes:

- **A postinstall script that downloads.** One package. Literally what the
  plan describes. But `npm install --ignore-scripts` is now common and in
  some setups the default, and under it the package installs "successfully"
  with no binary. It also needs network at install time beyond the registry,
  which fails behind a registry mirror.
- **One package per platform, pulled in by `optionalDependencies`.** npm
  picks the right one from the `os` and `cpu` fields and skips the rest. No
  script, no download, no network beyond the registry. This is what esbuild
  and swc do. Costs seven published packages per release instead of one.

## Decision

One package per platform: `forgelore-linux-x64`, `forgelore-darwin-arm64`,
`forgelore-win32-x64` and so on, each holding only the binary, plus a
`forgelore` wrapper that depends on all six as optional dependencies.

The wrapper is one file. It resolves
`forgelore-${process.platform}-${process.arch}` and runs the binary with
stdio inherited — inherited rather than piped because `forgelore mcp` speaks
a protocol on stdin and stdout, and anything in between would have to stay
correct about framing forever.

`scripts/npm-pack.sh` assembles the packages from what `release.sh` built.
It does not publish, for the same reason `release.sh` does not: a release
nobody looked at is a release nobody checked.

## Consequences

**The plan's wording is not what was built.** It said download and verify a
checksum. There is no download and no checksum to verify, because the
registry delivers the binary as package content and npm checks its integrity
hash. The plan's intent — nothing clever at install time — is better served
this way than by the mechanism it named.

**Seven packages are published per release instead of one**, and in order:
the platform packages first, because a wrapper whose exact-version
dependencies are not on the registry yet installs with no binary at all.
`npm-pack.sh` prints the two commands in that order.

**The version is the binary's version without the leading `v`.** The
wrapper and the platform packages always move together; a wrapper may never
be republished alone, because its dependencies are pinned exactly.

**Publishing by hand means no npm provenance.** That badge requires
publishing from a workflow. The GitHub release is attested (ADR-0023) and
the npm packages are not, which is a real gap and is written down rather
than papered over. Wiring the publish into CI needs an automation token in
the repository's secrets, and is the obvious next step if npm turns out to
be used.

**Node is not a dependency of Forgelore.** It is a dependency of this one
packaging path. `npm-pack.sh` is not in `ci.sh`, nothing in the Go build
knows about it, and a user who installs with `install.sh` never meets it.

**A user who runs `npm install --no-optional` gets a wrapper with nothing
behind it.** The wrapper says so in a sentence and points at the direct
install, rather than failing with a module resolution stack trace.
