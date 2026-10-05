package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/forgeprint/forgelore/internal/record"
	"github.com/forgeprint/forgelore/internal/redact"
	"github.com/forgeprint/forgelore/internal/store"
)

// cmdPromote moves a local record into the team scope, where it will be
// committed and read by everybody.
//
// Automatically captured memories always start local (ADR-0010), and this is
// the step with a human in it. A tainted record — one derived from content
// outside the project (ADR-0013) — needs the intent spelled out, because
// promoting it publishes something nobody in the team wrote.
func cmdPromote(e env, args []string) error {
	fs := newFlags(e, "promote")
	forceTainted := fs.Bool("force-tainted", false, "promote even though the record came from outside the project")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	words, err := parseInterspersed(fs, args)
	if err != nil {
		return err
	}
	if len(words) != 1 {
		return fmt.Errorf("promote needs exactly one record id")
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	r, err := s.Get(words[0])
	if err != nil {
		return err
	}
	if r.Scope == record.ScopeTeam {
		return fmt.Errorf("%s is already a team record", r.ID)
	}
	if r.Tainted && !*forceTainted {
		return fmt.Errorf("%s is tainted: it came from outside the project, and promoting it shares it with the team.\nRead it first with: forgelore show %s\nThen, if it should be shared: forgelore promote %s --force-tainted", r.ID, r.ID, r.ID)
	}

	if err := s.Promote(r.ID); err != nil {
		return err
	}
	if idx, err := openIndex(s); err == nil {
		idx.Close()
	}
	path, err := s.Path(record.ScopeTeam, r.ID)
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(e.stdout, struct {
			ID      string `json:"id"`
			Path    string `json:"path"`
			Tainted bool   `json:"tainted"`
		}{r.ID, path, r.Tainted})
	}
	fmt.Fprintf(e.stdout, "%s  %s\n", r.ID, relativeTo(e.wd, path))
	fmt.Fprintln(e.stdout, "Commit it to share it.")
	return nil
}

// leak is one thing found in a record that should not be shared.
type leak struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
	From string `json:"found_by"`
}

// cmdCheck is what stands between a secret and a commit.
//
// Redaction already runs when a record is written, so a record Forgelore
// created is clean. This reads the files as they are on disk instead,
// because a record can also arrive by hand, by merge, or by an editor, and
// none of those went through the writer.
//
// Only team records are read. Local ones never leave the machine, and
// scanning them would mean reporting a person's own notes back at them.
func cmdCheck(e env, args []string) error {
	fs := newFlags(e, "check")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	teamDir, err := s.Dir(record.ScopeTeam)
	if err != nil {
		return err
	}

	found, scanned, err := scanTeamRecords(teamDir)
	if err != nil {
		return err
	}

	// gitleaks knows far more patterns than this project does, so it runs
	// too when it is installed. It is not required: a check that cannot run
	// without a tool the user has never heard of is a check that gets
	// skipped.
	gitleaksRan := false
	if path, lookErr := exec.LookPath("gitleaks"); lookErr == nil {
		gitleaksRan = true
		if leaked, err := runGitleaks(path, teamDir); err != nil {
			return err
		} else if leaked {
			found = append(found, leak{Path: relativeTo(e.wd, teamDir), Kind: "see gitleaks output", From: "gitleaks"})
		}
	}

	if *asJSON {
		if err := writeJSON(e.stdout, struct {
			Scanned     int    `json:"scanned"`
			GitleaksRan bool   `json:"gitleaks_ran"`
			Leaks       []leak `json:"leaks"`
		}{scanned, gitleaksRan, found}); err != nil {
			return err
		}
	} else {
		fmt.Fprintf(e.stdout, "%d team record(s) scanned\n", scanned)
		if !gitleaksRan {
			fmt.Fprintln(e.stdout, "gitleaks is not installed, so only Forgelore's own patterns ran")
		}
		for _, l := range found {
			fmt.Fprintf(e.stdout, "%s: %s (%s)\n", l.Path, l.Kind, l.From)
		}
		if len(found) == 0 {
			fmt.Fprintln(e.stdout, "nothing to hide")
		}
	}

	if len(found) > 0 {
		return fmt.Errorf("%d team record(s) carry something that should not be shared", len(found))
	}
	return nil
}

// scanTeamRecords reads every committed record as bytes, so that a secret in
// a field Forgelore does not know about is still found.
func scanTeamRecords(dir string) ([]leak, int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	var found []leak
	scanned := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, scanned, err
		}
		scanned++
		if _, findings := redact.Text(string(data)); len(findings) > 0 {
			for _, f := range findings {
				found = append(found, leak{Path: entry.Name(), Kind: f.Kind, From: "forgelore"})
			}
		}
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Path < found[j].Path })
	return found, scanned, nil
}

// runGitleaks reports whether gitleaks found anything. Its own output goes
// to the user; only the verdict comes back here.
func runGitleaks(bin, dir string) (bool, error) {
	cmd := exec.Command(bin, "dir", dir, "--redact", "--no-banner")
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	err := cmd.Run()
	if err == nil {
		return false, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return true, nil
	}
	return false, fmt.Errorf("running gitleaks: %w", err)
}

// duplicate is a group of records answering one fingerprint.
type duplicate struct {
	Fingerprint string    `json:"fingerprint"`
	Records     []hitJSON `json:"records"`
}

// cmdDedupe finds records that answer the same error and, on request, marks
// one as superseded by another.
//
// Two fixes for one fingerprint are both injected, which costs tokens twice
// and tells the model two things where one would do. The record format has
// carried `superseded_by` since schema 1 and nothing has ever written it;
// this is what writes it.
func cmdDedupe(e env, args []string) error {
	fs := newFlags(e, "dedupe")
	supersede := fs.String("supersede", "", "the id of the record to retire")
	by := fs.String("by", "", "the id of the record that replaces it")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	if *supersede != "" || *by != "" {
		return supersedeRecord(e, s, *supersede, *by)
	}

	groups, err := duplicateGroups(s)
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(e.stdout, struct {
			Duplicates []duplicate `json:"duplicates"`
		}{groups})
	}
	if len(groups) == 0 {
		fmt.Fprintln(e.stdout, "no fingerprint has more than one record")
		return nil
	}
	for _, g := range groups {
		fmt.Fprintf(e.stdout, "\n%s\n", g.Fingerprint)
		for _, r := range g.Records {
			fmt.Fprintf(e.stdout, "  %s  %-9s %s\n", r.ID, r.Type, r.Title)
		}
	}
	fmt.Fprintf(e.stdout, "\n%d fingerprint(s) with more than one record. Retire one with:\n", len(groups))
	fmt.Fprintln(e.stdout, "  forgelore dedupe --supersede <old id> --by <new id>")
	return nil
}

// duplicateGroups returns the fingerprints answered by more than one live
// record. A superseded record is already retired and does not count.
func duplicateGroups(s *store.Store) ([]duplicate, error) {
	records, _, err := s.All()
	if err != nil {
		return nil, err
	}

	byFingerprint := map[string][]hitJSON{}
	for _, r := range records {
		if r.Fingerprint == "" || r.SupersededBy != "" {
			continue
		}
		byFingerprint[r.Fingerprint] = append(byFingerprint[r.Fingerprint], hitJSON{
			ID: r.ID, Type: r.Type, Title: r.Title, Scope: string(r.Scope),
		})
	}

	var out []duplicate
	for fingerprint, records := range byFingerprint {
		if len(records) < 2 {
			continue
		}
		sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
		out = append(out, duplicate{Fingerprint: fingerprint, Records: records})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Fingerprint < out[j].Fingerprint })
	return out, nil
}

// supersedeRecord retires one record in favour of another. The retired
// record is kept: it stays searchable and marked, because a dead end that
// was tried is worth knowing about even once something better exists
// (ADR-0016).
func supersedeRecord(e env, s *store.Store, oldID, newID string) error {
	if oldID == "" || newID == "" {
		return fmt.Errorf("--supersede and --by are both needed")
	}
	old, err := s.Get(oldID)
	if err != nil {
		return err
	}
	replacement, err := s.Get(newID)
	if err != nil {
		return err
	}
	if old.ID == replacement.ID {
		return fmt.Errorf("a record cannot supersede itself")
	}
	if old.SupersededBy != "" {
		return fmt.Errorf("%s is already superseded by %s", old.ID, old.SupersededBy)
	}

	old.SupersededBy = replacement.ID
	if _, err := s.Put(old); err != nil {
		return err
	}
	if idx, err := openIndex(s); err == nil {
		idx.Close()
	}
	fmt.Fprintf(e.stdout, "%s is now superseded by %s\nIt stays searchable and will not be injected again.\n", old.ID, replacement.ID)
	return nil
}

// preCommitHook is what a project adds to stop a secret reaching a commit.
const preCommitHook = `#!/bin/sh
# Added by forgelore init --with-git-hook
exec forgelore check
`

// hookLine is the same thing as one line, for a project that already has a
// pre-commit hook of its own.
const hookLine = "forgelore check || exit 1"

// installPreCommitHook writes the hook, refusing to overwrite one that is
// already there. A tool that silently replaces a project's own hook is a
// tool that loses somebody a morning.
func installPreCommitHook(wd string) (string, error) {
	gitDir := filepath.Join(wd, ".git")
	if info, err := os.Stat(gitDir); err != nil || !info.IsDir() {
		return "", fmt.Errorf("%s is not a git repository, so there is nowhere to put a hook", wd)
	}
	hooks := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(hooks, "pre-commit")
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("%s already exists; add this line to it yourself:\n  %s", path, hookLine)
	}
	if err := os.WriteFile(path, []byte(preCommitHook), 0o755); err != nil {
		return "", err
	}
	return path, nil
}
