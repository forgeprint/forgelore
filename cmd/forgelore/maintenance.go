package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/forgeprint/forgelore/internal/config"
	"github.com/forgeprint/forgelore/internal/record"
	"github.com/forgeprint/forgelore/internal/store"
)

// cmdIndex holds the index subcommands. Only rebuild is here: every other
// command syncs the index on its way in, so there is nothing for a `sync`
// subcommand to do that has not already happened.
func cmdIndex(e env, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("index needs a subcommand (rebuild)")
	}
	switch args[0] {
	case "rebuild":
		return cmdIndexRebuild(e, args[1:])
	default:
		return fmt.Errorf("unknown index subcommand %q (rebuild)", args[0])
	}
}

func cmdIndexRebuild(e env, args []string) error {
	fs := newFlags(e, "index rebuild")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	idx, err := s.OpenIndex()
	if err != nil {
		return err
	}
	defer idx.Close()

	stats, err := idx.Rebuild()
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(e.stdout, struct {
			Path     string   `json:"path"`
			Scanned  int      `json:"scanned"`
			Added    int      `json:"added"`
			Removed  int      `json:"removed"`
			Millis   int64    `json:"milliseconds"`
			Problems []string `json:"problems,omitempty"`
		}{s.IndexPath(), stats.Scanned, stats.Added, stats.Removed,
			stats.Duration.Milliseconds(), problemStrings(stats.Problems)})
	}

	fmt.Fprintf(e.stdout, "rebuilt %s\n%d scanned, %d indexed, %d removed, %s\n",
		relativeTo(e.wd, s.IndexPath()), stats.Scanned, stats.Added, stats.Removed, stats.Duration.Round(1e6))
	for _, p := range stats.Problems {
		fmt.Fprintf(e.stdout, "  %s\n", p)
	}
	return nil
}

// cmdDoctor reports what is wrong without changing anything. A record that
// cannot be read is skipped rather than repaired (ADR-0007), which means
// something has to say out loud that it was skipped, or a typo silently
// removes a memory.
func cmdDoctor(e env, args []string) error {
	fs := newFlags(e, "doctor")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	records, problems, err := s.All()
	if err != nil {
		return err
	}

	cfg, configProblems := loadConfig(e, s, nil)

	idx, err := openIndex(s)
	if err != nil {
		return err
	}
	defer idx.Close()
	indexed, err := idx.Count()
	if err != nil {
		return err
	}

	var skipped int
	for _, p := range problems {
		if p.Skipped {
			skipped++
		}
	}

	settings := cfg.Effective()

	// The secret scan is only as good as what is installed, so doctor says
	// which half of it will run.
	gitleaks := "not installed; forgelore check will run its own patterns only"
	if path, err := exec.LookPath("gitleaks"); err == nil {
		gitleaks = path
	}

	if *asJSON {
		type setting struct {
			Key    string `json:"key"`
			Value  string `json:"value"`
			Source string `json:"source"`
			From   string `json:"from"`
		}
		rows := make([]setting, 0, len(settings))
		for _, s := range settings {
			rows = append(rows, setting{s.Key, s.Value, s.Source, s.From})
		}
		return writeJSON(e.stdout, struct {
			Root           string    `json:"root"`
			Records        int       `json:"records"`
			Indexed        int       `json:"indexed"`
			Skipped        int       `json:"skipped"`
			Problems       []string  `json:"problems,omitempty"`
			ConfigProblems []string  `json:"config_problems,omitempty"`
			Gitleaks       string    `json:"gitleaks"`
			Config         []setting `json:"config"`
		}{s.Root(), len(records), indexed, skipped,
			problemStrings(problems), configProblemStrings(configProblems), gitleaks, rows})
	}

	fmt.Fprintf(e.stdout, "store    %s\n", relativeTo(e.wd, s.Root()))
	fmt.Fprintf(e.stdout, "records  %d readable, %d skipped\n", len(records), skipped)
	fmt.Fprintf(e.stdout, "index    %d records\n", indexed)
	fmt.Fprintf(e.stdout, "gitleaks %s\n", gitleaks)

	// ADR-0018: five sources means no file shows the effective value, so
	// this has to say where each one actually came from.
	fmt.Fprintln(e.stdout, "\nsettings")
	for _, s := range settings {
		from := s.From
		if filepath.IsAbs(from) {
			from = relativeTo(e.wd, from)
		}
		fmt.Fprintf(e.stdout, "  %-30s %-10s  %s\n", s.Key, s.Value, from)
	}

	if len(problems) == 0 && len(configProblems) == 0 {
		fmt.Fprintln(e.stdout, "\nnothing to report")
		return nil
	}
	fmt.Fprintln(e.stdout)
	for _, p := range problems {
		fmt.Fprintf(e.stdout, "%s\n", p)
	}
	for _, p := range configProblems {
		fmt.Fprintf(e.stdout, "%s\n", p)
	}
	return nil
}

func configProblemStrings(problems []config.Problem) []string {
	var out []string
	for _, p := range problems {
		out = append(out, p.String())
	}
	return out
}

// cmdStats summarises what the store holds.
func cmdStats(e env, args []string) error {
	fs := newFlags(e, "stats")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	records, _, err := s.All()
	if err != nil {
		return err
	}

	byType := map[string]int{}
	byScope := map[string]int{}
	var superseded, fingerprinted int
	for _, r := range records {
		byType[r.Type]++
		byScope[string(r.Scope)]++
		if r.SupersededBy != "" {
			superseded++
		}
		if r.Fingerprint != "" {
			fingerprinted++
		}
	}

	var indexBytes int64
	if info, err := os.Stat(s.IndexPath()); err == nil {
		indexBytes = info.Size()
	}

	if *asJSON {
		return writeJSON(e.stdout, struct {
			Root          string         `json:"root"`
			Total         int            `json:"total"`
			ByType        map[string]int `json:"by_type"`
			ByScope       map[string]int `json:"by_scope"`
			Superseded    int            `json:"superseded"`
			Fingerprinted int            `json:"fingerprinted"`
			IndexBytes    int64          `json:"index_bytes"`
		}{s.Root(), len(records), byType, byScope, superseded, fingerprinted, indexBytes})
	}

	fmt.Fprintf(e.stdout, "%d record(s) in %s\n\n", len(records), relativeTo(e.wd, s.Root()))
	for _, t := range []string{record.TypeFix, record.TypeDeadEnd, record.TypeDecision, record.TypeCommand, record.TypeNote} {
		if byType[t] > 0 {
			fmt.Fprintf(e.stdout, "  %-10s %d\n", t, byType[t])
		}
		delete(byType, t)
	}
	for _, t := range sortedMapKeys(byType) {
		fmt.Fprintf(e.stdout, "  %-10s %d  (unrecognised type, treated as a note)\n", t, byType[t])
	}
	fmt.Fprintf(e.stdout, "\n  team       %d\n  local      %d\n", byScope["team"], byScope["local"])
	fmt.Fprintf(e.stdout, "\n  with a fingerprint  %d\n  superseded          %d\n", fingerprinted, superseded)
	fmt.Fprintf(e.stdout, "  index               %d bytes\n", indexBytes)
	return nil
}

func problemStrings(problems []store.Problem) []string {
	var out []string
	for _, p := range problems {
		out = append(out, p.String())
	}
	return out
}

func sortedMapKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
