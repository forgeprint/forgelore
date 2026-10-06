---
schema: 1
id: 01M47KN8JPFT0YPBP4W43WTVKY
type: decision
scope: team
title: "npm refuses a package named win32; the Windows packages are called windows"
created: 2026-10-06T03:21:03Z
source: user
tainted: false
tags: [npm, publishing, windows]
---
Publishing forgelore-win32-x64 returns 403 "Package name triggered spam detection", twice on two days, while every other platform publishes in the same burst. The os field inside the package still says win32, because npm matches that against process.platform. The name and the field disagree on purpose.
