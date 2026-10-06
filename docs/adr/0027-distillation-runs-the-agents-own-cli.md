# ADR-0027: Distillation runs the agent's own CLI, and only when asked

- Status: Accepted
- Date: 2026-10-06
- Locked decision: K8 (amended)
- Supersedes in part: ADR-0008

## Context

ADR-0008 forbade model calls outright. Three of its four reasons were about an
API key: money spent on Forgelore's initiative, at an unpredictable rate, by a
tool whose whole claim is that it saves money, turning a local binary into a
configured service. The fourth was reproducibility.

Work memory (ADR-0026) changes the balance. Deterministic extraction can say
that a file was edited four times and a command started passing; it cannot say
what was wrong with the first three attempts. The agent's own closing message
covers part of the gap, and the rest is the kind of writing a model is good at
and a parser is not. A record that is actually worth injecting saves tokens on
every later session; the one-off cost of writing it well can be smaller than
what a rough record costs by being ignored.

So the question was never "model or no model". It was: whose key, whose money,
whose decision, and does the spend show up in the comparison.

## Decision

Distillation is allowed, as a third capture layer on top of A and B, under four
constraints.

**It runs a CLI the user has already installed.** `claude -p` and its
equivalents — the agent the user is already paying for, already signed in to.
Forgelore holds no API key, ships no HTTP client for a model API, and asks for
no credentials. Which binary, which flags, how the prompt goes in and how the
reply comes out is written in the mapping file, not in code (K5), so an agent
that changes its flags does not need a new binary.

**It is off by default.** Nobody's transcript reaches a vendor because they
installed Forgelore. Enabling it is a config change, and `doctor` reports
whether it is on and which binary it would run.

**It is triggered at session boundaries, and no hook ever waits for it.** The
hook spawns one short-lived child and exits inside its deadline (K7). The child
does its work and dies. This is not the daemon ADR-0008 ruled out: there is no
resident process, no supervision, no state that outlives the work. The result
is read at the next session start, which is where work memory is read anyway
(ADR-0026).

**At most one call per session, and the record says who wrote it.** A
distilled candidate carries `source: distill`, so a sentence a model wrote is
never indistinguishable from one a person wrote. The cap is one call per
session id, written by the child before it calls, so two hooks firing at once
cannot turn into two calls. An unbounded rate would make "the spend is
measured" a sentence with nothing behind it.

**It produces candidates, never records.** `review` still stands between a
model's sentence and the team's memory, exactly as it stands between a hook's
guess and the team's memory. What is sent is redacted first (ADR-0013), and
`<private>` content is never sent at all.

## Consequences

- The cost of running Forgelore is no longer zero when this is on, so it cannot
  be left out of the comparison. The spend goes on the cost side of the ledger
  and `report` nets it off. A release that cannot measure the spend has no
  business claiming the saving (ADR-0019).
- Reproducibility, which was ADR-0008's fourth reason, is genuinely weakened: a
  distilled candidate differs between runs. This is survivable only because the
  deterministic layer is the one under test. Contract tests and the A/B
  measurement run on A and B, which still behave identically every time.
- "No network calls" stops being true of the whole product and becomes true of
  the default. The README, the site and `llms.txt` currently promise it without
  qualification; they are accurate for the binary that is released today and
  have to be corrected in the same release that ships this, not before.
- `source` is a closed enum, validated strictly, so adding `distill` is a
  schema-visible change in a way that adding a field is not (ADR-0011 preserves
  unknown fields, not unknown enum values). An older binary reading a distilled
  record reports it as a problem and leaves it out — it degrades to not having
  the record, which is the safe direction, but it is not invisible.
- A second CLI surface to drift against: the agent's non-interactive flags and
  its output shape, on top of its hook payloads. It belongs in the mapping and
  in the drift detector (Phase 9).
- Sending a transcript to a model is the largest privacy step this project has
  taken. Redaction was written for build logs; here it guards the user's own
  words, and a miss is a leak rather than noise.
- A user with no agent CLI installed, or one that is not signed in, gets A and B
  and nothing less. Distillation degrades to silence, like every other optional
  path here.
