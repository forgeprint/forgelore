package main

import (
	"fmt"
	"time"

	"github.com/forgeprint/forgelore/internal/fingerprint"
	"github.com/forgeprint/forgelore/internal/measure"
	"github.com/forgeprint/forgelore/internal/store"
)

// recallMatch is one error found in the output, with whatever is known about
// it. An error with no matches is still reported: knowing that nothing is
// known is the answer to the question.
type recallMatch struct {
	Fingerprint string    `json:"fingerprint"`
	Command     string    `json:"command"`
	Code        string    `json:"code,omitempty"`
	Message     string    `json:"message"`
	Matches     []hitJSON `json:"matches"`
}

// cmdRecall is the injection path: an error has just happened and the
// question is whether it has happened before.
//
// It never fails on a miss and never fails on a store it cannot read. A tool
// in this position that returns an error teaches the agent to stop calling it
// (ADR-0007), and an agent that stops calling it loses the memory entirely.
func cmdRecall(e env, args []string) error {
	fs := newFlags(e, "recall")
	command := fs.String("command", "", "the command that failed")
	errorFile := fs.String("error-file", "-", "the captured output; - for stdin")
	session := fs.String("session", "", "the agent's session id, for the measurement ledger")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	output, err := readFileOrStdin(e, *errorFile)
	if err != nil {
		return err
	}
	events := fingerprint.Scan(*command, output)

	results := make([]recallMatch, 0, len(events))
	for _, ev := range events {
		results = append(results, recallMatch{
			Fingerprint: ev.Sum,
			Command:     ev.Command,
			Code:        ev.Code,
			Message:     ev.Message,
			Matches:     []hitJSON{},
		})
	}

	// Everything above works without a store. Only the lookup needs one, so
	// a project that has not been initialised still gets its errors
	// fingerprinted rather than an error message.
	if s, err := openStore(e, *dir); err == nil {
		cfg, _ := loadConfig(e, s, nil)
		group := measure.Assign(
			cfg.String("measure.ab.salt"),
			*session,
			int(cfg.Int("measure.ab.control_percent")),
		)

		if idx, err := openIndex(s); err == nil {
			defer idx.Close()
			for i := range results {
				hits, err := idx.ByFingerprint(results[i].Fingerprint)
				if err != nil {
					continue
				}
				results[i].Matches = toHitJSON(hits)
			}
		}

		// In the control arm the matches are dropped before anything is
		// printed, so a control session is indistinguishable from a miss.
		// The ledger still records that a match existed, which is what
		// makes the two arms comparable; `report` is where a human sees
		// the difference.
		hadMatch := make([]bool, len(results))
		for i := range results {
			hadMatch[i] = len(results[i].Matches) > 0
		}
		if group == measure.GroupControl {
			for i := range results {
				results[i].Matches = []hitJSON{}
			}
		}

		if cfg.Bool("measure.ledger") {
			logRecall(s, *session, group, results, hadMatch)
		}
	}

	if *asJSON {
		return writeJSON(e.stdout, struct {
			Errors []recallMatch `json:"errors"`
		}{results})
	}
	return printRecall(e, results)
}

func printRecall(e env, results []recallMatch) error {
	if len(results) == 0 {
		fmt.Fprintln(e.stdout, "no error recognised in that output")
		return nil
	}

	var known int
	for _, r := range results {
		if len(r.Matches) > 0 {
			known++
		}
	}
	fmt.Fprintf(e.stdout, "%d error(s), %d with something recorded\n", len(results), known)

	for _, r := range results {
		fmt.Fprintf(e.stdout, "\n%s  %s\n", r.Fingerprint, r.Message)
		if len(r.Matches) == 0 {
			fmt.Fprintln(e.stdout, "  nothing recorded")
			continue
		}
		for _, m := range r.Matches {
			fmt.Fprintf(e.stdout, "  %-9s %s  %s\n", m.Type, m.ID, m.Title)
		}
	}
	return nil
}

func toHitJSON(hits []store.Hit) []hitJSON {
	out := make([]hitJSON, 0, len(hits))
	for _, h := range hits {
		out = append(out, hitJSON{
			ID:         h.ID,
			Type:       h.Type,
			Title:      h.Title,
			Scope:      string(h.Scope),
			Superseded: h.Superseded,
		})
	}
	return out
}

// logRecall writes one ledger line per error looked up.
//
// The bytes counted are the bytes a treatment session would receive: the
// titles that make up the hint. A control session is charged nothing, because
// nothing was handed over, but the line still says a match was there.
func logRecall(s *store.Store, session, group string, results []recallMatch, hadMatch []bool) {
	root := ledgerRoot(s)
	now := time.Now().UTC().Truncate(time.Second)

	for i, r := range results {
		e := measure.Entry{
			Time:        now,
			Session:     session,
			Group:       group,
			Event:       measure.EventMiss,
			Fingerprint: r.Fingerprint,
		}
		if hadMatch[i] {
			e.Event = measure.EventInject
			if group == measure.GroupControl {
				e.Event = measure.EventControl
			}
		}

		for _, m := range r.Matches {
			e.RecordID = m.ID
			e.Bytes += len(m.Title)
		}
		e.EstTokens = measure.EstimateTokensFromBytes(e.Bytes)

		// A ledger that cannot be written must not break the injection
		// path (ADR-0007). The loss shows up as a gap in `report`.
		_ = measure.AppendEntry(root, e)
	}
}
