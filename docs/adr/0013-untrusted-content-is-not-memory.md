# ADR-0013: External content never becomes memory on its own

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K13

## Context

Forgelore writes records that a future agent will read as trusted context. That
makes the record store an instruction channel, and automatic capture makes it a
writable one.

If a session fetched a web page, and a candidate record derived from that
session is stored and later injected, then whoever wrote that page has written
into every future session of everyone on the team. A prompt injection that
normally lasts one session becomes permanent.

The same mechanism leaks secrets: an error message with a token in it, captured
automatically, committed to a repository.

## Decision

Three rules, applied before anything is written:

1. Candidate records derived from a session that read external content are
   marked tainted. They stay local, and promoting one to team scope requires an
   explicit extra confirmation.
2. Known credential patterns are masked at write time, not at display time.
3. Content inside a private tag is never stored at all.

## Consequences

- The path from untrusted input to trusted memory always passes through a human.
- Redaction runs before the file is written, so a secret never reaches disk in
  cleartext and cannot be recovered from an earlier version of the file.
- Masking is pattern-based and will both miss things and occasionally mask
  something harmless. It is a layer, not a guarantee: team records are scanned
  again before they can be committed (phase 6).
- Some genuinely useful knowledge from documentation stays a suggestion rather
  than becoming memory automatically. That is the intended cost.
