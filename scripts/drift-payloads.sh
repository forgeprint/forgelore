#!/usr/bin/env bash
# Captures an agent's hook payloads again and reports how their shape has
# changed since the corpus was recorded.
#
# This is the half of drift detection that can actually tell whether an
# agent has broken a mapping, and it needs that agent installed and signed
# in, so it runs on a person's machine and not in CI.
#
# It compares field paths, not values. A session id and a timestamp differ on
# every run and mean nothing; a field that appeared or disappeared is the
# whole story, because a field path is exactly what a mapping depends on.
set -euo pipefail
cd "$(dirname "$0")/.."

agent="${1:-}"
[ -n "$agent" ] || { echo "usage: $0 <agent>" >&2; exit 1; }

mapping="internal/agent/mappings/$agent.json"
[ -f "$mapping" ] || { echo "no mapping for $agent" >&2; exit 1; }

old_version="$(sed -n 's/.*"verified_against"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$mapping")"
if [ -z "$old_version" ]; then
	echo "$agent has never been verified, so there is nothing to compare against." >&2
	echo "Capture a first corpus with: ./scripts/capture-agent-events.sh $agent" >&2
	exit 1
fi
old_dir="testdata/agents/$agent/$old_version"
[ -d "$old_dir" ] || { echo "$old_dir is missing" >&2; exit 1; }

fresh="$(mktemp -d)"
trap 'rm -rf "$fresh"' EXIT

echo "==> capturing $agent again"
FORGELORE_CAPTURE_OUT="$fresh" ./scripts/capture-agent-events.sh "$agent" >/dev/null

echo "==> comparing field shapes against $old_dir"
python3 - "$old_dir" "$fresh" <<'PYTHON'
import json, os, sys

def paths(value, prefix=""):
    """Every dotted path in a payload, which is what a mapping reads."""
    if isinstance(value, dict):
        out = set()
        for key, child in value.items():
            out |= paths(child, f"{prefix}.{key}" if prefix else key)
        return out
    if isinstance(value, list):
        # One entry stands for the list: a mapping addresses the field, not
        # the index.
        return paths(value[0], prefix + "[]") if value else {prefix + "[]"}
    return {prefix}

def shapes(directory):
    out = {}
    for name in sorted(os.listdir(directory)):
        if not name.endswith(".json"):
            continue
        event = name.split("-", 1)[0]
        with open(os.path.join(directory, name)) as f:
            out.setdefault(event, set()).update(paths(json.load(f)))
    return out

old, new = shapes(sys.argv[1]), shapes(sys.argv[2])
drifted = False

for event in sorted(set(old) | set(new)):
    if event not in new:
        print(f"!  {event}: no longer captured")
        drifted = True
        continue
    if event not in old:
        print(f"+  {event}: new event, not in the corpus")
        drifted = True
        continue
    gone, added = sorted(old[event] - new[event]), sorted(new[event] - old[event])
    if not gone and not added:
        print(f"ok {event}: unchanged ({len(old[event])} fields)")
        continue
    drifted = True
    for field in gone:
        print(f"!  {event}: {field} is gone")
    for field in added:
        print(f"+  {event}: {field} is new")

if drifted:
    print("\nA field a mapping reads may have moved. Check "
          "internal/agent/mappings, then recapture the corpus.")
    sys.exit(1)
PYTHON
