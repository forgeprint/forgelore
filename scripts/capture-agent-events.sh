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
# One exception, and it is read-only: an agent writes its session transcript
# under the user's own home, not under the throwaway directory, so keeping a
# transcript means reading from there. Nothing is written or deleted there,
# which leaves the capture's session file behind in the agent's own history.
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
	# Three things have to be forced, and only the first two were in the
	# documentation this profile was written from.
	#
	# Hooks are off by default; project configuration needs trust; and
	# `codex exec` sandboxes model-run commands, so the build the capture
	# depends on may never execute. The sandbox flag is reasoned from the
	# CLI's own --help and is **unverified** until a capture runs: if the
	# session produces no tool event, that is the first thing to look at.
	run_session() {
		"$bin" exec --dangerously-bypass-hook-trust \
			--dangerously-bypass-approvals-and-sandbox \
			-c features.hooks=true "$1"
	}
	;;
gemini-cli)
	bin=gemini
	version_cmd=(--version)
	config_path=".gemini/settings.json"
	# No PostToolUseFailure equivalent: AfterTool is the only tool event,
	# and whether a command failed has to be read out of its payload.
	events=(SessionStart AfterTool SessionEnd)
	write_config() {
		local h hooks=""
		for e in "${events[@]}"; do
			h="{\"type\":\"command\",\"command\":\"$1 $e\",\"timeout\":10000}"
			hooks="$hooks,\"$e\":[{\"hooks\":[$h]}]"
		done

		# The CLI reads the auth method from the *merged* settings, so a
		# workspace file can choose one for this capture alone. With
		# GEMINI_API_KEY set, say so here: otherwise a machine whose user
		# settings say oauth-personal ignores the key and the session dies
		# on IneligibleTierError before any tool runs — which is exactly
		# how the first two attempts failed.
		local auth=""
		if [ -n "${GEMINI_API_KEY:-}" ]; then
			auth=',"security":{"auth":{"selectedType":"gemini-api-key"}}'
		fi

		# timeout is milliseconds here, unlike Claude Code's seconds.
		printf '{"hooks":{%s}%s}' "${hooks:1}" "$auth"
	}
	# Two things a capture has to force.
	#
	# --yolo because run_shell_command "requires manual confirmation" and
	# there is nobody here to confirm it.
	#
	# GEMINI_CLI_TRUST_WORKSPACE because an untrusted folder makes the CLI
	# ignore the workspace .gemini/settings.json entirely — the hooks would
	# never load — and it downgrades --yolo back to prompting. A throwaway
	# directory is never trusted. The alternative was writing to the user's
	# own ~/.gemini/trustedFolders.json, which this script exists not to do.
	#
	# The documentation says folder trust is "disabled by default"; the code
	# reads `settings.security?.folderTrust?.enabled ?? true`, so it is on.
	# Checked against v0.62.0.
	run_session() { GEMINI_CLI_TRUST_WORKSPACE=true "$bin" -p "$1" --yolo; }
	;;
cursor)
	bin=cursor-agent
	version_cmd=(--version)
	config_path=".cursor/hooks.json"
	# afterShellExecution hands over the command and its output directly and
	# is one of the few hooks reported to fire in the CLI; postToolUseFailure
	# is the only one that states a failure outright. Both are mapped: if the
	# second never arrives, nothing is lost, and the capture says so by
	# simply not producing a file for it.
	events=(sessionStart postToolUse postToolUseFailure afterShellExecution sessionEnd)
	write_config() {
		local hooks=""
		for e in "${events[@]}"; do
			hooks="$hooks,\"$e\":[{\"type\":\"command\",\"command\":\"$1 $e\",\"timeout\":10}]"
		done
		printf '{"version":1,"hooks":{%s}}' "${hooks:1}"
	}
	# -p is mandatory: without it the CLI opens a TUI that hangs in a
	# subshell. --force lets it run the commands, --trust skips the
	# workspace prompt a throwaway directory would otherwise get.
	run_session() { "$bin" -p --force --trust "$1"; }
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
	echo "unknown agent $agent (claude-code, codex-cli, copilot-cli, cursor, gemini-cli)" >&2
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
# "exactly as written" is load-bearing. Left to itself a model writes
# `go build ./... 2>&1 | head -40`, and a pipeline exits with head's status,
# so the build failure never reaches the agent's failure event and the
# capture quietly comes back without one.
prompt='Run these two shell commands in order and then stop. Run each one exactly as written: no pipes, no redirection, no extra flags, no wrapping. Do not fix anything, do not edit any file. First: go build ./...   Second: go version'
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

# Emptied first. Payloads are numbered per event, so a capture whose model
# happened to call fewer tools than the last one leaves the extra files
# behind and the directory ends up holding two sessions at once — which
# reads as one and is not. Safe here because the run above already refused
# to continue if it captured nothing.
rm -rf "$out"
mkdir -p "$out"
account="$(id -un)"

# A function because the transcript below has to be scrubbed exactly the way
# a payload is. Cursor puts user_email in every payload, so the address goes
# by shape rather than by name: an agent that starts sending one tomorrow is
# covered without anybody noticing it needed to be.
scrub() {
	sed -e "s|/Users/$account|/Users/dev|g" \
		-e "s|/home/$account|/home/dev|g" \
		-e "s|\\\\$account|\\\\dev|g" \
		-e "s|$account|dev|g" \
		-e "s|$work|/work|g" \
		-e 's|[A-Za-z0-9._%+-]\{1,\}@[A-Za-z0-9.-]\{1,\}\.[A-Za-z]\{2,\}|dev@example.com|g' \
		"$1"
}

for f in "${captured[@]}"; do
	scrub "$f" > "$out/$(basename "$f")"
done
echo "==> wrote ${#captured[@]} payload(s) to $out"

# The session transcript, which is a second surface and a more fragile one.
# Phase 10's extractor reads it at the next session start (ADR-0026), and no
# agent documents its format, so the corpus is the only statement of what it
# actually looked like at a given version.
#
# The path comes out of the payloads rather than from a guess about where an
# agent keeps its sessions. An agent that sends no transcript_path — Copilot
# CLI sends none — simply leaves no file here, and that absence is itself
# true of the agent.
transcript="$(sed -n 's/.*"transcript_path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
	"${captured[@]}" | sed 's|\\\\|\\|g' | head -1)"
if [ -n "$transcript" ]; then
	# Documented as written asynchronously, so it can still be lagging the
	# SessionEnd hook that fired a moment ago. Waiting a few seconds is the
	# difference between a corpus with a transcript and one without.
	for _ in 1 2 3 4 5; do
		[ -s "$transcript" ] && break
		sleep 1
	done
fi
if [ -n "$transcript" ] && [ -s "$transcript" ]; then
	scrub "$transcript" > "$out/transcript.jsonl"
	echo "==> wrote the transcript ($(du -k "$out/transcript.jsonl" | cut -f1) KB)"
elif [ -n "$transcript" ]; then
	echo "==> no transcript at $transcript (it never appeared)"
else
	echo "==> $agent sent no transcript_path, so there is no transcript to keep"
fi

# Every file, not only the payloads: the transcript carries whole prompts and
# whole file contents, which is the likeliest place for a name to survive.
if grep -rl "$account" "$out" 2>/dev/null; then
	echo "the account name survived scrubbing" >&2
	exit 1
fi
echo "==> scrubbed"
