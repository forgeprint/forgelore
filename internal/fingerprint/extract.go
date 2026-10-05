package fingerprint

import (
	"regexp"
	"strings"
)

// diag is one diagnostic as it was found in the output, before the command is
// attached and the fingerprint is taken.
type diag struct {
	code    string
	message string
}

// The matchers below recognise the shapes a diagnostic takes. Everything a
// matcher does not recognise is noise and is dropped, which is why there is no
// list of noise patterns: progress lines, separators, counts, durations,
// banners and stack frames are all simply not diagnostics.
//
// Each pattern is anchored at the start of the line. That anchoring does more
// work than it looks: a stack frame such as
//
//	at Object.<anonymous> (C:\work\app.js:1:13)
//
// holds a file, a line and a column, but they are not the first token on the
// line, so fileLineDiag cannot reach them.
var (
	// MSBuild and tsc share one shape: file(line,col): error CODE: message.
	codedDiag = regexp.MustCompile(`^\s*\S.*?\(\d+,\d+\): error ([A-Za-z]+\d+): (.+)$`)

	// MSBuild repeats the project file after the message.
	msbuildProject = regexp.MustCompile(`\s*\[[^\[\]]*\.(?:csproj|vbproj|fsproj|proj|sln)\]$`)

	// Compiler diagnostics keyed to a position: file:line:col: message, or
	// file:line: message. The file must carry an extension, which is what
	// keeps "goroutine 1 [running]:" and "node:internal/modules/cjs/loader"
	// out.
	//
	// The optional word before the position is for a tool that labels its
	// own output. `go vet` prints "vet: ./main.go:3:2: undefined: greet"
	// for the error `go build` prints bare, and without this the whole vet
	// path was invisible — which defeated the point of leaving the verb out
	// of a fingerprint in the first place.
	fileLineDiag = regexp.MustCompile(`^\s*(?:[A-Za-z][A-Za-z0-9_-]*:\s+)?\S*?[^\s:]\.[A-Za-z0-9_+-]+:\d+(?::\d+)?: (.+)$`)

	// go test's per-test header. The duration is dropped: it is different
	// every run.
	goTestFail = regexp.MustCompile(`^\s*--- (?:FAIL|ERROR): (\S+) \([0-9.]+m?s\)$`)

	// A Go panic or runtime fatal error.
	goPanic = regexp.MustCompile(`^(?:panic|fatal error): .+$`)

	// A .NET unhandled exception, whose own prefix is dropped so that the
	// exception reads the same as it would anywhere else.
	dotnetUnhandled = regexp.MustCompile(`^Unhandled exception\. (.+)$`)

	// A JavaScript or .NET exception header: an optionally dotted name
	// ending in Error or Exception. The optional prefix is what lets a bare
	// "Error: ..." match alongside "TypeError: ..." and
	// "System.NullReferenceException: ...".
	exceptionLine = regexp.MustCompile(`^(?:[A-Za-z_$][\w$.]*?)?(?:Error|Exception): .+$`)

	// Python's unittest failure header.
	unittestFail = regexp.MustCompile(`^(?:FAIL|ERROR): .+$`)

	// Node attaches a code to the error it just printed.
	nodeErrorCode = regexp.MustCompile(`^\s*code: '([A-Za-z0-9_]+)',?$`)

	// A Python traceback frame. Its path, line number and function name are
	// all workspace detail, and the two lines that follow it echo the
	// source; none of it reaches the fingerprint.
	pyFrame = regexp.MustCompile(`^\s+File "[^"]*", line \d+`)

	spaceRun = regexp.MustCompile(`\s+`)
)

const pyTracebackHeader = "Traceback (most recent call last):"

// extract walks the output once and returns the diagnostics it recognises.
func extract(output string) []diag {
	lines := splitLines(output)

	var out []diag
	for i := 0; i < len(lines); i++ {
		if next, d, ok := pythonTraceback(lines, i); ok {
			if d.message != "" {
				out = append(out, d)
			}
			i = next - 1
			continue
		}

		line := lines[i]

		if m := nodeErrorCode.FindStringSubmatch(line); m != nil {
			// A code belongs to the diagnostic it follows. With nothing to
			// attach it to there is nothing to record.
			if n := len(out); n > 0 && out[n-1].code == "" {
				out[n-1].code = m[1]
			}
			continue
		}

		if m := codedDiag.FindStringSubmatch(line); m != nil {
			message := msbuildProject.ReplaceAllString(m[2], "")
			out = append(out, diag{code: m[1], message: normaliseMessage(message)})
			continue
		}

		if m := goTestFail.FindStringSubmatch(line); m != nil {
			out = append(out, diag{message: "--- FAIL: " + m[1]})
			continue
		}

		if m := fileLineDiag.FindStringSubmatch(line); m != nil {
			out = append(out, diag{message: normaliseMessage(m[1])})
			continue
		}

		if m := dotnetUnhandled.FindStringSubmatch(line); m != nil {
			out = append(out, diag{message: normaliseMessage(m[1])})
			continue
		}

		if goPanic.MatchString(line) || exceptionLine.MatchString(line) || unittestFail.MatchString(line) {
			out = append(out, diag{message: normaliseMessage(line)})
			continue
		}
	}
	return out
}

// pythonTraceback recognises a traceback beginning at lines[i] and returns the
// exception line that closes it, which is the first line back at column zero.
//
// ok reports that a traceback was found, whether or not it was closed, so that
// the caller skips the frames either way. A traceback cut off by a timeout or
// a truncated log has no exception line and yields no diagnostic: the frames
// alone say where the program was, not what went wrong.
func pythonTraceback(lines []string, i int) (next int, d diag, ok bool) {
	if lines[i] != pyTracebackHeader && !pyFrame.MatchString(lines[i]) {
		return 0, diag{}, false
	}
	for j := i + 1; j < len(lines); j++ {
		line := lines[j]
		if strings.TrimSpace(line) == "" {
			continue
		}
		if line[0] == ' ' || line[0] == '\t' {
			continue
		}
		return j + 1, diag{message: normaliseMessage(line)}, true
	}
	return len(lines), diag{}, true
}

// normaliseMessage collapses runs of whitespace, so that a message reflowed or
// re-indented by a wrapping tool still hashes the same.
func normaliseMessage(s string) string {
	return strings.TrimSpace(spaceRun.ReplaceAllString(s, " "))
}
