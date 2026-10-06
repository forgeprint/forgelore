---
schema: 1
id: 01M47KN8HDA7SZFZ3V1NT4VHA8
type: decision
scope: team
title: "The claude CLI signs in separately from the desktop app; run claude then /login"
created: 2026-10-06T03:21:03Z
source: user
tainted: false
tags: [agents, claude-code, capture]
---
Its OAuth session also expires on its own schedule and headless `claude -p` then fails with "OAuth session expired and could not be refreshed" while an interactive session may still work. Capturing agent payloads needs a live session, so check this first when a capture comes back empty.
