# ADR-0016: Record types and what each one costs

- Status: Accepted
- Date: 2026-09-25
- Phase: 1

## Context

Forgelore's whole argument is that it spends fewer tokens than it saves. That
argument is decided by one question per record type: when does this type reach
a model's context?

"All memory, all the time" is the design that fails. A session-start dump of
everything is a fixed tax on every session, paid whether or not anything in it
is relevant.

## Decision

Five types, each with a fixed injection behaviour:

| Type | Session-start index | On a matching error | In search |
|---|---|---|---|
| `fix` | no | yes, full hint | yes |
| `dead_end` | no | yes, alongside the fix | yes |
| `command` | title only | no | yes |
| `decision` | title only | no | yes |
| `note` | no | no | yes |

`fix` and `dead_end` are event-triggered and cost nothing until a fingerprint
matches. `command` and `decision` are stable project knowledge, worth a title in
the session-start index and nothing more until asked for. `note` is never
injected automatically; it exists to be searched.

Two link fields:

- `related: [id, ...]` — stored one way, treated as undirected by the index.
- `superseded_by: id` — a superseded record is never injected, by any path. It
  stays searchable, marked as superseded, because why something was replaced is
  often the useful part.

An unrecognised type is preserved verbatim (ADR-0011) and behaves like `note`:
searchable, never injected.

## Consequences

- The session-start budget is spent on titles of stable knowledge, not on error
  history, which keeps the fixed cost per session small and predictable.
- Anything injected at error time is charged to a fingerprint match, so the
  ledger (ADR-0012) can attribute every injected byte to a trigger.
- Writing a good `title` matters more than writing a good body: the title is
  what gets injected, the body is what gets fetched on request.
- A record cannot opt out of its type's behaviour. Wanting different behaviour
  means a different type, and a new type is an ADR.
- An older Forgelore reading a newer record with an unknown type degrades to
  the safe option: it stays out of context rather than guessing.
