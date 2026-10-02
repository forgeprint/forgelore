package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/forgeprint/forgelore/internal/fingerprint"
	"github.com/forgeprint/forgelore/internal/record"
	"github.com/forgeprint/forgelore/internal/redact"
)

func cmdRecord(e env, args []string) error {
	fs := newFlags(e, "record")
	typ := fs.String("type", "", "fix, dead_end, decision, command or note")
	title := fs.String("title", "", "the one line that gets injected when this record matches")
	scope := fs.String("scope", string(record.ScopeTeam), "team or local")
	tags := fs.String("tags", "", "comma-separated, lowercase kebab-case")
	related := fs.String("related", "", "comma-separated ids of related records")
	fp := fs.String("fingerprint", "", "the error fingerprint this record answers")
	errorFile := fs.String("error-file", "", "derive the fingerprint from captured output in this file; - for stdin")
	command := fs.String("command", "", "the command that produced --error-file")
	body := fs.String("body", "", "the markdown body")
	bodyFile := fs.String("body-file", "", "read the body from this file; - for stdin")
	tainted := fs.Bool("tainted", false, "the content came from outside the project (ADR-0013)")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *typ == "" {
		return fmt.Errorf("--type is required (fix, dead_end, decision, command or note)")
	}
	if strings.TrimSpace(*title) == "" {
		return fmt.Errorf("--title is required: it is the line that reaches the model")
	}
	if *body != "" && *bodyFile != "" {
		return fmt.Errorf("--body and --body-file say the same thing two ways; use one")
	}
	if *fp != "" && *errorFile != "" {
		return fmt.Errorf("--fingerprint and --error-file say the same thing two ways; use one")
	}
	if *errorFile == "-" && *bodyFile == "-" {
		return fmt.Errorf("--error-file and --body-file cannot both read stdin")
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}

	bodyText := *body
	if *bodyFile != "" {
		bodyText, err = readFileOrStdin(e, *bodyFile)
		if err != nil {
			return err
		}
	}

	sum := *fp
	if *errorFile != "" {
		sum, err = fingerprintFromOutput(e, *command, *errorFile)
		if err != nil {
			return err
		}
	}

	id, err := record.NewULID()
	if err != nil {
		return err
	}
	r := &record.Record{
		Schema:      record.CurrentSchema,
		ID:          id,
		Type:        *typ,
		Scope:       record.Scope(*scope),
		Title:       *title,
		Created:     time.Now().UTC().Truncate(time.Second),
		Source:      record.SourceUser,
		Tainted:     *tainted,
		Fingerprint: sum,
		Tags:        splitList(*tags),
		Related:     splitList(*related),
		Body:        bodyText,
	}

	found, err := s.Put(r)
	if err != nil {
		return err
	}

	// The record is indexed now rather than on the next read, so that a
	// recall immediately after a record finds it.
	if idx, err := openIndex(s); err == nil {
		idx.Close()
	}

	path, err := s.Path(r.Scope, r.ID)
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(e.stdout, struct {
			ID          string   `json:"id"`
			Path        string   `json:"path"`
			Type        string   `json:"type"`
			Scope       string   `json:"scope"`
			Fingerprint string   `json:"fingerprint,omitempty"`
			Redacted    []string `json:"redacted,omitempty"`
		}{r.ID, path, r.Type, string(r.Scope), r.Fingerprint, redactedKinds(found)})
	}

	fmt.Fprintf(e.stdout, "%s  %s\n", r.ID, relativeTo(e.wd, path))
	if r.Fingerprint != "" {
		fmt.Fprintf(e.stdout, "fingerprint %s\n", r.Fingerprint)
	}
	if len(found) > 0 {
		fmt.Fprintf(e.stdout, "redacted before writing: %s\n", redact.Summary(found))
	}
	return nil
}

// fingerprintFromOutput derives the fingerprint from real captured output.
//
// It refuses to choose when the output holds more than one error. Picking the
// first would attach the fix to whichever diagnostic the tool happened to
// print first, which is not a decision this command is entitled to make.
func fingerprintFromOutput(e env, command, path string) (string, error) {
	output, err := readFileOrStdin(e, path)
	if err != nil {
		return "", err
	}
	events := fingerprint.Scan(command, output)
	switch len(events) {
	case 0:
		return "", fmt.Errorf("no error found in that output; pass --fingerprint if you know it")
	case 1:
		return events[0].Sum, nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "that output holds %d errors; pass one of these as --fingerprint:", len(events))
	for _, ev := range events {
		fmt.Fprintf(&b, "\n  %s  %s", ev.Sum, ev.Message)
	}
	return "", fmt.Errorf("%s", b.String())
}

func readFileOrStdin(e env, path string) (string, error) {
	if path == "-" {
		data, err := io.ReadAll(e.stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func redactedKinds(found []redact.Finding) []string {
	var out []string
	for _, f := range found {
		out = append(out, f.Kind)
	}
	return out
}
