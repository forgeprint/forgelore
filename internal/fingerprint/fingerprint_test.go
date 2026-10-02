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
		{"go build ./...", "go build", "tool and verb"},
		{"go run .", "go run", "a verb followed by a path"},
		{"go test ./...", "go test", ""},
		{"dotnet build --nologo", "dotnet build", "flags are dropped"},
		{"node app.js", "node", "a file argument is not a verb"},
		{"python test_answer.py", "python", ""},
		{
			"GOFLAGS=-mod=mod GOPROXY=off go build ./...",
			"go build",
			"the environment is not part of the error",
		},
		{"npx --yes -p typescript tsc", "tsc", "the wrapper is stepped over, with its flag's value"},
		{"npx tsc --noEmit", "tsc", "a wrapper with no flags"},
		{"/usr/local/go/bin/go test ./...", "go test", "the install location is not part of the error"},
		{`C:\Go\bin\go.exe build ./...`, "go build", "the same tool on Windows"},
		{"", "", "nothing to normalise"},
		{"   ", "", "whitespace only"},
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
