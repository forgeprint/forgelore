#!/usr/bin/env bash
# Captures a coding agent's real hook payloads into testdata/agents/.
#
# Rule 7 of this project says never to write about an agent's hook format
# from memory. Documentation is the next best thing and it is what every
# mapping starts as, but documentation describes an intent; these files are
# what the agent actually sent. The Claude Code mapping was wrong in the one
# place that mattered until a captured payload said so.
#
# Each capture runs in a throwaway directory with its own settings, so
# nothing touches the user's own configuration. Paths are scrubbed the way
# testdata/errors is: the account name goes, the shape of the path stays.
#
# Usage: capture-agent-events.sh [agent]   (default: claude-code)
set -euo pipefail
cd "$(dirname "$0")/.."
out_root="$PWD/testdata/agents"

agent="${1:-claude-code}"

# Each profile answers four questions: what the binary is called, how to ask
# it its version, where its hook configuration goes, and how to write that
# configuration. Adding an agent is adding a profile.
case "$agent" in
claude-code)
	bin=claude
	version_cmd=(--version)
	config_path=".claude/settings.json"
	events=(SessionStart PostToolUse PostToolUseFailure SessionEnd)
	write_config() { # $1 = hook script
		local h hooks=""
		for e in "${events[@]}"; do
			h="{\"type\":\"command\",\"command\":\"$1 $e\",\"timeout\":10}"
			hooks="$hooks,\"$e\":[{\"hooks\":[$h]}]"
		done
		printf '{"hooks":{%s}}' "${hooks:1}"
	}
	run_session() { "$bin" -p "$1" --allowedTools Bash --permission-mode acceptEdits; }
	;;
codex-cli)
	bin=codex
	version_cmd=(--version)
	config_path=".codex/hooks.json"
	events=(SessionStart PostToolUse SessionEnd)
	write_config() {
		local h hooks=""
		for e in "${events[@]}"; do
			h="{\"type\":\"command\",\"command\":\"$1 $e\"}"
			hooks="$hooks,\"$e\":[$h]"
		done
		printf '{"hooks":{%s}}' "${hooks:1}"
	}
	# Hooks are off by default and project config needs trust, so both are
	# forced here rather than left to whatever the machine is set to.
	run_session() {
		"$bin" exec --dangerously-bypass-hook-trust \
			-c features.hooks=true "$1"
	}
	;;
copilot-cli)
	bin=copilot
	version_cmd=(--version)
	# Repo-level .github/hooks was not read in a throwaway directory, and
	# writing to the real ~/.copilot would touch the user's own setup.
	# COPILOT_HOME points the agent at a home of its own.
	config_path="copilot-home/hooks/forgelore.json"
	events=(sessionStart postToolUse postToolUseFailure sessionEnd)
	write_config() {
		local hooks=""
		for e in "${events[@]}"; do
			hooks="$hooks,\"$e\":[{\"type\":\"command\",\"bash\":\"$1 $e\",\"timeoutSec\":10}]"
		done
		printf '{"version":1,"hooks":{%s}}' "${hooks:1}"
	}
	run_session() { COPILOT_HOME="$work/copilot-home" "$bin" -p "$1" --allow-all-tools; }
	;;
*)
	echo "unknown agent $agent (claude-code, codex-cli, copilot-cli)" >&2
	exit 1
	;;
esac

command -v "$bin" >/dev/null || { echo "$bin is not on PATH" >&2; exit 1; }
version="$("$bin" "${version_cmd[@]}" | tr -d '\r' | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)"
[ -n "$version" ] || { echo "could not read $bin's version" >&2; exit 1; }
echo "==> $agent $version"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/$(dirname "$config_path")" "$work/captured"

cat > "$work/capture.sh" <<'HOOK'
#!/usr/bin/env bash
# $1 is the event, passed by the hook entry that fired. It is an argument
# rather than something parsed out of the payload because Copilot CLI's
# payloads do not name their event at all.
event="${1:-unknown}"
n=1; while [ -e "$CAPTURE_DIR/$event-$n.json" ]; do n=$((n+1)); done
cat > "$CAPTURE_DIR/$event-$n.json"
exit 0
HOOK
chmod +x "$work/capture.sh"
export CAPTURE_DIR="$work/captured"
write_config "$work/capture.sh" > "$work/$config_path"

# A Go package that does not compile, so the agent meets a real failure, and
# a command that works, so both tool events fire.
printf 'module example.com/broken\n\ngo 1.26\n' > "$work/go.mod"
mkdir -p "$work/cmd/app"
printf 'package main\n\nfunc main() {\n\tgreet("world")\n}\n' > "$work/cmd/app/main.go"

echo "==> running a session"
prompt='Run these two shell commands in order and then stop. Do not fix anything, do not edit any file. First: go build ./...   Second: go version'
( cd "$work" && run_session "$prompt" 2>&1 | tail -3 ) || true

shopt -s nullglob
captured=("$work"/captured/*.json)
if [ ${#captured[@]} -eq 0 ]; then
	echo "nothing was captured; is $bin signed in, and are its hooks enabled?" >&2
	exit 1
fi

# FORGELORE_CAPTURE_OUT redirects the capture somewhere other than the
# corpus, which is how drift-payloads.sh compares a fresh capture against the
# committed one without overwriting it first.
out="${FORGELORE_CAPTURE_OUT:-$out_root/$agent/$version}"
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

echo "==> wrote ${#captured[@]} payload(s) to $out"
if grep -l "$account" "$out"/*.json 2>/dev/null; then
	echo "the account name survived scrubbing" >&2
	exit 1
fi
echo "==> scrubbed"
