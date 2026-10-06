---
schema: 1
id: 01M47KN8K96VXX3F9YKJC081DN
type: decision
scope: team
title: "Gemini CLI: signing in with Google no longer reaches a model"
created: 2026-10-06T03:21:03Z
source: user
tainted: false
tags: [agents, gemini-cli, capture]
---
It fails with IneligibleTierError. A capture needs GEMINI_API_KEY, and setting the variable is not enough because the CLI reads security.auth.selectedType from merged settings; capture-agent-events.sh writes gemini-api-key into the throwaway workspace. An untrusted folder also makes it ignore workspace settings entirely, so the profile sets GEMINI_CLI_TRUST_WORKSPACE.
