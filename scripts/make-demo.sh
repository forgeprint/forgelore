#!/usr/bin/env bash
# Builds site/demo.svg from site/demo.txt.
#
# A GIF was considered for the README on 2026-10-05 and refused, because
# producing one needed tooling this machine does not have and the result
# would have been the one artefact in the repository nobody could
# regenerate. This keeps the moving picture and loses neither property:
# the source is text, the output is text, and both are diffable.
#
# The SVG uses CSS keyframes, which every browser animates and which no
# build step is needed to produce. A reader with animation turned off, or
# a screen reader, gets the same transcript as plain text beside it —
# site/demo.txt is readable on its own.
set -euo pipefail
cd "$(dirname "$0")/.."

src=site/demo.txt
out=site/demo.svg
[ -f "$src" ] || { echo "$src is missing" >&2; exit 1; }

python3 - "$src" "$out" <<'PYTHON'
import html
import sys

src, out = sys.argv[1], sys.argv[2]

# Lines, as (kind, text). "cmd" is typed and carries a prompt, "cont" is
# the rest of the same command and carries none, "out" is printed. A
# continuation with a prompt in front of it would read as a second command,
# which is the one thing a terminal picture must not get wrong.
lines = []
delays = []          # seconds to wait *after* the line at the same index
for raw in open(src, encoding="utf-8"):
    raw = raw.rstrip("\n")
    if raw.startswith("#") or (not raw.strip() and not raw.startswith("=")):
        continue
    if raw == "=":
        lines.append(("out", ""))
        delays.append(0.0)
    elif raw.startswith("~ "):
        if delays:
            delays[-1] += float(raw[2:])
    elif raw.startswith("$ "):
        lines.append(("cmd", raw[2:]))
        delays.append(0.0)
    elif raw.startswith("| "):
        lines.append(("cont", raw[2:]))
        delays.append(0.0)
    elif raw.startswith("> "):
        lines.append(("out", raw[2:]))
        delays.append(0.0)
    elif raw == ">":
        lines.append(("out", ""))
        delays.append(0.0)

if not lines:
    sys.exit("demo.txt holds no lines")

# A printed line lands quickly; a typed one takes a moment, as if somebody
# were typing it. Neither is measured from anything — this is a reading
# pace, not a recording.
BASE = {"cmd": 0.55, "cont": 0.45, "out": 0.22}

start = []
t = 0.0
for i, (kind, text) in enumerate(lines):
    start.append(t)
    t += BASE[kind] if text else 0.12
    t += delays[i]
total = t

PAD_X, PAD_Y, LINE_H, FONT = 22, 44, 21, 14
width = 860
height = PAD_Y + LINE_H * len(lines) + 26

def css_time(v):
    # Keyframe percentages, so the whole thing is one animation that loops
    # with the rest. Clamped: a line must never start before the frame.
    return max(0.0, min(100.0, v / total * 100.0))

rules = []
for i, s in enumerate(start):
    at = css_time(s)
    nxt = min(100.0, at + 0.01)
    rules.append(
        f"#demo-term .l{i}{{animation:r{i} {total:.2f}s linear infinite}}"
        f"@keyframes r{i}{{0%,{at:.3f}%{{opacity:0}}{nxt:.3f}%,100%{{opacity:1}}}}"
    )

body = []
for i, (kind, text) in enumerate(lines):
    y = PAD_Y + LINE_H * (i + 1)
    cls = "out" if kind == "out" else "cmd"
    prompt = '<tspan class="sign">$</tspan> ' if kind == "cmd" else ""
    body.append(
        f'<text class="l{i} {cls}" x="{PAD_X}" y="{y}">{prompt}'
        f"{html.escape(text)}</text>"
    )

svg = f"""<svg xmlns="http://www.w3.org/2000/svg" id="demo-term" viewBox="0 0 {width} {height}"
     width="{width}" height="{height}" role="img"
     aria-label="A terminal: a build fails, the fix is recorded, and a different command finds it again.">
  <title>Forgelore: one error, recorded once, found again</title>
  <desc>The same transcript is in site/demo.txt as plain text.</desc>
  <style>
    #demo-term .bg{{fill:#13151a}}
    #demo-term .bar{{fill:#1d2026}}
    #demo-term text{{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;
          font-size:{FONT}px;white-space:pre}}
    #demo-term .cmd{{fill:#e8eaed}} #demo-term .out{{fill:#9aa2ad}}
    #demo-term .sign{{fill:#5ab77f}} #demo-term .dot{{opacity:.6}}
    @media (prefers-reduced-motion:reduce){{
      #demo-term text{{animation:none!important;opacity:1!important}}
    }}
    {"".join(rules)}
  </style>
  <rect class="bg" width="{width}" height="{height}" rx="10"/>
  <rect class="bar" width="{width}" height="30" rx="10"/>
  <rect class="bar" y="20" width="{width}" height="10"/>
  <circle class="dot" cx="20" cy="15" r="5" fill="#e06c75"/>
  <circle class="dot" cx="38" cy="15" r="5" fill="#e5c07b"/>
  <circle class="dot" cx="56" cy="15" r="5" fill="#98c379"/>
  {chr(10).join("  " + b for b in body)}
</svg>
"""

with open(out, "w", encoding="utf-8") as f:
    f.write(svg)

# And inlined into each page, between markers.
#
# Referencing the file from <img> froze it at frame zero, and from <object>
# it froze on GitHub Pages while playing from a local server — measured both
# ways on 2026-10-06. An inline SVG is part of the document and has no
# embedding mode to get wrong, so the uncertainty is removed rather than
# worked around. Every selector is scoped to #demo-term, because an inline
# <style> in HTML applies to the whole page.
START, END = "<!-- demo:start -->", "<!-- demo:end -->"
indented = "\n".join("  " + ln if ln.strip() else ln for ln in svg.strip().split("\n"))
for page in ("site/index.html", "site/tr/index.html"):
    text = open(page, encoding="utf-8").read()
    a, b = text.index(START), text.index(END)
    text = text[: a + len(START)] + "\n" + indented + "\n  " + text[b:]
    open(page, "w", encoding="utf-8").write(text)

print(f"{out}: {len(lines)} lines, {total:.1f}s loop; inlined into 2 pages")
PYTHON
