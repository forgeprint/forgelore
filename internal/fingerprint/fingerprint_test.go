package fingerprint

import (
	"regexp"
	"strings"
	"testing"
)

func TestNormalizeCommand(t *testing.T) {
	cases := []struct {
		command string
		want    string
		why     string
	}{
		{"go build ./...", "go", "the verb is dropped with everything else"},
		{"go run .", "go", "so that one compile error is one fingerprint"},
		{"go test ./...", "go", "whichever verb surfaced it"},
		{"dotnet build --nologo", "dotnet", "flags are dropped"},
		{"node app.js", "node", "a file argument is not the tool"},
		{"python test_answer.py", "python", ""},
		{
			"GOFLAGS=-mod=mod GOPROXY=off go build ./...",
			"go",
			"the environment is not part of the error",
		},
		{"npx --yes -p typescript tsc", "tsc", "the wrapper is stepped over, with its flag's value"},
		{"npx tsc --noEmit", "tsc", "a wrapper with no flags"},

		// A package manager is the tool. "npm run build" names a script,
		// and whatever that script invokes is nowhere on the command
		// line, so stepping over npm used to fingerprint under "run" —
		// splitting one error across every script that can provoke it,
		// which is the thing this function exists to prevent.
		{"npm run build", "npm", "a script name is not a tool"},
		{"npm test", "npm", "neither is a subcommand"},
		{"npm ci", "npm", "nor one of npm's own operations"},
		{"yarn build", "yarn", "the same for yarn"},
		{"pnpm -r build", "pnpm", "flags before the subcommand change nothing"},
		{"npm", "npm", "on its own"},

		// Except where the subcommand does name a binary.
		{"pnpm exec tsc --noEmit", "tsc", "exec hands over"},
		{"yarn dlx tsc", "tsc", "so does dlx"},
		{"bun x tsc", "tsc", "and bun's x"},
		{"npm exec -- tsc", "tsc", "the -- separator is dropped as a flag"},
		{"/usr/local/go/bin/go test ./...", "go", "the install location is not part of the error"},
		{`C:\Go\bin\go.exe build ./...`, "go", "the same tool on Windows"},
		{"", "", "nothing to normalise"},
		{"   ", "", "whitespace only"},

		// A chain is reduced to its last link. An agent writes
		// `cd web && npm run build` far more often than it runs a build
		// on its own, and attributing the compiler's diagnostic to `cd`
		// produced a fingerprint that matched nothing — watched happening
		// in a real Claude Code session on 2026-10-06.
		{"cd /x && go build ./...", "go", "the setup is not the tool"},
		{"ls -a && go build ./... 2>&1 | head -40", "go", "the real miss that found this"},
		{"npm run build && go build ./...", "go", "the last link, not the first"},
		{"a || b && c", "c", "whichever separator came last"},
		{"cd web ; npm test", "npm", "a semicolon chains too"},
		{"go build ./... &&", "go", "a trailing separator chains to nothing"},
		{
			`find . -name '*.tmp' -exec rm {} \;`,
			"find",
			"an escaped semicolon is find's argument, not a chain",
		},
		{"cd /x && CGO_ENABLED=0 npx -p typescript tsc", "tsc", "the rest still applies to the last link"},
	}
	for _, c := range cases {
		if got := NormalizeCommand(c.command); got != c.want {
			t.Errorf("NormalizeCommand(%q) = %q, want %q (%s)", c.command, got, c.want, c.why)
		}
	}
}

// TestSumFormat holds the fingerprint to the shape docs/record-format.md
// gives the field: lowercase hex, and the sixteen characters its examples use.
func TestSumFormat(t *testing.T) {
	events := Scan("go build ./...", "./main.go:5:14: undefined: greet\n")
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if !regexp.MustCompile(`^[0-9a-f]{16}$`).MatchString(events[0].Sum) {
		t.Errorf("fingerprint %q is not sixteen lowercase hex characters", events[0].Sum)
	}
}

// TestSumDependsOnEveryPart guards the join in sum: no two different triples
// may collapse onto one fingerprint because their concatenations happen to
// line up.
func TestSumDependsOnEveryPart(t *testing.T) {
	base := sum("go build", "", "undefined: greet")
	cases := map[string]string{
		"a different command":  sum("go test", "", "undefined: greet"),
		"a different code":     sum("go build", "TS2304", "undefined: greet"),
		"a different message":  sum("go build", "", "undefined: parse"),
		"the parts reshuffled": sum("go", "build", "undefined: greet"),
	}
	for why, got := range cases {
		if got == base {
			t.Errorf("%s produced the same fingerprint %s", why, got)
		}
	}
}

// TestGoPackageHeaderIsIgnored covers the decision that the package a Go error
// was found in does not enter the fingerprint: the fix for "undefined: greet"
// does not depend on which package failed to compile.
func TestGoPackageHeaderIsIgnored(t *testing.T) {
	alpha := Scan("go build ./...", "# alpha\n./main.go:5:14: undefined: greet\n")
	beta := Scan("go build ./...", "# beta\n./main.go:12:14: undefined: greet\n")
	if len(alpha) != 1 || len(beta) != 1 {
		t.Fatalf("got %d and %d events, want 1 each", len(alpha), len(beta))
	}
	if alpha[0].Sum != beta[0].Sum {
		t.Errorf("the package header changed the fingerprint: %s vs %s", alpha[0].Sum, beta[0].Sum)
	}
}

// TestLineEndingsDoNotMatter: the same error captured on Windows and on Unix
// is the same error.
func TestLineEndingsDoNotMatter(t *testing.T) {
	const output = "# alpha\n./main.go:5:14: undefined: greet\n"
	unix := Scan("go build ./...", output)
	windows := Scan("go build ./...", strings.ReplaceAll(output, "\n", "\r\n"))
	if len(unix) != 1 || len(windows) != 1 {
		t.Fatalf("got %d and %d events, want 1 each", len(unix), len(windows))
	}
	if unix[0].Sum != windows[0].Sum {
		t.Errorf("CRLF changed the fingerprint: %s vs %s", unix[0].Sum, windows[0].Sum)
	}
}

// TestNoDiagnostic is the quiet case. A command can fail without printing
// anything this package understands, and then there is nothing to look up.
func TestNoDiagnostic(t *testing.T) {
	cases := map[string]string{
		"empty output":    "",
		"a bare exit":     "exit status 1\n",
		"progress only":   "  Determining projects to restore...\n  All projects are up-to-date for restore.\n",
		"a build summary": "FAIL\nFAIL\talpha\t0.671s\nFAIL\n",
		"a banner":        "Node.js v22.12.0\n",

		// npm prefixes every line of a failure with its own name, and only
		// the code line is kept. These are the lines that must not be: the
		// log path carries a timestamp and would hash differently on every
		// run, and the others say nothing at all.
		"npm's log path":     "npm error A complete log of this run can be found in: /x/_logs/2026-10-06T03_23_50_645Z-debug-0.log\n",
		"npm's blank line":   "npm error\n",
		"npm's continuation": "npm error   npm run\n",
		"a warning":          "warning: this package is deprecated\n",
		"Error inside prose": "The build had an Error somewhere in it\n",
	}
	for why, output := range cases {
		if got := Scan("go build ./...", output); len(got) != 0 {
			t.Errorf("%s produced %d events, want none: %+v", why, len(got), got)
		}
	}
}

// TestTruncatedTracebackIsNotADiagnostic: a traceback cut off before its
// exception line says where the program was, not what went wrong. Guessing
// from the frames would attach a memory to a workspace path.
func TestTruncatedTracebackIsNotADiagnostic(t *testing.T) {
	output := "Traceback (most recent call last):\n" +
		`  File "/work/app.py", line 2, in <module>` + "\n" +
		"    print(value.appendx(1))\n"
	if got := Scan("python app.py", output); len(got) != 0 {
		t.Errorf("got %d events, want none: %+v", len(got), got)
	}
}

// TestCLIErrorsAreRecognised covers the shapes that have no file and no
// line, and that the extractor was blind to until 2026-10-06.
//
// Every string here was copied from a real failure met while building this
// project, which is also how the gap was found: Forgelore could not record
// the errors of its own build day. Two are in the corpus as captured
// families; the rest cannot be provoked on demand — an expired OAuth
// session, an account a provider has stopped serving — so they are held
// here rather than invented into testdata.
func TestCLIErrorsAreRecognised(t *testing.T) {
	cases := []struct {
		why     string
		output  string
		message string
	}{
		{
			"a package manager's code, with the prose around it dropped",
			"npm error code E403\n" +
				"npm error 403 403 Forbidden - PUT https://registry.npmjs.org/x - spam detection\n" +
				"npm error A complete log of this run can be found in: /x/_logs/2026-10-06T03_23_50_645Z-debug-0.log\n",
			"code E403",
		},
		{
			"a one-time password demand",
			"npm error code EOTP\nnpm error This operation requires a one-time password.\n",
			"code EOTP",
		},
		{
			"git, which says fatal rather than error",
			"fatal: not a git repository (or any of the parent directories): .git\n",
			"not a git repository (or any of the parent directories): .git",
		},
		{
			"git's other half",
			"error: pathspec 'nope' did not match any file(s) known to git\n",
			"pathspec 'nope' did not match any file(s) known to git",
		},
		{
			"an exception behind a label, with the label dropped",
			"Error authenticating: IneligibleTierError: This client is no longer supported.\n",
			"IneligibleTierError: This client is no longer supported.",
		},
		{
			"a sentence, which is all some tools give",
			"Failed to authenticate: OAuth session expired and could not be refreshed\n",
			"Failed to authenticate: OAuth session expired and could not be refreshed",
		},
	}
	for _, c := range cases {
		got := Scan("x", c.output)
		if len(got) != 1 {
			t.Errorf("%s: got %d events, want 1: %+v", c.why, len(got), got)
			continue
		}
		if got[0].Message != c.message {
			t.Errorf("%s:\n  got  %q\n  want %q", c.why, got[0].Message, c.message)
		}
	}
}

// TestALabelDoesNotChangeTheFingerprint: the same exception reported bare
// and reported behind a caller's label is one error. Keeping the label
// would split it by whoever happened to print it.
func TestALabelDoesNotChangeTheFingerprint(t *testing.T) {
	bare := Scan("gemini", "IneligibleTierError: This client is no longer supported.\n")
	labelled := Scan("gemini", "Error authenticating: IneligibleTierError: This client is no longer supported.\n")
	if len(bare) != 1 || len(labelled) != 1 {
		t.Fatalf("got %d and %d events", len(bare), len(labelled))
	}
	if bare[0].Sum != labelled[0].Sum {
		t.Errorf("%s with a label, %s without", labelled[0].Sum, bare[0].Sum)
	}
}

// TestStackFramesAreNotDiagnostics: a frame carries a file, a line and a
// column, and is still not an error. Only the first token on a line is ever
// read as a position, which is what keeps these out.
func TestStackFramesAreNotDiagnostics(t *testing.T) {
	output := "TypeError: value.compute is not a function\n" +
		`    at Object.<anonymous> (/work/app.js:2:19)` + "\n" +
		"    at Module._compile (node:internal/modules/cjs/loader:1565:14)\n" +
		"\n" +
		"Node.js v22.12.0\n"
	got := Scan("node app.js", output)
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1: %+v", len(got), got)
	}
	if got[0].Message != "TypeError: value.compute is not a function" {
		t.Errorf("got %q", got[0].Message)
	}
}

// TestCodeAttachesToItsOwnError: Node prints the code after the message, and
// it must land on that message rather than on an unrelated later one.
func TestCodeAttachesToItsOwnError(t *testing.T) {
	output := "Error: Cannot find module 'nope'\n" +
		"    at Function._resolveFilename (node:internal/modules/cjs/loader:1249:15)\n" +
		"  code: 'MODULE_NOT_FOUND',\n"
	got := Scan("node app.js", output)
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1: %+v", len(got), got)
	}
	if got[0].Code != "MODULE_NOT_FOUND" {
		t.Errorf("got code %q, want MODULE_NOT_FOUND", got[0].Code)
	}
}

// TestVerbDoesNotSplitAnError is the point of reducing a command to its tool.
// One compile error reaches the user through whichever verb happened to
// invoke the compiler, and a fix recorded under one of them has to be found
// under the others.
func TestVerbDoesNotSplitAnError(t *testing.T) {
	const output = "# alpha\n./main.go:5:14: undefined: greet\n"
	verbs := []string{"go build ./...", "go test ./...", "go run .", "go vet ./..."}

	var first string
	for _, command := range verbs {
		events := Scan(command, output)
		if len(events) != 1 {
			t.Fatalf("%q: got %d events, want 1", command, len(events))
		}
		if first == "" {
			first = events[0].Sum
			continue
		}
		if events[0].Sum != first {
			t.Errorf("%q fingerprinted %s, but %q fingerprinted %s",
				command, events[0].Sum, verbs[0], first)
		}
	}
}

// TestDifferentToolsStaySeparate is the other side of it: the tool itself
// still counts, so two tools that happen to print the same words do not share
// a memory.
func TestDifferentToolsStaySeparate(t *testing.T) {
	const output = "SyntaxError: '(' was never closed\n"
	python := Scan("python app.py", output)
	node := Scan("node app.js", output)
	if len(python) != 1 || len(node) != 1 {
		t.Fatalf("got %d and %d events, want 1 each", len(python), len(node))
	}
	if python[0].Sum == node[0].Sum {
		t.Errorf("python and node share the fingerprint %s", python[0].Sum)
	}
}
