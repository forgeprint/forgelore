# ADR-0011: Versioned records, unknown fields preserved

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K11

## Context

Records are committed to a repository and outlive the version of Forgelore that
wrote them. On one team, several developers will run different versions at the
same time, and one of them will open, rewrite and commit a record written by a
newer version.

The usual failure is quiet data loss: an older parser does not recognise a
field, drops it, and writes the file back without it.

## Decision

Every record carries a schema version in its frontmatter. A parser that meets an
unknown field reads it, keeps it, and writes it back unchanged. A record from a
newer schema version is usable rather than rejected, as far as the fields it
does understand allow.

## Consequences

- Rolling out a new version across a team does not have to be coordinated.
- A new field can be introduced without a migration, because older versions
  carry it through untouched.
- Round-tripping has to be tested explicitly: write, read with unknown fields
  present, write again, and the file must be unchanged.
- Preserving unknown fields means never silently normalising a record file.
  Rewrites stay minimal and targeted.
- It does not license unlimited schema growth. A field that stops being used is
  removed in a deliberate migration, documented in docs/versioning.md.
- The same rule covers unknown values, not only unknown fields. A record whose
  type this version does not recognise is preserved and treated as a note
  (ADR-0016): searchable, never injected. Degrading to the option that costs
  nothing is safer than guessing.
