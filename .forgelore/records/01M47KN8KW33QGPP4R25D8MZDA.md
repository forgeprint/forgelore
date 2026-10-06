---
schema: 1
id: 01M47KN8KW33QGPP4R25D8MZDA
type: decision
scope: team
title: "An agent chains and pipes its commands, and both used to hide a failure"
created: 2026-10-06T03:21:03Z
source: user
tainted: false
tags: [fingerprint, agents]
---
A pipeline exits with the status of its last element, so a build piped through head looks successful; a chain was attributed to its first word, so cd or ls owned the compiler's diagnostic. Both are handled now. When a capture or a lookup comes back empty, read the command the model actually wrote before suspecting the mapping.
