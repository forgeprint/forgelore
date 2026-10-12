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

# The transcript entries and fields internal/worklog actually reads. Nothing
# else in that file is compared, and the first run of this check is why.
#
# A transcript holds the agent's own bookkeeping as well as the Messages API
# shaped entries, and that bookkeeping is keyed by things that change every
# run: `wireToolInputs` is a map whose keys are tool-use ids, and an
# attachment carries the JSON schema of whichever tools the machine had
# loaded. Comparing those reported drift on every capture, which is the same
# as reporting none — a check nobody can read is a check nobody runs.
#
# So the question this asks is the only one worth failing on: has the shape
# of what the parser reads changed?
TRANSCRIPT_TYPES = {"user", "assistant"}
TRANSCRIPT_FIELDS = ("message.", "type", "version", "timestamp", "sessionId",
                     "cwd", "gitBranch")


def transcript_shapes(directory):
    """Field paths per entry type in a captured transcript, parser's view only.

    No agent documents this format (ADR-0026), so the corpus is the only
    statement of what it looked like. Values differ on every run — uuids,
    times, the session's own content — so what is compared is the same thing
    the payloads are compared on: which fields exist.
    """
    path = os.path.join(directory, "transcript.jsonl")
    if not os.path.exists(path):
        return None
    out = {}
    with open(path) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                entry = json.loads(line)
            except ValueError:
                out.setdefault("<unreadable>", set())
                continue
            if not isinstance(entry, dict):
                continue
            kind = entry.get("type", "<none>")
            if kind not in TRANSCRIPT_TYPES:
                continue
            kept = {p for p in paths(entry) if p.startswith(TRANSCRIPT_FIELDS)}
            out.setdefault(kind, set()).update(kept)
    return out

old, new = shapes(sys.argv[1]), shapes(sys.argv[2])
drifted = False

old_t, new_t = transcript_shapes(sys.argv[1]), transcript_shapes(sys.argv[2])
if old_t is None and new_t is not None:
    print("+  transcript: captured now, nothing to compare against")
elif old_t is not None and new_t is None:
    print("!  transcript: the corpus has one and this capture produced none")
    drifted = True
elif old_t is not None:
    for kind in sorted(set(old_t) | set(new_t)):
        if kind not in new_t:
            # An entry type that stopped appearing is not necessarily a
            # break: a type the agent writes only sometimes is absent from a
            # capture that did not provoke it. It is still worth saying.
            print(f"?  transcript/{kind}: not in this capture")
            continue
        if kind not in old_t:
            print(f"+  transcript/{kind}: new entry type")
            continue
        gone, added = sorted(old_t[kind] - new_t[kind]), sorted(new_t[kind] - old_t[kind])
        if not gone and not added:
            print(f"ok transcript/{kind}: unchanged ({len(old_t[kind])} fields)")
            continue
        for field in gone:
            print(f"!  transcript/{kind}: {field} is gone")
            drifted = True
        for field in added:
            print(f"+  transcript/{kind}: {field} is new")
            drifted = True

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
    print("\nA field a mapping or the transcript parser reads may have moved. "
          "Check internal/agent/mappings and internal/worklog, then recapture "
          "the corpus.")
    sys.exit(1)
PYTHON
