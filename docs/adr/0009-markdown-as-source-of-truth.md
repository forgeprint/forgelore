# ADR-0009: Markdown files are the source of truth

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K9

## Context

Team memory has to be shared, reviewed and trusted. The natural way to share
anything in a software project is the repository it already has.

A database as the source of truth fails every part of that: it is a binary blob
in git, it produces merge conflicts nobody can resolve, it cannot be reviewed in
a pull request, and it is unreadable without the tool that wrote it.

## Decision

One markdown file per record, with frontmatter for metadata and a markdown body.
These files are authoritative. The SQLite index (ADR-0002) is derived: it is
gitignored, it can be deleted at any time, and it is rebuilt from the files.

## Consequences

- Records are reviewed like code. A bad record is caught in a pull request.
- Concurrent additions by different developers do not conflict, because each
  record is its own file with a unique identifier.
- Memory is readable with an editor, and survives Forgelore being uninstalled.
- Reading every file is slower than querying a database, which is why the index
  exists. Keeping the two consistent is our problem: the index tracks file
  modification time and a content hash, and a mismatch triggers a rebuild rather
  than a wrong answer.
- Record files must stay small and human-sized. Anything that only a machine can
  read belongs in the index, not in a record.
