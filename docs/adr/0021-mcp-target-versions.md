# ADR-0021: Two MCP revisions, because one client needs both

- Status: Accepted
- Date: 2026-10-05
- Phase: 5

## Context

ADR-0004 committed to writing the MCP server rather than taking an SDK, and
deferred the target specification version to this phase, "verified against
the current specification rather than from memory".

Reading it changed the shape of the work. The current revision, `2026-07-28`,
**removed the initialize handshake**. The protocol is stateless: every
request carries its own protocol version and the client's capabilities in
`_meta.io.modelcontextprotocol/*`, a server must implement `server/discover`,
and a version it cannot speak is refused with `UnsupportedProtocolVersion`
(`-32022`) listing what it can. ADR-0004's scope — "initialize with version
negotiation" — describes a revision that is now the previous era.

The specification names both eras and says a server **MAY** implement both.
Whether Forgelore needs to was a question about one client, so it was
measured rather than reasoned about. Claude Code 2.1.289, against a recording
server:

- default runtime: `server/discover` with `_meta` at `2026-07-28`, and no
  `initialize` at any point;
- with `MCP_SDK_GENERATION=v1`: `initialize` at `2025-11-25`, then
  `notifications/initialized`, then `tools/list` carrying no `_meta` at all.

Which runtime a session uses depends on a feature flag, not on anything
Forgelore controls. The specification's compatibility matrix is explicit that
a legacy client meeting a modern-only server fails, with no fall-forward.

## Decision

Forgelore's MCP server is **dual-era**: `2026-07-28` and `2025-11-25`.

The era is chosen per request from how the client opens. An `initialize`
request selects legacy semantics for the life of the process. A request
carrying `_meta.io.modelcontextprotocol/protocolVersion` is served
statelessly. A request with neither, before any handshake, is malformed and
rejected with `-32602`.

Modern replies carry `resultType` and the server's identity in `_meta`.
Legacy replies carry neither, because a legacy client predates both.

## Consequences

- Half of Claude Code's sessions would not see Forgelore at all under a
  modern-only server. Supporting one revision would have been simpler and
  wrong.
- The server keeps exactly one piece of connection state, a flag saying this
  process was opened with `initialize`. The modern path never reads it, which
  is what keeps the statelessness requirement honest.
- Two eras mean two reply shapes and two sets of tests. The captured traffic
  in `testdata/mcp` is replayed against both, so a change that would break
  either fails a test.
- When `2025-11-25` stops appearing in the wild, the legacy half is a single
  branch to delete.
- Replaying captured requests is not the same as satisfying a client.
  `tools/list` carrying `cacheScope` with no `ttlMs` passed every test here
  and was refused by Claude Code's own schema validation. Connecting a real
  client is part of the acceptance criterion because of cases like that.
