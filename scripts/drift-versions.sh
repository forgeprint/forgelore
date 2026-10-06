#!/usr/bin/env bash
# Reports agents that have released a version nobody has checked Forgelore
# against.
#
# This is the half of drift detection that needs no credentials, so it can
# run unattended. It does not look at a single payload: it compares each
# mapping's verified_against with what the agent has published. A difference
# does not mean anything is broken, only that nobody knows yet.
#
# The other half, scripts/drift-payloads.sh, is the one that can actually
# tell. It needs the agents installed and signed in, so it runs on a person's
# machine.
#
# Exits 1 when something has drifted, so CI can act on it.
set -euo pipefail
cd "$(dirname "$0")/.."

# One source per agent, each a published endpoint that needs no token.
latest_version() {
	case "$1" in
	claude-code)
		curl -fsSL https://downloads.claude.ai/claude-code-releases/latest
		;;
	copilot-cli)
		curl -fsSL https://registry.npmjs.org/@github/copilot/latest |
			sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
		;;
	codex-cli)
		curl -fsSL https://registry.npmjs.org/@openai/codex/latest |
			sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
		;;
	gemini-cli)
		curl -fsSL https://registry.npmjs.org/@google/gemini-cli/latest |
			sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
		;;
	cursor)
		# Not on a registry: the install script pins the version it fetches.
		curl -fsSL https://cursor.com/install |
			sed -n 's|.*downloads\.cursor\.com/lab/\([0-9.]*\)-.*|\1|p' | head -1
		;;
	*)
		return 1
		;;
	esac
}

verified_against() {
	sed -n 's/.*"verified_against"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
		"internal/agent/mappings/$1.json"
}

drifted=0
for mapping in internal/agent/mappings/*.json; do
	agent="$(basename "$mapping" .json)"

	latest="$(latest_version "$agent" || true)"
	if [ -z "$latest" ]; then
		echo "?  $agent: could not read its published version"
		drifted=1
		continue
	fi

	verified="$(verified_against "$agent")"
	case "$verified" in
	"")
		# Never verified is its own state. It must not be reported as
		# up to date just because no version has moved since.
		echo "!  $agent: never verified against a captured payload (latest is $latest)"
		drifted=1
		;;
	"$latest")
		echo "ok $agent: verified against $verified, which is the latest"
		;;
	*)
		echo "!  $agent: verified against $verified, but $latest has been released"
		drifted=1
		;;
	esac
done

if [ "$drifted" -ne 0 ]; then
	echo
	echo "Recapture with: ./scripts/capture-agent-events.sh <agent>"
	echo "Then compare:   ./scripts/drift-payloads.sh <agent>"
	exit 1
fi
