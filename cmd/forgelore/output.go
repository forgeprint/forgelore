package main

import (
	"encoding/json"
	"io"
	"strings"
)

// writeJSON emits one indented object per command. Indented rather than
// compact: these outputs are read by people at least as often as by programs,
// and a diff of two runs should be readable.
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// hitJSON is a search result. It carries an id, a type and a title and
// nothing else: staged access is enforced by what the caller cannot get
// without asking for it by id (ADR-0006).
type hitJSON struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Scope      string `json:"scope"`
	Superseded bool   `json:"superseded,omitempty"`
}

// splitList parses a comma-separated flag value. Empty entries are dropped so
// that a trailing comma is not an error worth stopping for.
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

// indent prefixes every line, for quoting a block inside a sentence.
func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
