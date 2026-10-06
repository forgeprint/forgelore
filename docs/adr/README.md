# Architecture Decision Records

ADR-0001 to ADR-0015 record the locked decisions in docs/plan.md, section 2.
From ADR-0016 on they record decisions taken while implementing a phase; the
second column says which. Each is short on purpose: context, decision,
consequences, including the costs rather than only the benefits.

A locked decision is not re-argued in a pull request. If one turns out to be
wrong, the plan and the ADR change first and the code follows. A superseded ADR
keeps its number and its text; the ADR that replaces it says so.

| ADR | Source | Decision |
|---|---|---|
| [0001](0001-single-static-go-binary.md) | K1 | Go, compiled to a single static binary |
| [0002](0002-pure-go-sqlite.md) | K2 | modernc.org/sqlite for the derived index |
| [0003](0003-stdlib-only-dependencies.md) | K3 | The standard library, one dependency, all vendored |
| [0004](0004-own-minimal-mcp-server.md) | K4 | Our own minimal MCP server |
| [0005](0005-adapters-as-data.md) | K5 | Agent adapters are data, not code |
| [0006](0006-cli-as-universal-base.md) | K6 | The CLI is the universal base |
| [0007](0007-fail-open-hooks.md) | K7 | Hooks fail open, always |
| [0008](0008-no-background-model-calls.md) | K8 | No background model calls, no daemon (amended by 0027) |
| [0009](0009-markdown-as-source-of-truth.md) | K9 | Markdown files are the source of truth |
| [0010](0010-team-and-local-scopes.md) | K10 | Two scopes, team and local |
| [0011](0011-record-schema-versioning.md) | K11 | Versioned records, unknown fields preserved |
| [0012](0012-measurement-in-core.md) | K12 | Measurement in the core from day one |
| [0013](0013-untrusted-content-is-not-memory.md) | K13 | External content never becomes memory on its own |
| [0014](0014-apache-license-and-dco.md) | K14 | Apache-2.0 with DCO sign-off |
| [0015](0015-separate-repository.md) | K15 | A separate repository under forgeprint |
| [0016](0016-record-types-and-injection.md) | Phase 1 | Record types and what each one costs |
| [0017](0017-restricted-yaml.md) | Phase 1 | A restricted YAML dialect, written canonically |
| [0018](0018-configuration-and-precedence.md) | Phase 1 | Configuration files and their precedence |
| [0019](0019-cost-in-dollars-is-the-comparison.md) | Phase 3 | Cost in dollars is what the two arms are compared on |
| [0020](0020-canonical-events-and-json-mappings.md) | Phase 4 | Four canonical events, and mappings in JSON |
| [0021](0021-mcp-target-versions.md) | Phase 5 | Two MCP revisions, because one client needs both |
| [0022](0022-a-failure-test-that-is-not-data.md) | post-release | A failure test the mapping cannot answer alone |
| [0023](0023-build-provenance-instead-of-a-detached-signature.md) | post-release | Build provenance, not a detached signature |
| [0024](0024-npm-one-package-per-platform.md) | post-release | npm, with one package per platform |
| [0025](0025-the-extractor-learns-cli-errors.md) | post-release | The extractor learns what a CLI error looks like |
| [0026](0026-work-memory-is-captured-twice.md) | Phase 10 | Work memory, captured two ways |
| [0027](0027-distillation-runs-the-agents-own-cli.md) | K8 (amended) | Distillation runs the agent's own CLI, and only when asked |

## Writing a new one

Number sequentially, name the file after the decision, and keep the three
headings. State what was rejected and why, not only what was chosen.
