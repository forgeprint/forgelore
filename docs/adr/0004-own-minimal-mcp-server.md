# ADR-0004: Our own minimal MCP server

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K4

## Context

Model Context Protocol support lets any MCP client query Forgelore without an
agent-specific integration. The official SDKs are the normal way to implement
it, and they bring a dependency tree that ADR-0003 does not permit.

What Forgelore actually needs is narrow: stdio transport, initialize with
version negotiation, tools/list, tools/call, and correct error codes. Three to
five tools. No resources, no prompts, no sampling.

## Decision

Write the server ourselves: JSON-RPC over stdio using encoding/json, covering
only the methods above. The targeted specification version is recorded in its
own ADR when phase 5 starts, verified against the current specification rather
than from memory.

## Consequences

- No dependency, and the wire format is entirely under our control and fully
  tested.
- We own protocol compatibility. When MCP revises its specification we read the
  changelog and update; an SDK would have absorbed part of that work.
- The implementation has to stay minimal. Growing it into a general-purpose MCP
  library is the point at which adopting an SDK deserves reconsideration.
- Conformance is tested against the example messages in the specification, so
  drift surfaces as a failing test rather than as a client that silently
  refuses to connect.
