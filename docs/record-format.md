# Record format, schema version 1

Normative description of a Forgelore record. The parser and the canonical
writer are tested against this document; where code and document disagree, the
document is the bug report.

Background and rationale live in the ADRs: [0009](adr/0009-markdown-as-source-of-truth.md)
(files are the source of truth), [0011](adr/0011-record-schema-versioning.md)
(versioning and unknown fields), [0016](adr/0016-record-types-and-injection.md)
(types and injection), [0017](adr/0017-restricted-yaml.md) (the YAML dialect).

## File layout

One record per file, named after its identifier:

```
.forgelore/records/<ULID>.md     team scope, committed
.forgelore/local/<ULID>.md       local scope, gitignored
```

The directory determines the scope. The `scope` field is written for
readability and validated against the directory; if they disagree the directory
wins and `doctor` reports it. Promoting a record from local to team scope is a
file move, not an edit.

Identifiers are ULIDs in canonical form: 26 characters, Crockford base32,
uppercase. Lookups accept any case.

A file is frontmatter delimited by `---` on its own line, then a markdown body.

## Fields

Listed in canonical order. The writer emits them in exactly this order.

| Field | Type | Required | Notes |
|---|---|---|---|
| `schema` | integer | yes | `1` for this document |
| `id` | ULID | yes | Matches the file name |
| `type` | enum | yes | `fix`, `dead_end`, `decision`, `command`, `note` |
| `scope` | enum | yes | `team` or `local`; the directory is authoritative |
| `title` | string | yes | One line. This is the text that gets injected |
| `created` | timestamp | yes | ISO-8601 UTC, second precision, always `Z` |
| `source` | enum | yes | `user`, `hook`, `import` |
| `tainted` | boolean | yes | `true` if derived from external content (ADR-0013) |
| `fingerprint` | string | for `fix` and `dead_end` | Lowercase hex, one per record |
| `tags` | string list | no | Lowercase, kebab-case; omitted when empty |
| `related` | ULID list | no | One-way; the index treats links as undirected |
| `superseded_by` | ULID | no | A superseded record is never injected |

Unknown fields are preserved and written back after the known ones, in the
order they were read (ADR-0011). Optional fields are omitted entirely when
absent or empty; an empty list is never written as `[]`.

There is no `updated` field. Team records carry their history in git, and
adding a field that changes on every write would mean carving an exception out
of the round-trip test.

### title

The title is the payload. When a fingerprint matches, the title is what reaches
the model's context; the body is fetched only on request. So a title states the
answer rather than the question:

```
good: Set CGO_ENABLED=0 before cross-building; the arm64 link step fails otherwise
bad:  arm64 build error
```

It must be a single line. Redaction (ADR-0013) applies to the title and the
body before either is written.

### Behaviour by type

| Type | Session-start index | On a matching fingerprint | In search |
|---|---|---|---|
| `fix` | no | yes, full hint | yes |
| `dead_end` | no | yes, alongside the fix | yes |
| `command` | title only | no | yes |
| `decision` | title only | no | yes |
| `note` | no | no | yes |
| unrecognised | no | no | yes |

## Canonical writing

A record read and written again is byte-identical. The rules that make that
true:

- Fields in the order of the table above; unknown fields after them, in read
  order.
- `title` is always double-quoted, with `\` and `"` escaped.
- Values matching `^[A-Za-z0-9_.:+-]+$` are never quoted. That covers enums,
  ULIDs, fingerprints, timestamps, integers and booleans.
- Lists are written in flow form, `[a, b]`. The reader also accepts block form.
- Booleans are `true` and `false`.
- LF line endings, one trailing newline, no trailing spaces.

## Validation

Reading is strict and fail-open (ADR-0007, ADR-0017). A file that violates this
document is skipped: it does not exist for that run, nothing is injected from
it, and nothing is rewritten. `doctor` reports the file, the line and the rule
that was broken.

Skipped, not repaired: a malformed record may be malformed because someone is
mid-edit, and rewriting it would destroy their work.

## Example

```markdown
---
schema: 1
id: 01K68P7YQZ3M4N5R6S7T8V9W0X
type: fix
scope: team
title: "Set CGO_ENABLED=0 before cross-building; the arm64 link step fails otherwise"
created: 2026-09-25T17:42:03Z
source: hook
tainted: false
fingerprint: 9f2c4a1e7b3d0856
tags: [go, build, cross-compile]
related: [01K68P9AB2C3D4E5F6G7H8J9K0]
---

## Symptom

`GOARCH=arm64 go build ./...` fails in the link step.

## Fix

Set `CGO_ENABLED=0`. The default toolchain tries to link against the host's
C library, which is not available for the target.
```
