// Package candidate holds fixes Forgelore noticed but did not record.
//
// A candidate is the gap between observing that an error stopped and knowing
// why. Hooks produce them when a failing command starts working; the MCP
// `propose` tool produces them when an agent believes it has understood a
// fix. Neither becomes a memory without a person saying so, which is what
// `forgelore review` is for.
//
// They live in the local scope, so they are never shared by accident.
package candidate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// file is where candidates accumulate inside a .forgelore directory.
const file = "local/candidates.jsonl"

// A Candidate is a proposed memory.
type Candidate struct {
	Time        time.Time `json:"time"`
	Session     string    `json:"session,omitempty"`
	Fingerprint string    `json:"fingerprint"`
	Command     string    `json:"command,omitempty"`

	// Title is set when whoever proposed it could say what the fix was. A
	// hook cannot: it saw an error stop, not a reason. An agent calling
	// `propose` can, and `review` then has something to accept.
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`

	// Tainted marks a proposal derived from content outside the project
	// (ADR-0013). It carries through to the record if one is made.
	Tainted bool `json:"tainted,omitempty"`
}

// Path returns the candidates file for a store root.
func Path(root string) string { return filepath.Join(root, file) }

// Append adds one candidate.
func Append(root string, c Candidate) error {
	path := Path(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	line, err := json.Marshal(c)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(line, '\n'))
	return err
}

// Read returns the waiting candidates. A line that cannot be parsed is
// skipped: one bad line must not hide the rest.
func Read(root string) ([]Candidate, error) {
	data, err := os.ReadFile(Path(root))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Candidate
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var c Candidate
		if json.Unmarshal([]byte(line), &c) == nil {
			out = append(out, c)
		}
	}
	return out, nil
}

// Write replaces the file, removing it when nothing is left.
func Write(root string, cs []Candidate) error {
	path := Path(root)
	if len(cs) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	var b strings.Builder
	for _, c := range cs {
		line, err := json.Marshal(c)
		if err != nil {
			return err
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// Take removes the candidate with this fingerprint and returns it with what
// remains.
func Take(cs []Candidate, fingerprint string) (taken *Candidate, rest []Candidate, err error) {
	for _, c := range cs {
		if c.Fingerprint == fingerprint && taken == nil {
			found := c
			taken = &found
			continue
		}
		rest = append(rest, c)
	}
	if taken == nil {
		return nil, cs, fmt.Errorf("no candidate with fingerprint %s", fingerprint)
	}
	return taken, rest, nil
}
