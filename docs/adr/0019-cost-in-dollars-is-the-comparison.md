# ADR-0019: Cost in dollars is what the two arms are compared on

- Status: Accepted
- Date: 2026-10-05
- Phase: 3

## Context

ADR-0012 commits to showing whether Forgelore saves anything, by comparing a
treatment arm against a control arm. That needs a per-session figure for what
the session consumed, read from the agent rather than guessed.

Checking what Claude Code actually exposes, against the official reference on
2026-10-05, turned up less than the plan assumed:

- **Hooks receive no usage data at all.** The common input fields are
  `session_id`, `prompt_id`, `transcript_path`, `cwd`, `permission_mode`,
  `effort` and `hook_event_name`. So the code path that does the injecting
  cannot see what the session spent.
- **The status line receives a cumulative cost**, `cost.total_cost_usd`,
  documented as "the estimated cost of all API calls in the current session".
- **The status line's token counts are not cumulative.**
  `context_window.total_input_tokens` is documented as "tokens currently in
  the context window, from the most recent API response". Summing those
  readings, or taking the last one, answers a different question than "what
  did this session cost".
- **OpenTelemetry does export cumulative tokens** as `claude_code.token.usage`,
  but there is no file exporter. Collecting it means running a receiver, and a
  long-running process is what K8 rules out.

## Decision

The arms are compared on **cost in dollars per session**, taken as the highest
reading of `cost.total_cost_usd` seen for that session. Tokens are reported
only as what Forgelore itself injected, labelled as an estimate of four bytes
per token.

Usage reaches Forgelore through `forgelore usage`, which reads the status line
payload on stdin. The user wires it into their status line; nothing is
collected behind their back and no process stays running.

A second comparison, **repeated errors per session**, is computed from
Forgelore's own ledger. It needs nothing from the agent.

## Consequences

- The headline number is in dollars, not tokens, which is further from the
  claim the project makes and closer to what the user actually pays. It is the
  agent's client-side estimate at list price, so the report says it is an
  estimate of a bill and not a bill.
- Cost is only available where the user has wired the status line up. Where
  they have not, the report shows what was spent and claims no saving, which
  ADR-0012 already requires.
- Repeated errors per session works everywhere, needs no wiring, and measures
  the mechanism directly rather than its monetary shadow. It may turn out to
  be the more useful of the two.
- Reading cost from a status line means the figure arrives on the status
  line's schedule, not ours. Since the counter is cumulative this costs
  nothing but a little duplication in the ledger: the report takes the maximum
  per session, never the sum.
- If an agent later exposes cumulative tokens somewhere a one-shot process can
  read them, this decision should be revisited. Tokens are the thing the claim
  is about.
