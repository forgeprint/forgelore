---
schema: 1
id: 01M47M72D9WFARADYPSCSK4AB0
type: fix
scope: team
title: "Name the Windows npm packages windows, not win32; the os field stays win32"
created: 2026-10-06T03:30:46Z
source: user
tainted: false
fingerprint: 882cd007bd7a2c16
tags: [npm, publishing, windows]
---
npm refuses to create a package whose name contains win32, twice on two separate days, while every other platform publishes in the same burst. forgelore-windows-x64 published first try. Keep the os field as win32 inside the package: npm matches that against process.platform.
