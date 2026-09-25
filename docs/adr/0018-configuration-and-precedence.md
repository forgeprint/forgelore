# ADR-0018: Configuration files and their precedence

- Status: Accepted
- Date: 2026-09-25
- Phase: 1

## Context

Some settings belong to a team and are worth committing: the injection budget,
whether the session-start index is on. Others belong to one developer and must
not be: their A/B assignment, their personal overrides.

The two scopes already exist for records (ADR-0010), and configuration needs the
same split, plus the usual ability to override a value for one command.

## Decision

Configuration uses the restricted dialect of ADR-0017, with dotted keys for
grouping: `inject.budget_tokens`, `hook.timeout_ms`. No nested maps.

Two files: `.forgelore/config.yaml` for the team, committed; and
`.forgelore/local/config.yaml` for the individual, gitignored.

Precedence, highest first: command-line flag, environment variable, local
config, team config, built-in default.

An unknown key is a warning, never an error. It is reported by `doctor` and
ignored otherwise.

## Consequences

- A team can set a policy and an individual can override it locally, without
  either of them editing the other's file.
- Dotted keys give grouping without nesting, so the parser stays as simple as
  ADR-0017 requires, and every setting has exactly one spelling.
- Unknown keys being non-fatal means a config written for a newer Forgelore does
  not break an older one, matching how records behave (ADR-0011). The cost is
  that a typo silently does nothing until someone runs `doctor`.
- Five sources means the effective value of a setting is not obvious from any
  one file. `doctor` has to be able to show where each value came from.
