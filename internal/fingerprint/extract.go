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
	//
	// One label may sit in front of it, and only the text from the
	// identifier onwards is kept. The Gemini CLI writes
	//
	//	Error authenticating: IneligibleTierError: This client is no longer…
	//
	// and the label is the caller's, not the error's: keeping it would make
	// the same failure hash differently depending on who reported it.
	exceptionLine = regexp.MustCompile(`^(?:[A-Za-z][^:]{0,40}: )?((?:[A-Za-z_$][\w$.]*?)?(?:Error|Exception): .+)$`)

	// A shell refusing to parse the command it was given. Everything
	// before the phrase is position — the shell's own name, the script,
	// a line number — and everything after it is the error, so the
	// message starts at the phrase and the token it names is what
	// separates one of these from another.
	//
	// The phrase has to be preceded by the start of the line or by a
	// colon and a space, which is what keeps it from firing on prose:
	// "that was a syntax error near the top" has a space in front of it
	// and is a sentence, not a diagnostic. Capitals are allowed because
	// not every shell lower-cases the word; the two forms captured in the
	// corpus are the ones /bin/sh produced on the machine that captured
	// them.
	//
	// zsh reaches the same message through fileLineDiag instead, because
	// it writes "probe.sh:3: parse error near …" with no space before the
	// line number. Both roads arrive at the same text.
	shellSyntaxError = regexp.MustCompile(`(?:^|: )((?:[Ss]yntax|[Pp]arse) error(?: near | ?: )\S.*)$`)

	// A package manager's error code. npm, yarn and pnpm prefix every line
	// of a failure with their own name, and only one of those lines is
	// worth keeping: the code. The prose lines carry a log path with a
	// timestamp in it, which would hash differently on every run — this
	// matcher is narrow on purpose, and the cost is that an npm failure
	// with no code line is not recognised at all.
	toolErrorCode = regexp.MustCompile(`^\s*[a-z][a-z0-9_.-]* (?:error|ERR!) code ([A-Za-z][A-Za-z0-9_]+)\s*$`)

	// The shape git, cargo, rustc and clang use. The bracketed code is
	// rustc's: "error[E0433]: failed to resolve".
	cliErrorLine = regexp.MustCompile(`^\s*(?:error|fatal)(\[[A-Za-z0-9_-]+\])?: (.+)$`)

	// A sentence a CLI writes when an operation did not happen. It is the
	// loosest matcher here, and it earns its place: "Failed to
	// authenticate: OAuth session expired and could not be refreshed" is
	// the shape of half the operational failures this project met while
	// being built, and none of the older matchers saw any of them.
	failedToLine = regexp.MustCompile(`^\s*((?:Failed|Unable|Cannot) to .+)$`)

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

		if m := shellSyntaxError.FindStringSubmatch(line); m != nil {
			out = append(out, diag{message: normaliseMessage(m[1])})
			continue
		}

		if m := dotnetUnhandled.FindStringSubmatch(line); m != nil {
			out = append(out, diag{message: normaliseMessage(m[1])})
			continue
		}

		if m := toolErrorCode.FindStringSubmatch(line); m != nil {
			out = append(out, diag{message: "code " + m[1]})
			continue
		}

		if m := cliErrorLine.FindStringSubmatch(line); m != nil {
			out = append(out, diag{code: strings.Trim(m[1], "[]"), message: normaliseMessage(m[2])})
			continue
		}

		if m := exceptionLine.FindStringSubmatch(line); m != nil {
			out = append(out, diag{message: normaliseMessage(m[1])})
			continue
		}

		if goPanic.MatchString(line) || unittestFail.MatchString(line) {
			out = append(out, diag{message: normaliseMessage(line)})
			continue
		}

		if m := failedToLine.FindStringSubmatch(line); m != nil {
			out = append(out, diag{message: normaliseMessage(m[1])})
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
