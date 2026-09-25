#!/usr/bin/env bash
# Builds the host binary into dist/.
# CGO_ENABLED=0 is not optional: decision K1 requires a static binary that
# needs no runtime on the user's machine.
set -euo pipefail
cd "$(dirname "$0")/.."

version="${FORGELORE_VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
ext="$(go env GOEXE)"

mkdir -p dist
CGO_ENABLED=0 go build \
	-trimpath \
	-ldflags "-s -w -X main.version=${version}" \
	-o "dist/forgelore${ext}" \
	./cmd/forgelore

echo "built dist/forgelore${ext} (${version})"
"./dist/forgelore${ext}" version
