# ADR-0005: Agent adapters are data, not code

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K5

## Context

Forgelore aims to support many coding agents. Each has its own hook event names,
its own JSON shape for hook input and output, and its own configuration file
location. These change between releases, and they change without asking us.

If support for an agent is Go code, then every renamed event field is a code
change, a release, and an upgrade for every user, for something that is really
one string in one place.

## Decision

Support for an agent is a mapping file: event name translation, field paths,
response format, timeout. Mapping files live in adapters/, are embedded into the
binary, and carry their own format version. A file in the user's project under
.forgelore/adapters/ overrides the embedded one.

## Consequences

- Fixing a broken agent integration means editing one data file. A user can do
  it locally, before a release exists, without a Go toolchain.
- Contributors can add agent support without writing Go, which widens who can
  usefully contribute (see CONTRIBUTING.md).
- The mapping format has to be expressive enough for real differences between
  agents, and every bit of expressiveness is a small interpreter we maintain.
  Where an agent needs behaviour the format cannot express, that is a signal to
  extend the format deliberately, not to add a special case in code.
- Every mapping needs real captured event samples and a contract test, so that
  a drift detector (phase 9) can catch an agent update before users do.
