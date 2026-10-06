# ADR-0025: The extractor learns what a CLI error looks like

- Status: Accepted
- Date: 2026-10-06
- Phase: post-release

## Context

Forgelore was installed in its own repository, and the first thing that
happened was that it could not record the day's work. Five real failures,
all met while building and releasing this project:

```
Error authenticating: IneligibleTierError: This client is no longer supported…
npm error code E403 … Package name triggered spam detection
npm error code EOTP
Failed to authenticate: OAuth session expired and could not be refreshed
the working tree is not clean; commit or stash first
```

`fingerprint.Scan` found nothing in any of them. They could not be `fix`
records, because a fix is keyed to a fingerprint and there was none.

The reason is the corpus. All twenty-nine families are compiler or runtime
diagnostics — `go`, `dotnet`, `node`, `python`, `typescript` — and the
matchers were written from them: a file with a line and a column, a panic,
an exception header, a test failure. Every one of them is anchored at the
start of a line, which is what keeps stack frames and progress output out.

Nothing was broken. The scope was simply narrower than the product's own
claim: "when a shell command fails, remember the fix". A shell command
fails for reasons that never reach a compiler, and the injection path was
blind to all of them.

## Decision

Four changes, each grounded in output captured from a real tool rather than
remembered.

**`toolErrorCode`** — `npm error code E403`, `npm ERR! code E404`. Only the
code line is taken. npm prefixes every line of a failure with its own name,
and the prose lines carry a log path with a timestamp in it, which would
hash differently on every run.

**`cliErrorLine`** — a line beginning `error:` or `fatal:`, with rustc's
optional bracketed code. This is how git, cargo, rustc and clang report.

**`failedToLine`** — `Failed to …`, `Unable to …`, `Cannot to …`. The
loosest matcher here, and the only one that recognises a tool which gives
nothing but a sentence.

**A label in front of an exception is dropped.** `exceptionLine` now
accepts one `Something: ` prefix and keeps only the text from the
identifier, so Gemini's `Error authenticating: IneligibleTierError: …`
hashes the same as the bare `IneligibleTierError: …`. The label belongs to
whoever reported the error, not to the error.

Two families join the corpus, captured rather than written:
`npm/registry-404` and `git/not-a-repository`.

## Consequences

**A line that begins `error:` is now a failure, and that changes an agent
mapping.** `claude-code`'s `PostToolUse` decides failure with
`output_has_diagnostic` (ADR-0022), so a successful command whose output
starts a line with `error:` is now treated as a failed one. The test that
held the old behaviour, `TestSuccessIsNotMistakenForFailure`, was rewritten
to state the narrower guarantee it still buys: the word *in the middle of a
sentence* proves nothing. That is what line anchoring is for, and it is the
whole defence against false positives here.

**`failedToLine` will match prose that is not an error.** A tool that
prints `Unable to reach the cache, continuing` during a successful run now
produces a diagnostic. The cost is bounded — a lookup that misses, and a
candidate proposal that a human rejects — and it was accepted because the
alternative is not seeing an entire class of failure.

**Some failures still have no shape, and that is correct.** `the working
tree is not clean` is a sentence with nothing diagnostic about it. Matching
it would mean matching any sentence.

**The corpus now needs the network and npm.** `capture-errors.sh cli` does
a registry lookup. It is not in `ci.sh`, like the rest of the capture
scripts.

**git/not-a-repository has identical variants.** git says the same sentence
wherever it is not a repository, so the two captures are byte for byte the
same and the family is listed in `variantsIdentical` rather than letting a
test claim a comparison it did not make.

## Addendum, 2026-10-06: the shell that will not parse the command

The same gap, found the same way, later the same day. Wiring a status line
for the trial week turned up that this project's own
`forgelore usage --help-wiring` printed a command `/bin/sh` cannot parse —
and that the error it printed was invisible to the extractor:

```
/bin/sh: -c: line 0: syntax error near unexpected token `('
```

This is the failure an agent causes most easily, because the agent writes
the command. Nothing above recognised it: there is no file with a line and
a column, no `error:` prefix, no exception, and `syntax error near …` is
not a sentence beginning `Failed to`.

**`shellSyntaxError`** takes the text from `syntax error` or `parse error`
onwards, and only when the phrase is continued by ` near `, `: ` or `:`.
What sits in front of it — the shell's own name, the script, a line
number — is position and is dropped, exactly as a file and a line are
dropped everywhere else. The token the shell names is kept, because it is
the only thing that separates one of these from another.

The phrase must be preceded by the start of the line or by a colon and a
space. That is the whole defence against prose: *that was a syntax error
near the top of the file* has a space in front of it, and is held out by a
case in `TestNoDiagnostic`.

Two families join the corpus, captured: `shell/unexpected-token` and
`shell/unexpected-eof`.

### Consequences

**Which shell `/bin/sh` is decides the wording, so the fingerprint is
per-shell.** bash writes `syntax error near unexpected token`, zsh writes
`parse error near`, and a shell that capitalises `Syntax error:` puts the
token at the end. The same mistake under two shells is two fingerprints.
Normalising them into one would mean deciding that these sentences mean the
same thing, which is a claim about shells and not about text; two memories
is the honest outcome, and each is still found by the shell that produces
it.

**It feeds `output_has_diagnostic` too.** A command that succeeds while
quoting a shell syntax error in its output — a log being echoed, a test
printing one — is now inferred to have failed under `claude-code`'s
`PostToolUse` (ADR-0022). The cost is the same bounded one as before: a
lookup that misses, which since 2026-10-06 leaves no trace in the ledger
when the failure was inferred rather than stated.

**zsh never reaches this matcher.** It writes `probe.sh:3: parse error
near …` with no space before the line number, so `fileLineDiag` matches
first and lands on the same message. Both roads are covered by a test,
because either one changing would be a change.

**The capitalised form is accepted but was not captured.** `/bin/sh` on the
machine that made the corpus is bash. The matcher takes the shape and a
test holds it to that; neither claims which shell writes it. If that corpus
is ever captured on a machine whose `/bin/sh` is something else, the two
families will come back with different text — which is the capture doing
its job, not a regression.
