# ADR-0008: No background model calls, no daemon

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K8

## Context

The tempting design is to have a model summarise each session into memories.
It produces better-written records than deterministic capture does.

It also means Forgelore spends the user's money on its own initiative, at a rate
they cannot predict, to do the thing it claims to save money on. It needs an API
key, which turns a local tool into a configured service. And its output varies
between runs, which makes both testing and the savings measurement (ADR-0012)
much harder to trust.

A long-running daemon has a smaller version of the same problem: state that
outlives the session, a process to supervise, and a new class of bugs that only
appear after days of uptime.

## Decision

Capture is deterministic: fingerprints, exit codes, command text, timing. No
model calls, no network access, no background service. Forgelore runs only when
invoked, does its work, and exits.

## Consequences

- Cost of running Forgelore is zero, and the savings figure is not offset by a
  hidden spend.
- Behaviour is reproducible, which is what makes contract tests and the A/B
  measurement meaningful.
- Records are rougher than a model would write them. Summarising is the agent's
  job, at the moment it proposes a record, using tokens the user was spending
  anyway.
- Anything requiring cross-session background work has to be redesigned as work
  that happens during an invocation.
