---
schema: 1
id: 01M47M72FDAAXDEDYJNWKT7SNE
type: fix
scope: team
title: "Gemini CLI needs GEMINI_API_KEY and the matching selectedType"
created: 2026-10-06T03:30:46Z
source: user
tainted: false
fingerprint: e44d719003f9c5c1
tags: [agents, gemini-cli, capture]
---
Signing in with Google no longer reaches a model from this client. Setting the variable is not enough: the CLI reads security.auth.selectedType from merged settings, so a machine that once signed in with Google keeps doing so. scripts/capture-agent-events.sh writes gemini-api-key into the throwaway workspace for the capture only.
