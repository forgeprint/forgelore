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
`forgelore-windows-x64` and so on, each holding only the binary, plus a
`forgelore` wrapper that depends on all six as optional dependencies.

The wrapper is one file. It resolves
`forgelore-${platform}-${process.arch}`, where `platform` is
`process.platform` with `win32` spelled `windows`, and runs the binary with
stdio inherited — inherited rather than piped because `forgelore mcp` speaks
a protocol on stdin and stdout, and anything in between would have to stay
correct about framing forever.

`scripts/npm-pack.sh` assembles the packages and writes the order they must
go out in to `dist/npm/PUBLISH_ORDER`.

Publishing runs in `.github/workflows/npm.yml`, on `release: published` —
after a person has read the drafted release and pressed publish, not when
the tag is pushed. It authenticates with npm **trusted publishing** over
OIDC, so there is no npm token anywhere, and npm generates provenance for
each package on its own. The binaries are downloaded from the release
rather than rebuilt, so what npm ships and what the release attests cannot
drift apart.

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

**Publishing moved into CI after two releases of doing it by hand**, and
both of those releases were damaged by it: `0.1.4` went out with the
wrapper published *before* its dependencies, and carrying a development
build. The ordering is now read from a file the packing script wrote, and
the bytes come from the release.

Trusted publishing removes the automation token this would otherwise need —
nothing long-lived exists to leak — and brings npm provenance with it. It
costs a one-time configuration **per package** on npmjs.com, seven of them,
each naming the repository and `npm.yml`, each with `npm publish` allowed
rather than only `npm stage publish`. It also pins the floor: npm 11.5.1 and
Node 22.14, and Node 22 ships npm 10.x, so the workflow upgrades npm before
it publishes.

`forgelore@0.1.4` and `0.1.5` were published by hand and have no
provenance. Nothing retroactive fixes that.

**Node is not a dependency of Forgelore.** It is a dependency of this one
packaging path. `npm-pack.sh` is not in `ci.sh`, nothing in the Go build
knows about it, and a user who installs with `install.sh` never meets it.

**A user who runs `npm install --no-optional` gets a wrapper with nothing
behind it.** The wrapper says so in a sentence and points at the direct
install, rather than failing with a module resolution stack trace.

**The Windows packages are named `windows`, not `win32`.** npm's spam
detection refused to create `forgelore-win32-x64` with a 403, twice, on two
separate days, while the other four went through in the same burst; under
the new name it published. The `os` field inside still says `win32`,
because npm matches that against `process.platform` and nothing else, so
the name and the field disagree on purpose and the wrapper maps between
them. Worth knowing before anyone "fixes" the inconsistency.
