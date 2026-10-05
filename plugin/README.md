# Forgelore as a Claude Code plugin

This directory is the plugin. It carries no binary: a plugin with a `bin/`
directory is not installable from claude.ai, so `forgelore` has to be on your
`PATH` already. Build it with `./scripts/build.sh` and put `dist/forgelore`
somewhere on your path.

The hooks it registers, and why each one:

| Event | What Forgelore does | Timeout |
|---|---|---|
| `SessionStart` | Hands over a budgeted list of command and decision titles | 5 s |
| `PostToolUseFailure` (Bash) | Fingerprints the error and injects a hint if one is known | 5 s |
| `PostToolUse` (Bash) | Notices a command that failed earlier and now works, and proposes a fix | 5 s |
| `SessionEnd` | Says how many proposals are waiting | 1 s |

The timeouts are the outer guard. Forgelore holds itself to `hook.deadline_ms`
(500 ms by default) and gives up quietly rather than making anybody wait.
`SessionEnd` gets 1 second because Claude Code gives every `SessionEnd` hook
together a 1.5 second budget.

Every hook exits successfully, whatever happened (K7). If Forgelore is broken,
missing, or looking at a damaged store, the session carries on without it.
`forgelore doctor` is where the damage shows.
