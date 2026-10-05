#!/usr/bin/env bash
# Captures real error output into testdata/errors/.
#
# Every sample here was produced by running the command it names. That matters:
# fingerprinting is calibrated against these files, and a sample copied from a
# web page is an assertion about what a tool prints rather than an observation
# of it.
#
# Each family is captured twice, in workspaces with different names and with
# the failing line in a different place. The two variants are the same error and
# must produce the same fingerprint; different families must not.
#
# Usage: scripts/capture-errors.sh [go|python|ts|dotnet|all]
set -uo pipefail
cd "$(dirname "$0")/.."

REPO="$PWD"
OUT="$REPO/testdata/errors"
WS="${TMPDIR:-/tmp}/forgelore-errcap"
TODAY="$(date -u +%Y-%m-%d)"

rm -rf "$WS"
mkdir -p "$WS"

# scrub removes anything identifying while keeping the shape of the text: a
# Windows user directory stays a Windows user directory, so the normalisation
# rules still have the same job to do.
#
# Node prints paths inside JavaScript string literals, where every separator is
# doubled: C:\\Users\\name\\app.js. A rule written for single separators walks
# straight past it, so the doubled form is handled first and the account name
# is substituted by itself at the end, whatever shape it arrived in.
scrub() {
	local me esc
	me="${USERNAME:-${USER:-}}"
	if [ -n "$me" ]; then
		esc="$(printf '%s' "$me" | sed -e 's|[][\.*^$/&|]|\\&|g')"
	else
		esc=''
	fi

	sed -E \
		-e 's|\\\\Users\\\\[^\\]+\\\\|\\\\Users\\\\dev\\\\|g' \
		-e 's|\\Users\\[^\\]+\\|\\Users\\dev\\|g' \
		-e 's|:/Users/[^/]+/|:/Users/dev/|g' \
		-e 's|/c/Users/[^/]+/|/c/Users/dev/|g' \
		-e 's|/home/[^/]+/|/home/dev/|g' \
		-e 's|/Users/[^/]+/|/Users/dev/|g' |
		if [ -n "$esc" ]; then sed -e "s|$esc|dev|g"; else cat; fi
}

# capture <family> <variant> <workdir> <tool> <command...>
capture() {
	local family="$1" variant="$2" dir="$3" tool="$4"
	shift 4
	local cmd="$*"

	local out code
	out="$(cd "$dir" && eval "$cmd" 2>&1)"
	code=$?

	if [ "$code" -eq 0 ]; then
		echo "  !! ${family}/${variant}: command succeeded; this should have failed" >&2
		return 1
	fi

	local target="$OUT/$family/$variant.txt"
	mkdir -p "$(dirname "$target")"
	{
		echo "---"
		echo "command: \"$cmd\""
		echo "exit: $code"
		echo "family: $family"
		echo "tool: \"$tool\""
		# The platform of the machine that captured it. Hard-coding this was
		# wrong: the Go corpus was captured on Windows and regenerating it on
		# a Mac left every file claiming Windows while holding Unix paths.
		echo "platform: $(go env GOOS 2>/dev/null || uname -s | tr 'A-Z' 'a-z')/$(go env GOARCH 2>/dev/null || uname -m)"
		echo "captured: $TODAY"
		echo "---"
		printf '%s\n' "$out" | scrub
	} >"$target"
	echo "  ok ${family}/${variant} (exit $code, $(wc -l <"$target") lines)"
}

# pad prints n blank lines, so the same error lands on a different line number
# in the second variant.
pad() {
	local n="$1"
	local i
	for ((i = 0; i < n; i++)); do echo; done
}

# ---------------------------------------------------------------- Go

capture_go() {
	echo "== go =="
	local tool
	tool="$(go version)"

	local w
	for w in alpha beta; do
		local pad_lines=0 dir="$WS/go-$w"
		[ "$w" = beta ] && pad_lines=7
		mkdir -p "$dir"
		printf 'module %s\n\ngo 1.26\n' "$w" >"$dir/go.mod"
	done

	local variant dir pad_lines
	for variant in a b; do
		if [ "$variant" = a ]; then
			dir="$WS/go-alpha"
			pad_lines=0
		else
			dir="$WS/go-beta"
			pad_lines=7
		fi

		{
			echo "package main"
			echo
			echo 'import "fmt"'
			pad "$pad_lines"
			echo "func main() {"
			echo '	fmt.Println(greet("world"))'
			echo "}"
		} >"$dir/main.go"
		capture go/undefined-identifier "$variant" "$dir" "$tool" "go build ./..."

		{
			echo "package main"
			echo
			echo 'import "fmt"'
			pad "$pad_lines"
			echo "func main() {"
			echo '	var count int = "many"'
			echo "	fmt.Println(count)"
			echo "}"
		} >"$dir/main.go"
		capture go/type-mismatch "$variant" "$dir" "$tool" "go build ./..."

		{
			echo "package main"
			echo
			echo 'import "fmt"'
			pad "$pad_lines"
			echo "func main() {"
			echo "	unused := 42"
			echo '	fmt.Println("hello")'
			echo "}"
		} >"$dir/main.go"
		capture go/unused-variable "$variant" "$dir" "$tool" "go build ./..."

		# go vet reports the same compiler error through a different shape,
		# with a "vet: " prefix in front of the position. Dropping the verb
		# from a fingerprint was meant to make one error one memory across
		# build, test, run and vet; this is the family that proves it.
		{
			echo "package main"
			pad "$pad_lines"
			echo "func main() {"
			echo '	greet("world")'
			echo "}"
		} >"$dir/main.go"
		capture go/vet-undefined "$variant" "$dir" "$tool" "go vet ./..."

		{
			echo "package main"
			echo
			echo 'import "fmt"'
			pad "$pad_lines"
			echo "func main() {"
			echo '	fmt.Println("hello")'
		} >"$dir/main.go"
		capture go/syntax-error "$variant" "$dir" "$tool" "go build ./..."

		{
			echo "package main"
			echo
			echo 'import "github.com/forgeprint/does-not-exist/nope"'
			pad "$pad_lines"
			echo "func main() {"
			echo "	nope.Run()"
			echo "}"
		} >"$dir/main.go"
		capture go/unknown-import "$variant" "$dir" "$tool" "GOFLAGS=-mod=mod GOPROXY=off go build ./..."

		{
			echo "package main"
			echo
			echo 'import "fmt"'
			pad "$pad_lines"
			echo "func main() {"
			echo "	var counts map[string]int"
			echo '	counts["a"] = 1'
			echo "	fmt.Println(counts)"
			echo "}"
		} >"$dir/main.go"
		capture go/panic-nil-map "$variant" "$dir" "$tool" "go run ."

		{
			echo "package main"
			echo
			echo 'import "fmt"'
			pad "$pad_lines"
			echo "func main() {"
			echo "	items := []string{\"a\", \"b\"}"
			echo "	fmt.Println(items[5])"
			echo "}"
		} >"$dir/main.go"
		capture go/panic-index-range "$variant" "$dir" "$tool" "go run ."

		rm -f "$dir/main.go"
		{
			echo "package demo"
			echo
			echo 'import "testing"'
			pad "$pad_lines"
			echo "func TestAnswer(t *testing.T) {"
			echo "	got := 41"
			echo "	if got != 42 {"
			echo '		t.Errorf("answer = %d, want 42", got)'
			echo "	}"
			echo "}"
		} >"$dir/demo_test.go"
		capture go/test-failure "$variant" "$dir" "$tool" "go test ./..."
		rm -f "$dir/demo_test.go"
	done
}

# ---------------------------------------------------------------- Python

capture_python() {
	echo "== python =="
	local tool
	tool="$(python --version 2>&1)"

	local variant dir pad_lines
	for variant in a b; do
		if [ "$variant" = a ]; then
			dir="$WS/py-alpha"
			pad_lines=0
		else
			dir="$WS/py-beta"
			pad_lines=6
		fi
		mkdir -p "$dir"

		{ pad "$pad_lines"; echo "import nosuchpackage"; echo "print(nosuchpackage)"; } >"$dir/app.py"
		capture python/module-not-found "$variant" "$dir" "$tool" "python app.py"

		{ pad "$pad_lines"; echo "def main():"; echo "    print(total)"; echo; echo "main()"; } >"$dir/app.py"
		capture python/name-error "$variant" "$dir" "$tool" "python app.py"

		{ pad "$pad_lines"; echo "def add(a, b):"; echo "    return a + b"; echo; echo 'print(add("one", 1))'; } >"$dir/app.py"
		capture python/type-error "$variant" "$dir" "$tool" "python app.py"

		{ pad "$pad_lines"; echo "def broken("; echo "    return 1"; } >"$dir/app.py"
		capture python/syntax-error "$variant" "$dir" "$tool" "python app.py"

		{ pad "$pad_lines"; echo "value = \"text\""; echo "print(value.appendx(1))"; } >"$dir/app.py"
		capture python/attribute-error "$variant" "$dir" "$tool" "python app.py"

		{ pad "$pad_lines"; echo "def ratio(a, b):"; echo "    return a / b"; echo; echo "print(ratio(1, 0))"; } >"$dir/app.py"
		capture python/zero-division "$variant" "$dir" "$tool" "python app.py"

		{ pad "$pad_lines"; echo 'with open("missing-config.toml") as fh:'; echo "    print(fh.read())"; } >"$dir/app.py"
		capture python/file-not-found "$variant" "$dir" "$tool" "python app.py"

		{
			pad "$pad_lines"
			echo "import unittest"
			echo
			echo "class AnswerTest(unittest.TestCase):"
			echo "    def test_answer(self):"
			echo "        self.assertEqual(41, 42)"
			echo
			echo 'if __name__ == "__main__":'
			echo "    unittest.main()"
		} >"$dir/test_answer.py"
		capture python/unittest-failure "$variant" "$dir" "$tool" "python test_answer.py"
	done
}

# ---------------------------------------------------------------- TypeScript and Node

capture_ts() {
	echo "== typescript / node =="
	local tool
	tool="TypeScript $(npx --yes -p typescript tsc --version 2>&1 | tail -1 | tr -d '\r'), $(node --version)"

	local variant dir pad_lines
	for variant in a b; do
		if [ "$variant" = a ]; then
			dir="$WS/ts-alpha"
			pad_lines=0
		else
			dir="$WS/ts-beta"
			pad_lines=5
		fi
		mkdir -p "$dir"
		printf '{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"nodenext","moduleResolution":"nodenext"}}\n' >"$dir/tsconfig.json"

		{ pad "$pad_lines"; echo "function double(n: number): number {"; echo "  return n * 2;"; echo "}"; echo 'console.log(double("three"));'; } >"$dir/app.ts"
		capture typescript/argument-type "$variant" "$dir" "$tool" "npx --yes -p typescript tsc"

		{ pad "$pad_lines"; echo "console.log(missingName);"; } >"$dir/app.ts"
		capture typescript/cannot-find-name "$variant" "$dir" "$tool" "npx --yes -p typescript tsc"

		{ pad "$pad_lines"; echo "const count: number = \"many\";"; echo "console.log(count);"; } >"$dir/app.ts"
		capture typescript/not-assignable "$variant" "$dir" "$tool" "npx --yes -p typescript tsc"

		{ pad "$pad_lines"; echo "function broken( {"; echo "  return 1;"; } >"$dir/app.ts"
		capture typescript/syntax-error "$variant" "$dir" "$tool" "npx --yes -p typescript tsc"

		rm -f "$dir/app.ts"
		{ pad "$pad_lines"; echo 'const lib = require("no-such-package-here");'; echo "console.log(lib);"; } >"$dir/app.js"
		capture node/module-not-found "$variant" "$dir" "$tool" "node app.js"

		{ pad "$pad_lines"; echo "const value = { name: \"x\" };"; echo "console.log(value.compute());"; } >"$dir/app.js"
		capture node/not-a-function "$variant" "$dir" "$tool" "node app.js"

		{ pad "$pad_lines"; echo 'JSON.parse("{ broken: }");'; } >"$dir/app.js"
		capture node/json-parse "$variant" "$dir" "$tool" "node app.js"
	done
}

# ---------------------------------------------------------------- .NET

capture_dotnet() {
	echo "== dotnet =="
	local tool
	tool="dotnet $(dotnet --version 2>&1 | tr -d '\r')"

	local variant dir pad_lines name
	for variant in a b; do
		if [ "$variant" = a ]; then
			name="Alpha"
			pad_lines=0
		else
			name="Beta"
			pad_lines=6
		fi
		dir="$WS/dotnet-$name"
		mkdir -p "$dir"
		if ! (cd "$WS" && dotnet new console -o "dotnet-$name" --force >/dev/null 2>&1); then
			echo "  !! dotnet new failed; skipping" >&2
			return 1
		fi

		{ pad "$pad_lines"; echo "Console.WriteLine(Greet(\"world\"));"; } >"$dir/Program.cs"
		capture dotnet/name-does-not-exist "$variant" "$dir" "$tool" "dotnet build --nologo"

		{ pad "$pad_lines"; echo "int count = \"many\";"; echo "Console.WriteLine(count);"; } >"$dir/Program.cs"
		capture dotnet/cannot-convert "$variant" "$dir" "$tool" "dotnet build --nologo"

		{ pad "$pad_lines"; echo "Console.WriteLine(\"hello\")"; } >"$dir/Program.cs"
		capture dotnet/syntax-error "$variant" "$dir" "$tool" "dotnet build --nologo"

		{
			pad "$pad_lines"
			echo "int Describe(int n)"
			echo "{"
			echo "    if (n > 0) { return 1; }"
			echo "}"
			echo "Console.WriteLine(Describe(1));"
		} >"$dir/Program.cs"
		capture dotnet/not-all-paths-return "$variant" "$dir" "$tool" "dotnet build --nologo"

		{
			pad "$pad_lines"
			echo "string? name = null;"
			echo "Console.WriteLine(name!.Length);"
		} >"$dir/Program.cs"
		capture dotnet/nullreference-runtime "$variant" "$dir" "$tool" "dotnet run --nologo"
	done
}

target="${1:-all}"
case "$target" in
go) capture_go ;;
python) capture_python ;;
ts) capture_ts ;;
dotnet) capture_dotnet ;;
all)
	capture_go
	capture_python
	capture_ts
	capture_dotnet
	;;
*)
	echo "usage: $0 [go|python|ts|dotnet|all]" >&2
	exit 2
	;;
esac

echo
echo "samples: $(find "$OUT" -name '*.txt' | wc -l) files in $(find "$OUT" -mindepth 2 -maxdepth 2 -type d | wc -l) families"
