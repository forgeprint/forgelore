#!/usr/bin/env bash
# The single entry point CI runs. Keep the logic here, not in the workflow
# YAML, so the same checks run unchanged on a developer machine.
set -euo pipefail
cd "$(dirname "$0")/.."

./scripts/test.sh
./scripts/crosscheck.sh
./scripts/gitleaks.sh
