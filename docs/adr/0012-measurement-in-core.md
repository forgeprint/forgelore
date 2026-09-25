# ADR-0012: Measurement in the core from day one

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K12

## Context

Forgelore's claim is that it reduces token usage. Every tool in this category
makes that claim, and almost none of them can show it. Injecting context to save
context is not self-evidently a win: a hint that does not help is pure cost.

Measurement added later is measurement that never happens, because by then the
architecture has no place to put it and the team has no baseline to compare
against.

## Decision

The injection ledger is part of the core, built in phase 3, before the first
agent integration. Every injection records its time, session, record, byte count
and estimated token count. An optional A/B mode assigns a share of sessions to a
control group that gets no injection but is still logged.

The ledger is local and never leaves the machine.

## Consequences

- The savings figure is an observation with a method behind it, not a marketing
  number. Differences that are not statistically meaningful are reported as
  such.
- We will sometimes learn that a feature costs more than it saves. That is the
  point of building the measurement first.
- Estimated tokens are estimates. The report says so, and where an agent exposes
  real usage data we read it instead (phase 3).
- For agents whose usage data we cannot read, the report shows only what was
  spent and claims no saving. An honest half-answer beats a confident whole one.
