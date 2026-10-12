# ADR-0028: A fifth canonical event, for the moment somebody asks

- Status: Accepted
- Date: 2026-10-12
- Phase: 10
- Amends: ADR-0020

## Context

ADR-0020 settled on four canonical events and gave the reason: Claude Code
alone fires more than thirty, and mirroring an agent's lifecycle would mean
every new agent adding vocabulary to a model that only needed four questions
answered — has a session begun, did a command fail, did one succeed, has the
session ended.

Work memory asks a fifth question. A session-start index is budgeted and
general; it cannot hold what the project knows about the file somebody is
about to ask about. The moment that knows which part of the codebase is in
play is the moment a prompt is submitted, and none of the four events is that
moment.

## Decision

A fifth canonical event, `prompt_submitted`, carrying the submitted text.

The test ADR-0020 set — a question Forgelore answers, not a lifecycle an agent
has — is met: this is "has somebody named something", and an agent that sends
no such event maps nothing and loses only this one feature.

Acting on it is off by default (`inject.on_prompt`). It is the only path in
Forgelore that can spend on every turn rather than on an event, and in Claude
Code it fires more often than a person types: a scheduled task, a background
subagent reporting back, and a message from another session all arrive through
it.

What is matched is narrow, and in a package rather than in the hook: paths and
identifiers — snake_case, kebab-case, camel case either way up — and never the
prose around them. A search index asked about a whole prompt returns something
for almost anything, which is the same as returning nothing useful at a cost
per turn. Pasted content, which the documentation says arrives expanded between
`<pasted_content>` markers, is cut out before matching: it is not what the
person said.

What is offered is titles of `work` and `decision` records, under a budget of
their own (`inject.prompt_budget_tokens`, 150), each at most once per session.
Fixes stay out: they answer an error that has not happened, and the error path
already carries them.

The ledger gets its own event, `prompt`, rather than counting these as
injections. They are the one spend that is not answering a failure, and a
report that mixed the two could not say which half paid for itself.

## Consequences

- Five canonical events instead of four, and the argument for keeping the set
  small is weaker for having been broken once. The next addition should have to
  answer why it is not a mapping detail.
- The prompt reaches Forgelore. It is read for what it names and stored
  nowhere — not in the ledger, which keeps ids and byte counts, and not in the
  session state, which keeps record ids. A future feature that wanted to keep
  prompt text would be a new decision, not an extension of this one (ADR-0013).
- `prompt` was verified in the documentation and not in a captured payload: the
  capture script runs four events and this is not one of them. Until a capture
  includes it, the field name is documented and unverified, which is what the
  mapping's `verified_against` cannot express on its own.
- Per-session deduplication lives in the session state file, so it is lost when
  the scratchpad is cleared. The cost of losing it is one repeated title.
- One index query per term, up to eight. The index joins a query's words with
  AND, and a prompt naming two files is not asking for a record that mentions
  both.
