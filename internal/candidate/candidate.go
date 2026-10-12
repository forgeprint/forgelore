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

// Kinds of candidate. A candidate written before work memory existed has no
// kind at all, and an empty kind means KindFix: every candidate there was.
const (
	KindFix  = "fix"
	KindWork = "work"
)

// A Candidate is a proposed memory.
type Candidate struct {
	Time        time.Time `json:"time"`
	Session     string    `json:"session,omitempty"`
	Fingerprint string    `json:"fingerprint"`
	Command     string    `json:"command,omitempty"`

	// Kind separates a fix noticed when an error stopped from an account of
	// what a session did. They wait in the same queue and are reviewed by
	// the same command, but they are not keyed the same way: a fix is keyed
	// by the error it answers, and a session's work has no error.
	Kind string `json:"kind,omitempty"`

	// Title is set when whoever proposed it could say what the fix was. A
	// hook cannot: it saw an error stop, not a reason. An agent calling
	// `propose` can, and `review` then has something to accept.
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`

	// Source is who wrote this candidate's title, when anybody did. Empty
	// means nobody has: a hook cannot write one. `distill` means a model
	// did, through the user's own CLI, and that reaches the record so a
	// sentence a model wrote is never mistaken for one a person wrote
	// (ADR-0027).
	Source string `json:"source,omitempty"`

	// Tainted marks a proposal derived from content outside the project
	// (ADR-0013). It carries through to the record if one is made.
	Tainted bool `json:"tainted,omitempty"`
}

// Key is how a person names this candidate to `review`. A fix is keyed by its
// fingerprint, which is the thing it answers. Work has no fingerprint, so it
// is keyed by the session it came out of — a value the agent already chose,
// which is readable in a way a second identifier would not be.
func (c Candidate) Key() string {
	if c.Fingerprint != "" {
		return c.Fingerprint
	}
	return c.Session
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

// Take removes the candidate with this key and returns it with what remains.
func Take(cs []Candidate, key string) (taken *Candidate, rest []Candidate, err error) {
	for _, c := range cs {
		if c.Key() == key && taken == nil {
			found := c
			taken = &found
			continue
		}
		rest = append(rest, c)
	}
	if taken == nil {
		return nil, cs, fmt.Errorf("no candidate with key %s", key)
	}
	return taken, rest, nil
}

// Upsert adds a candidate, or folds it into the one already waiting under the
// same key and kind.
//
// Work memory arrives in two parts and at two moments — what the agent said
// about the session, and the facts read out of its transcript afterwards — and
// two entries in the review queue for one session would be two decisions for
// one thing. Appending text that is already there is skipped, so running the
// same command twice changes nothing.
func Upsert(root string, c Candidate) error {
	waiting, err := Read(root)
	if err != nil {
		return err
	}
	for i, old := range waiting {
		if old.Key() != c.Key() || old.Kind != c.Kind {
			continue
		}
		merged := old
		merged.Body = join(old.Body, c.Body)
		if merged.Title == "" {
			merged.Title = c.Title
		}
		merged.Tainted = old.Tainted || c.Tainted
		waiting[i] = merged
		return Write(root, waiting)
	}
	return Append(root, c)
}

// join puts two bodies together without repeating what is already said.
func join(old, add string) string {
	add = strings.TrimSpace(add)
	switch {
	case add == "":
		return old
	case strings.TrimSpace(old) == "":
		return add
	case strings.Contains(old, add):
		return old
	}
	return old + "\n\n" + add
}
