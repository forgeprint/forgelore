# ADR-0017: A restricted YAML dialect, written canonically

- Status: Accepted
- Date: 2026-09-25
- Phase: 1

## Context

Records need frontmatter, and frontmatter means YAML if it is to look familiar
next to AGENTS.md and SKILL.md. Full YAML is not an option: no parser for it
exists in the standard library, and ADR-0003 rules out adding one.

Full YAML is also more than we want. Anchors, aliases, nested maps, the many
string styles and the type coercion rules are a large surface for a file that
holds a dozen flat fields, and a large surface is where surprises live.

## Decision

One restricted dialect, used for both record frontmatter and config files.

Supported: flat key/value pairs, strings, integers, booleans, string lists in
both forms (`[a, b]` and `- a`), `#` comments, and ISO-8601 UTC timestamps as
strings.

Not supported: nested maps, anchors and aliases, multi-line string styles, and
tabs for indentation.

Reading is strict and fail-open: a file that uses something outside the dialect
is skipped rather than guessed at, the record simply does not exist for that
run, and `doctor` reports it with the file and the line.

Writing is canonical: fixed key order, fixed quoting rules, LF endings.
Free-text strings are always double-quoted; enum and identifier values, integers
and booleans never are. Unknown fields are preserved (ADR-0011) and written
after the known ones in the order they were read. Round-tripping a file must
produce a byte-identical result, and that is a required test.

## Consequences

- The parser is small enough to read in one sitting and to test exhaustively,
  which matters because everything else trusts it.
- Records stay diffable. A canonical writer means a one-field change is a
  one-line diff, which is what makes pull-request review of team memory
  workable (ADR-0009).
- Hand-written YAML that is valid elsewhere may be rejected here. The error
  message has to say which construct is unsupported, not just that parsing
  failed.
- Multi-line values have nowhere to go in frontmatter. They belong in the
  markdown body, which is the intended place for prose.
- If a genuine need for nesting appears, dotted keys (ADR-0018) are the answer
  rather than widening the dialect.
