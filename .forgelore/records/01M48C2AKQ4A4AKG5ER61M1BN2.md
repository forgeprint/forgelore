---
schema: 1
id: 01M48C2AKQ4A4AKG5ER61M1BN2
type: decision
scope: team
title: "A status line command must parse under /bin/sh, so no process substitution"
created: 2026-10-06T10:27:37Z
source: user
tainted: false
---
forgelore usage --help-wiring recommended tee >(forgelore usage >/dev/null) until 2026-10-06. Process substitution is a bashism: /bin/sh answers with "syntax error near unexpected token `('" and the status line dies before forgelore runs, so nothing is measured and nothing says why. Read stdin into a variable instead: in=$(cat); printf '%s' "$in" | forgelore usage. The shell error itself is not fingerprinted, so this will not come back as a hint.
