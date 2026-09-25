# ADR-0014: Apache-2.0 with DCO sign-off

- Status: Accepted
- Date: 2026-09-25
- Locked decision: K14

## Context

Forgelore is meant to be installed inside other people's repositories, including
commercial ones. The licence has to be one a company's legal review passes
without a conversation.

It also needs a defensible provenance story for contributions, without a
contributor licence agreement that requires a signature before a first pull
request and discourages drive-by fixes.

## Decision

Apache License 2.0, and the Developer Certificate of Origin, version 1.1.
Contributors sign off their commits with git commit -s. There is no CLA.

## Consequences

- Apache-2.0 includes an explicit patent grant, which MIT does not. For a tool
  that will be adopted inside companies, that is the difference between an easy
  approval and a long one.
- Copyright stays with each contributor. The project cannot relicense on its own
  later, which is a constraint accepted deliberately.
- Sign-off is enforced by the GitHub DCO application rather than by CI, keeping
  a legal check out of the build pipeline.
- LICENSE and DCO are verbatim copies of their canonical sources. The appendix
  in the Apache text keeps its placeholder form, as the licence intends.
