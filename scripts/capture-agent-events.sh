#!/usr/bin/env bash
# Captures a coding agent's real hook payloads into testdata/agents/.
#
# Rule 7 of this project says never to write about an agent's hook format from
# memory. Documentation is the next best thing and it is what the mapping was
# built from, but documentation describes an intent; these files are what the
# agent actually sent.
#
# The capture runs in a throwaway directory with its own settings, so nothing
# touches the user's own Claude Code configuration. Paths are scrubbed the
# same way testdata/errors is: the account name goes, the shape of the path
# stays, because the shape is what a mapping has to deal with.
#
# Requires the claude CLI, logged in:  claude  then  /login
set -euo pipefail
cd "$(dirname "$0")/.."
out_root="$PWD/testdata/agents/claude-code"

command -v claude >/dev/null || { echo "the claude CLI is not on PATH" >&2; exit 1; }
version="$(claude --version | awk '{print $1}')"
echo "==> claude $version"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/.claude" "$work/captured"

cat > "$work/capture.sh" <<'HOOK'
#!/usr/bin/env bash
payload=$(cat)
event=$(printf '%s' "$payload" | sed -n 's/.*"hook_event_name"[[:space:]]*:[[:space:]]*"\([A-Za-z]*\)".*/\1/p')
[ -z "$event" ] && event=unknown
n=1; while [ -e "captured/$event-$n.json" ]; do n=$((n+1)); done
printf '%s' "$payload" > "captured/$event-$n.json"
exit 0
HOOK
chmod +x "$work/capture.sh"

hook="{\"type\":\"command\",\"command\":\"$work/capture.sh\",\"timeout\":10}"
cat > "$work/.claude/settings.json" <<EOF
{
  "hooks": {
    "SessionStart": [{"hooks": [$hook]}],
    "PostToolUse": [{"matcher": "Bash", "hooks": [$hook]}],
    "PostToolUseFailure": [{"matcher": "Bash", "hooks": [$hook]}],
    "SessionEnd": [{"hooks": [$hook]}]
  }
}
EOF

# A Go package that does not compile, so the agent meets a real failure.
printf 'module example.com/broken\n\ngo 1.26\n' > "$work/go.mod"
mkdir -p "$work/cmd/app"
printf 'package main\n\nfunc main() {\n\tgreet("world")\n}\n' > "$work/cmd/app/main.go"

echo "==> running a session"
(
  cd "$work"
  claude -p 'Run exactly this one shell command and then stop, do not fix anything: go build ./...' \
    --allowedTools Bash --permission-mode acceptEdits >/dev/null 2>&1 || true
)

shopt -s nullglob
captured=("$work"/captured/*.json)
if [ ${#captured[@]} -eq 0 ]; then
  echo "nothing was captured; is the CLI logged in? run: claude  then  /login" >&2
  exit 1
fi

out="$out_root/$version"
mkdir -p "$out"
account="$(id -un)"
for f in "${captured[@]}"; do
  sed -e "s|/Users/$account|/Users/dev|g" \
      -e "s|/home/$account|/home/dev|g" \
      -e "s|\\\\$account|\\\\dev|g" \
      -e "s|$account|dev|g" \
      -e "s|$work|/work|g" \
      "$f" > "$out/$(basename "$f")"
done

echo "==> wrote ${#captured[@]} payload(s) to testdata/agents/claude-code/$version"
grep -l "$account" "$out"/*.json 2>/dev/null && { echo "the account name survived scrubbing" >&2; exit 1; }
echo "==> scrubbed"
