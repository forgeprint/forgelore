# ADR-0022: A failure test the mapping cannot answer alone

- Status: Accepted
- Date: 2026-10-05
- Phase: post-release

## Context

ADR-0020 made agent adapters data: a mapping names field paths and a failure
test, and supporting an agent should not need Go. Three failure tests
covered every agent met so far — a dedicated failure event, a field that
appears, a pattern in the output text.

Running the marketplace plugin in real Claude Code sessions found a case
none of them covers. A Bash command's exit status reaches Forgelore only
through which event fires: `PostToolUseFailure` for a command that failed,
`PostToolUse` for one that did not. Neither payload carries an exit code.
When the agent writes

```
go build ./... 2>&1 | head -40
```

the pipeline exits with `head`'s status, which is zero. Claude Code is right
to send `PostToolUse`, and the mapping was right to call it a success — but
the build failed, nothing was recorded, and nothing was recalled. It is
silent and indistinguishable from a session that met no errors.

Copilot CLI has the same shape and a way out: its result text ends
`<shellId: 0 completed with exit code 1>`, so `output_matches` reads the
exit code back. Claude Code's payload has no such marker. The only thing
left to ask is whether the output *looks* like a failure.

Two ways to ask it:

1. **A regex in the mapping.** No Go, no format change. But the pattern
   would be a second copy of what `internal/fingerprint` already knows, free
   to drift from it, and a file-and-line pattern reaches about half of the
   29 families in `testdata/errors` — panics, `ModuleNotFoundError`,
   `npm ERR!` and the rest would stay silent. A fix that looks complete and
   covers half is worse here than no fix, because nobody goes back to it.
2. **Ask the fingerprinter.** One copy of the knowledge, every family
   covered, and the answer improves whenever the corpus does.

## Decision

A fourth failure test, `output_has_diagnostic: true`, which is true when
`fingerprint.Scan` finds anything in the output. `claude-code.json` sets it
on `PostToolUse`. The mapping format moves to **2**.

The version moves although the field is an addition. The test is not whether
something was added but whether an older binary can read the file and still
be right: it does not know the field, drops it, and calls every piped
failure a success. Refusing the file out loud is the only honest behaviour,
and that is what a version number is for. A mapping that sets the field
while claiming version 1 is refused for the same reason.

## Consequences

**The format is no longer purely declarative.** One failure test is answered
by Go rather than read out of the payload, and ADR-0020's promise — that
adding an agent needs no Go — now has an exception: an agent needing a
*different* notion of "looks like an error" would need a change here. That
is the cost, and it buys the 29 families.

**The fingerprinter now decides what "failed" means** for every mapping that
sets the field. A change to its matchers changes which events are treated as
failures, not only how they are named. `internal/fingerprint` was already
the most carefully tested package in the project, which is the only reason
this is tolerable.

**A successful command that prints error-shaped text is read as a failure.**
`cat build.log`, `rg "undefined"`, an agent echoing an earlier error.

This was measured on 2026-10-06, by leaving Forgelore running in its own
repository for a session: **thirteen inferred failures, not one of them a
real one.** They were the output of scripts printing error examples while
the matchers were being worked on. No hint was injected, because nothing
matched, but the ledger filled with misses for commands that never failed
and the report's "nothing known" count stopped meaning anything.

So an inferred failure now **leaves no trace when it matches nothing**: no
ledger line, no session state. The lookup still happens, and a match is
still injected and still recorded, because then something really was known.
The residual risk is unchanged and real: a guess that does match injects a
line into a session where nothing broke.

Narrowing by command shape was tried first and discarded by the same
measurement — all thirteen came from commands containing a pipe or a chain,
so "only infer when the exit status could have been masked" would have
caught every one of them. There is no signal in the payload that separates
output which is an error from output which is about one.

**The gap closes only for agents that opt in.** Copilot CLI keeps
`output_matches`, because reading a reported exit code is narrower and
better than guessing from text. Codex CLI is unverified and unchanged.

## Verified

Claude Code 2.1.289, 2026-10-05, in live sessions with the plugin installed
from the marketplace: a piped failing build now injects the recorded fix,
and `go version` is still a success. The payload is in the corpus as
`testdata/agents/claude-code/2.1.289/PostToolUse-2.json`, beside the
succeeding one, and `TestAPipedBuildStillCountsAsAFailure` holds both.
