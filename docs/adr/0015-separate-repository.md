# ADR-0015: A separate repository under forgeprint

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K15

## Context

Forgelore is listed in the Forgeprint marketplace and shares its authors, which
makes a monorepo look convenient. The two projects have almost nothing else in
common: different languages, different release cadences, different audiences.

A user adopting Forgelore should not have to adopt Forgeprint, and should not
have to reason about whether the two are coupled.

## Decision

Its own repository in the forgeprint organisation, its own versioning, its own
releases, and no dependency on Forgeprint's code in either direction. The
marketplace listing is a link.

## Consequences

- Forgelore can be adopted on its own merits, by people who do not use
  Forgeprint.
- Release cycles are independent: a fix here does not wait for anything there.
- Shared code between the two would have to be duplicated or published
  separately. Given ADR-0003, duplication is the expected answer.
- Work that belongs to the marketplace listing is done in the Forgeprint
  repository, not here (plan, phase 8).
