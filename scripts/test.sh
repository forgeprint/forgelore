#!/usr/bin/env bash
# Correctness checks: formatting, vet, tests.
# CI calls this script; it must behave identically on a developer machine.
set -euo pipefail
cd "$(dirname "$0")/.."

go_files() {
	find . -name '*.go' -not -path './vendor/*' -not -path './.git/*'
}

echo "==> gofmt"
if [ -n "$(go_files)" ]; then
	unformatted="$(gofmt -l $(go_files))"
	if [ -n "$unformatted" ]; then
		echo "not gofmt'd:" >&2
		echo "$unformatted" >&2
		exit 1
	fi
fi

echo "==> go vet"
go vet ./...

echo "==> go test"
go test ./...
