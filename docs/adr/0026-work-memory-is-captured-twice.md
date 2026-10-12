# ADR-0026: Work memory, captured two ways

- Status: Accepted, amended by ADR-0027
- Date: 2026-10-06
- Phase: 10

## Context

Forgelore was built to spend fewer tokens than it saves, and every mechanism it
has is triggered by a failing command. That covers the loop a repeated error
costs, and misses the larger one: a new session rediscovers the codebase. Which
file does what, where a piece of work stopped, why a path was taken — none of it
is derivable from a fingerprint, and all of it is paid for again every session.

`claude-mem` solves this by recording tool-use observations and calling a model
to compress them into summaries, which it injects at session start. The storage
and retrieval halves of that design are what Forgelore already has: records,
SQLite with FTS5, a budgeted session index, progressive disclosure through
`search` and `show`. The capture half is what Forgelore lacks, and the way
`claude-mem` fills it is closed to us: K8 forbids a background model call.

Three ways to capture it were considered.

**A, deterministic extraction.** The agent writes a session transcript and
tells the hook where it is. Facts can be read out of it with no model at all:
which files were edited, which commands passed, which records were opened. Free
and exact, but it cannot produce a reason. A transcript shows that a file was
rewritten four times; it does not say what was wrong with the first three.

**B, the agent writes it.** At the end of a session the hook spends one line
asking for a record. The context is already in the agent, so no second model
call happens — the cost is output tokens in a session that is ending anyway,
and the risk is that the agent ignores the line.

**C, a background model call.** What `claude-mem` does. It is the only one of
the three that produces a reason without the agent's cooperation, and it costs
money per session, needs a key or a signed-in CLI, and breaks K8.

## Decision

**A and B together.** Facts come from the transcript, judgement comes from the
agent.

C was rejected here, and ADR-0027 reinstates it the same day as an optional
third layer: off by default, run through a CLI the user already has, producing
candidates only. A and B remain the floor — they cost nothing and they work
with no configuration, which is what makes C optional rather than necessary.

Extraction happens at **session start, from the previous session's
transcript** — not at the end of the session that produced it. The hooks
reference states the transcript is written asynchronously and may not include
the current turn when a hook fires, so reading it at `SessionEnd` is reading a
file that is structurally incomplete. By the next session it is finished, there
is no deadline pressure from a session that is trying to exit, and the reading
happens at the moment the result is needed.

B is the agent's own closing account of the session, taken from the last text
block of the last assistant entry in the transcript. Two earlier readings of
this were wrong, and checking is what corrected both:

- The nudge — a line at the end of a session asking the agent for a record —
  cannot work. The hooks reference says `SessionEnd` has "No decision control"
  and cannot add context; its only channel is `systemMessage`, which reaches
  the person. The one event that could carry a nudge to the model is a session
  *start*, and asking there for a summary of a session that has not happened
  yet is not a nudge.
- `last_assistant_message`, which the same reference documents on `SessionEnd`
  and `Stop`, is absent from the captured payloads. It was missing at 2.1.291
  and missing again at 2.1.296 in a session that did reach a model and did run
  tools — those payloads carry `reason` where the documentation says
  `exit_reason`, and nothing resembling the message. Whether an interactive
  session sends it is unknown and cannot be captured by a script that drives
  the CLI non-interactively. The transcript carries the same thing and was
  verified against a real file, so that is where it is read from, and the
  question is closed rather than left open: what can be observed does not
  carry the field.

New record type `work`, amending ADR-0016. It appears in the session-start
index by title only, under a budget of its own, newest first.

Retrieval on prompt submission — matching the prompt against the index and
offering titles — ships **off by default**. It is the one path that spends
tokens on every prompt rather than on an event.

A candidate derived from a transcript is **not** marked `tainted`. ADR-0013
exists for content that came from outside the project, and a transcript is a
local file the agent wrote about this project. Marking every candidate tainted
would make the mark meaningless and the promotion flow unusable. What does the
work instead is scope: anything extracted is born local and reaches the team
only through `review`.

The compaction event is a separate phase (10.1). Catching what is about to be
discarded is the most valuable moment to write, and it is a second dependency
on an agent's internal format. Phase 10 has to be complete without it.

## Consequences

- Reading at session start means choosing a file, and on a machine running
  several sessions at once "the previous session" is not well defined. A
  watermark of transcripts already extracted lives under `cache/`, which is
  derived and deletable, and a transcript still being appended to has to be
  left alone. Getting this wrong means either a session's work is never
  captured or another session's work is attributed to this project twice.
- The extractor reads only the subset of the transcript that is the Messages
  API wire format: `user`/`assistant` entries, the `text`, `thinking`,
  `tool_use` and `tool_result` blocks inside `message.content`, and `is_error`
  on a result. Everything else in the file is Claude Code's own bookkeeping —
  undocumented, unversioned, and in one measured session a quarter of the lines.
  Which files were edited and which commands passed are both available without
  it, so nothing is gained by reading further in.
- The transcript format is not part of any agent's documented interface. This
  adds a surface to the drift detector (Phase 9) that can break on a patch
  release, and the extractor has to treat a transcript it does not understand
  the way it treats a missing one: silence, inside the deadline (K7).
- Reading the closing message from the transcript means it arrives at the same
  moment as the facts, out of the same file, through one code path. B stops
  being a second channel and becomes a second kind of fact: what the agent did,
  and what the agent said about it.
- A work candidate is keyed by its session id, because it has no fingerprint to
  be keyed by. The id comes from the agent and is readable, which a second
  identifier would not be. The cost is that everything from one session folds
  into one entry, so a session is one decision in the review queue and not
  several.
- Until the `work` type exists, an accepted work candidate is recorded as a
  `note`: searchable, never injected. The alternative was a record whose type
  the binary that wrote it would refuse to validate.
- The strongest claim Forgelore can make about work memory is weaker than the
  one it makes about errors. There is no fingerprint to attribute a saving to,
  so the measurement is a count of discovery tool calls per session across the
  two arms — a proxy, and reported as one (ADR-0019).
- A transcript contains every prompt in the session, which makes this the
  hardest test redaction has faced. Redaction runs before anything is written,
  and a failure here leaks the user's own words rather than a build log.
- Overlap with the agent's own memory is now real and admitted in `plan.md`.
  What is left as the difference is that these records are local files, shared
  through git, measured, and tied to no agent.
- `work` is a type, so an older Forgelore reading one degrades the safe way:
  searchable, never injected (ADR-0011).
