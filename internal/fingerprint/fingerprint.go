// Package fingerprint turns a failed command's output into stable identifiers.
//
// A failing tool prints more than one diagnostic per run, and every diagnostic
// carries detail that changes between runs and between machines: absolute
// paths, line and column numbers, stack frames, timings, build durations. Scan
// pulls the diagnostics out of that noise and reduces each one to a
// fingerprint that survives the differences, so the same error met twice — in
// another workspace, on another machine — looks up the same memory.
//
// The rules here are calibrated against testdata/errors, a corpus of real
// output from real failing commands. A rule that cannot be demonstrated on
// that corpus does not belong in this package.
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// An Event is one diagnostic found in a command's output.
type Event struct {
	// Command is the normalised command: the tool and, where the tool has
	// one, its verb. "GOFLAGS=-mod=mod GOPROXY=off go build ./..." reduces to
	// "go build".
	Command string

	// Code is the tool's own number for the diagnostic, such as CS0103 or
	// TS2345. Tools that do not number their diagnostics leave it empty.
	Code string

	// Message is the diagnostic text with the varying detail removed.
	Message string

	// Sum is the fingerprint: sixteen lowercase hex characters, the form
	// docs/record-format.md gives for the `fingerprint` field.
	Sum string
}

// Scan extracts the diagnostics from a command's combined output. The events
// come back in the order they were printed, with exact repeats dropped: a
// tool that prints the same diagnostic twice — MSBuild prints each error
// again in its summary — describes one error, not two.
//
// Output that holds no recognisable diagnostic yields no events. That is the
// quiet case and it is the common one: a command can fail without printing
// anything this package understands, and inventing a fingerprint for an
// unparsed blob would attach memories to noise.
func Scan(command, output string) []Event {
	normalised := NormalizeCommand(command)

	var events []Event
	seen := make(map[string]bool)
	for _, d := range extract(output) {
		e := Event{Command: normalised, Code: d.code, Message: d.message}
		e.Sum = sum(e.Command, e.Code, e.Message)
		if seen[e.Sum] {
			continue
		}
		seen[e.Sum] = true
		events = append(events, e)
	}
	return events
}

// sum hashes the three fields that identify an error. The parts are joined
// with a newline, which none of them can contain, so no two different triples
// can produce the same input string.
//
// Sixteen hex characters is half of a SHA-256. The full digest buys nothing
// here: a fingerprint is a lookup key in one project's index, not a defence
// against someone constructing a collision, and a short key keeps the record
// files readable.
func sum(command, code, message string) string {
	h := sha256.Sum256([]byte(command + "\n" + code + "\n" + message))
	return hex.EncodeToString(h[:8])
}

// splitLines accepts either line ending. The corpus is stored with LF, but
// output captured on Windows arrives with CRLF and must fingerprint the same.
func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, "\r")
	}
	return lines
}
