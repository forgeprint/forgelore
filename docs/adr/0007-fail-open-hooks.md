# ADR-0007: Hooks fail open, always

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K7

## Context

A hook sits between a developer and their agent. If it hangs, the agent hangs.
If it exits non-zero at the wrong moment, it can interrupt work that has nothing
to do with memory.

Forgelore is an assistive layer. A missing hint costs a little efficiency. A
blocked agent costs the user's trust, once, permanently.

## Decision

Hook invocations always exit successfully. They run under a hard timeout, they
recover from panics, and any internal error results in injecting nothing and
moving on. Errors are written to a local log, never to the agent's channel.

## Consequences

- A corrupt index, a deleted .forgelore directory, a locked database, a mapping
  file for an agent version we have never seen: every one of these is silent and
  harmless.
- Failures are invisible by default, which is exactly the risk. The doctor
  command exists to surface what the hooks swallowed, and the phase 4 acceptance
  criteria include deliberately breaking the index, the directory and the
  timeout to confirm the agent is unaffected.
- Hook latency is measured (p50 and p95) and reported, because a timeout that is
  never reached is only useful if the normal path is far below it.
- This rule cannot be relaxed for a feature that would be nicer if it could
  block. There is no such feature.
