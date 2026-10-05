# ADR-0023: Build provenance, not a detached signature

- Status: Accepted
- Date: 2026-10-05
- Phase: post-release

## Context

The plan left signing open: "decide if wanted". Phase 8 deferred it for a
concrete reason rather than a vague one — releases were built on a laptop,
and keyless signing needs the identity of the workflow that produced the
artifacts. Decision 1 of that phase moved the build into tag-triggered CI
and said this would be cheap afterwards. It is now time to spend it.

`SHA256SUMS` already answers "are these the bytes that were published".
What it cannot answer is "who published them": the checksum file sits in
the same release as the binaries, so anyone who can replace one can replace
both. Neither can `install.sh`, which reads the checksums out of that same
release.

Two ways to close it:

- **cosign `sign-blob`.** Sign `SHA256SUMS` keyless, upload a `.sig` and a
  `.pem` beside it. Verification works without GitHub and the signature is
  anchored in the public Rekor log. Costs an installer step, two more
  release assets, and a tool the verifier has to have.
- **GitHub artifact attestations.** One step,
  `actions/attest-build-provenance`, producing SLSA provenance stored by
  GitHub rather than as release assets. Verification is
  `gh attestation verify`. Costs a dependency on GitHub for the record, and
  on `gh` for the easy path.

## Decision

Artifact attestations, with `actions/attest-build-provenance` pinned to a
commit SHA like every other action here. The step runs **before** the
release is created, so a failure means there is no release to correct
afterwards.

The subject is `dist/*` — every binary and `SHA256SUMS` — because a
checksum file nobody can attribute proves nothing on its own.

`gh attestation verify forgelore_linux_amd64 -R forgeprint/forgelore` is the
command, and it is in the README and in SECURITY.md rather than only here.

## Consequences

**The provenance lives with GitHub, not in the release.** Someone who has
the binaries but not the network cannot check them. `SHA256SUMS` still
works offline for integrity, which is the weaker but useful half, and that
is the trade accepted here.

**Verification is easy only with `gh` installed.** The attestation is a
Sigstore bundle and other tools can read it, but the one-line instruction
assumes `gh`. A user without it is no worse off than before this ADR.

**The workflow now holds `id-token: write` and `attestations: write`.** That
is a real widening of what the release job can do, and the reason the job
does nothing else: it checks out, builds, attests, drafts. Nothing in it
runs code from a pull request.

**npm packages are not covered.** They are published by hand (ADR-0024), and
npm provenance requires publishing from CI. Said plainly in that ADR rather
than left for someone to assume.

**This is not code signing.** macOS Gatekeeper and Windows SmartScreen know
nothing about Sigstore, and a user who downloads a binary still meets the
operating system's own warning. Fixing that needs an Apple Developer ID and
an Authenticode certificate, which are paid, named identities and a separate
decision.
