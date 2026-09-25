# ADR-0006: The CLI is the universal base

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K6

## Context

Coding agents differ in what they support. Some have hooks, some speak MCP, some
have neither. What every one of them can do is run a shell command.

Building on hooks first would mean the tool works for one agent and is useless
for the rest until each integration is written.

## Decision

Every feature exists as a CLI command first. Hooks and MCP are layers on top
that call the same code paths. Nothing is reachable through a hook or an MCP
tool that cannot be reached from a terminal.

Every command produces human-readable output by default and machine-readable
output with --json.

## Consequences

- The lowest support level is never zero. An agent with no integration at all
  still gets working memory through a one-line instruction in AGENTS.md telling
  it to run the command.
- The whole tool is testable and debuggable without running an agent, which
  makes the acceptance criteria in phase 2 something a human can verify by hand.
- Hook and MCP layers stay thin: translate input, call the core, translate
  output. Logic that drifts into them is a bug.
- Some agent-native conveniences are given up, because the command line is the
  narrowest common shape.
