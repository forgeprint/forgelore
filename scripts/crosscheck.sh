#!/usr/bin/env bash
# Cross-compiles every release target with CGO_ENABLED=0.
# This is the guard that catches a dependency or an import pulling cgo back in.
set -euo pipefail
cd "$(dirname "$0")/.."

targets="
linux/amd64
linux/arm64
darwin/amd64
darwin/arm64
windows/amd64
windows/arm64
"

version="${FORGELORE_VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
mkdir -p dist

for target in $targets; do
	goos="${target%/*}"
	goarch="${target#*/}"
	ext=""
	if [ "$goos" = "windows" ]; then ext=".exe"; fi

	echo "==> ${goos}/${goarch}"
	CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build \
		-trimpath \
		-ldflags "-s -w -X main.version=${version}" \
		-o "dist/forgelore_${goos}_${goarch}${ext}" \
		./cmd/forgelore
done

echo "all targets built with CGO_ENABLED=0"
