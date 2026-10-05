# Captured MCP client traffic

The requests a real MCP client sent to a stdio server, one JSON-RPC message
per line, captured by a recording server and scrubbed of the account name.

```
testdata/mcp/<client>/<version>/<era>.jsonl
```

`internal/mcp` replays each file against the real server and checks the
replies, so a change that would stop a client connecting fails a test instead
of a connection.

## Why both eras

Claude Code 2.1.289 speaks either revision depending on which client runtime
a session uses, and the choice is not ours to make:

- `modern.jsonl` — the v2 runtime. It opens with `server/discover` carrying
  `_meta.io.modelcontextprotocol/protocolVersion: 2026-07-28` and never sends
  `initialize`.
- `legacy.jsonl` — the v1 runtime, forced with `MCP_SDK_GENERATION=v1`. It
  opens with `initialize` at `2025-11-25`, follows with
  `notifications/initialized`, and sends no `_meta` at all.

A server that implemented only the current specification would be invisible
to the second. That is why Forgelore is dual-era, and the two files are the
evidence rather than the assumption.

## What a replay cannot catch

The client validates the replies against its own schema, and a reply this
corpus accepts may still be rejected there. `tools/list` carrying
`cacheScope` without `ttlMs` was found that way: every hand-written test
passed, and the real client refused the tool list. Connecting a real client
stays part of the acceptance criterion for this reason.
