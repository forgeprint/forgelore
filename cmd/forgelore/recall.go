package main

import (
	"fmt"

	"github.com/forgeprint/forgelore/internal/fingerprint"
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
