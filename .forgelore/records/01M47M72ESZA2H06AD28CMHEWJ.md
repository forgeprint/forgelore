---
schema: 1
id: 01M47M72ESZA2H06AD28CMHEWJ
type: fix
scope: team
title: "The claude CLI signs in separately: run claude, then /login"
created: 2026-10-06T03:30:46Z
source: user
tainted: false
fingerprint: 557cf6d000768eea
tags: [agents, claude-code, capture]
---
Its session expires on its own schedule and is not the desktop app's. An interactive session may still work while headless claude -p fails, so check both. Agent payload capture needs a live session; this is the first thing to check when a capture comes back empty.
