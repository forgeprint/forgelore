# Security Policy

## Reporting a vulnerability

Report privately through GitHub's [private vulnerability reporting](https://github.com/forgeprint/forgelore/security/advisories/new)
on this repository. Please do not open a public issue for a security problem.

Include what you did, what happened, and the Forgelore version (`forgelore version`).
You will get an acknowledgement within a week.

## Threat model

Forgelore stores a project's accumulated debugging knowledge and feeds parts of
it back into a coding agent's context. That makes two things security-relevant:

**Secrets reaching disk or a git remote.** Records are plain files in the
repository. Redaction runs before a record is written, common credential
patterns are masked, and content inside a `<private>` tag is never stored. Team
records are additionally scanned before they can be committed. Local records
stay out of git entirely.

**Persistent prompt injection.** A record that an agent reads later is an
instruction channel. Candidate records derived from content an agent fetched
from the network are marked `tainted` and never become team memory without an
explicit human review step.

Forgelore makes no network calls of its own, sends no telemetry, and runs no
background service.

## Scope

In scope: anything that writes an unredacted secret to disk or to git, anything
that lets untrusted content become a trusted record without review, and any path
where a hook failure can block or alter an agent's work beyond injecting a hint.

Out of scope: vulnerabilities in the coding agents themselves, and in the
projects Forgelore is installed into.
