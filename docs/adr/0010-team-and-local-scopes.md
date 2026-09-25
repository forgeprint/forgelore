# ADR-0010: Two scopes, team and local

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K10

## Context

Not everything worth remembering is worth sharing. A fix that took an afternoon
to find belongs to the team. A note about one developer's machine, a half-formed
guess, or a candidate record captured automatically and not yet reviewed, does
not.

With a single shared scope, one person's wrong record becomes everyone's wrong
context, and the cost of that is paid in tokens by every future session.

## Decision

Two scopes. Team records live in .forgelore/records/ and are committed. Local
records live in .forgelore/local/ and are gitignored. Automatically captured
candidates always start local; moving one to team scope is an explicit promote
step with a human in the loop.

## Consequences

- The blast radius of a bad record is one person until someone reviews it.
- Team memory carries a review history, because it went through a pull request.
- Two scopes means every read path merges both, and every write path picks one.
  Precedence rules have to be defined and tested rather than assumed.
- Local records are lost if the working copy is. That is the intended trade:
  they are unreviewed by definition.
- Usefulness statistics stay local (phase 6). Team records must not become a
  record of who struggled with what.
