// Package measure records what Forgelore spends and what the session spent,
// so that the saving it claims is an observation rather than a promise
// (ADR-0012).
//
// Two append-only logs live under .forgelore/ledger, one file per month:
//
//	injections-YYYY-MM.jsonl   one line per error looked up
//	usage-YYYY-MM.jsonl        one line per usage report from the agent
//
// They are not in the SQLite index on purpose. The index is derived and is
// deleted and rebuilt whenever it cannot be read; the ledger is the only copy
// of what happened and could not survive that. They are not records either:
// a record is a memory someone wrote, and these are measurements.
//
// The ledger is local and never leaves the machine.
package measure

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Dir is the ledger directory inside a .forgelore directory.
const Dir = "ledger"

// What happened to one error.
const (
	// EventInject: a memory matched and its hint was handed to the agent.
	EventInject = "inject"
	// EventControl: a memory matched but the session is in the control
	// group, so nothing was injected. The cost that was avoided is still
	// recorded, which is what makes the two groups comparable.
	EventControl = "control"
	// EventMiss: the error was fingerprinted and nothing matched. Free.
	EventMiss = "miss"
)

// The two arms of the trial.
const (
	GroupTreatment = "treatment"
	GroupControl   = "control"
)

// Entry is one error looked up.
type Entry struct {
	Time        time.Time `json:"time"`
	Session     string    `json:"session"`
	Group       string    `json:"group"`
	Event       string    `json:"event"`
	Fingerprint string    `json:"fingerprint"`
	RecordID    string    `json:"record_id,omitempty"`
	Bytes       int       `json:"bytes"`
	EstTokens   int       `json:"est_tokens"`
}

// Usage is one report of what a session has spent so far.
//
// CostUSD is cumulative over the session and is the figure the report
// compares, because it is the only cumulative number Claude Code documents.
// The context window token counts are not cumulative — they describe the
// window at the moment of the reading — so they are kept for context and
// never summed.
type Usage struct {
	Time         time.Time `json:"time"`
	Session      string    `json:"session"`
	Group        string    `json:"group"`
	Agent        string    `json:"agent"`
	AgentVersion string    `json:"agent_version,omitempty"`
	CostUSD      float64   `json:"cost_usd"`
	DurationMS   int64     `json:"duration_ms,omitempty"`
	ContextIn    int64     `json:"context_input_tokens,omitempty"`
	ContextOut   int64     `json:"context_output_tokens,omitempty"`
}

// AppendEntry adds one line to the month's injection log.
func AppendEntry(root string, e Entry) error {
	return appendLine(root, "injections", e.Time, e)
}

// AppendUsage adds one line to the month's usage log.
func AppendUsage(root string, u Usage) error {
	return appendLine(root, "usage", u.Time, u)
}

// appendLine writes one JSON object followed by a newline.
//
// The file is opened with O_APPEND and the line is written in a single call,
// which is what keeps two Forgelore processes — a hook and a status line can
// easily fire at once — from interleaving half-lines.
func appendLine(root, kind string, when time.Time, v any) error {
	dir := filepath.Join(root, Dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("measure: creating %s: %w", dir, err)
	}

	line, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("measure: encoding a ledger line: %w", err)
	}
	line = append(line, '\n')

	path := filepath.Join(dir, fmt.Sprintf("%s-%s.jsonl", kind, when.UTC().Format("2006-01")))
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("measure: opening %s: %w", path, err)
	}
	defer f.Close()
	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("measure: writing %s: %w", path, err)
	}
	return nil
}

// Ledger is everything read back for a period, with what could not be read.
type Ledger struct {
	Entries []Entry
	Usage   []Usage

	// Unreadable counts lines that could not be parsed. A truncated last
	// line is the normal way this happens — a process died mid-write — and
	// it must not stop a report, so the line is counted and skipped.
	Unreadable int
}

// Read returns the ledger lines in [since, until).
func Read(root string, since, until time.Time) (*Ledger, error) {
	dir := filepath.Join(root, Dir)
	names, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return &Ledger{}, nil
		}
		return nil, fmt.Errorf("measure: reading %s: %w", dir, err)
	}

	l := &Ledger{}
	for _, name := range names {
		if name.IsDir() || !strings.HasSuffix(name.Name(), ".jsonl") {
			continue
		}
		path := filepath.Join(dir, name.Name())
		switch {
		case strings.HasPrefix(name.Name(), "injections-"):
			err = scanLines(path, func(line []byte) {
				var e Entry
				if json.Unmarshal(line, &e) != nil {
					l.Unreadable++
					return
				}
				if inRange(e.Time, since, until) {
					l.Entries = append(l.Entries, e)
				}
			})
		case strings.HasPrefix(name.Name(), "usage-"):
			err = scanLines(path, func(line []byte) {
				var u Usage
				if json.Unmarshal(line, &u) != nil {
					l.Unreadable++
					return
				}
				if inRange(u.Time, since, until) {
					l.Usage = append(l.Usage, u)
				}
			})
		default:
			continue
		}
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(l.Entries, func(i, j int) bool { return l.Entries[i].Time.Before(l.Entries[j].Time) })
	sort.Slice(l.Usage, func(i, j int) bool { return l.Usage[i].Time.Before(l.Usage[j].Time) })
	return l, nil
}

func inRange(t, since, until time.Time) bool {
	return !t.Before(since) && t.Before(until)
}

// scanLines calls fn for each non-empty line. A read error ends the file
// quietly: a report built from most of the ledger is worth more than no
// report, and the count of unreadable lines is printed either way.
func scanLines(path string, fn func([]byte)) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("measure: opening %s: %w", path, err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	// A ledger line is small, but a long record title could push one past
	// the default limit, and a skipped line would be a lost measurement.
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for s.Scan() {
		if line := s.Bytes(); len(bytes.TrimSpace(line)) > 0 {
			fn(line)
		}
	}
	return s.Err()
}
