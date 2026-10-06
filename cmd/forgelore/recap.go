package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/forgeprint/forgelore/internal/worklog"
)

// cmdRecap prints what a session's transcript says happened. It is the CLI
// half of work memory, and K6 is why it exists before any hook reads a
// transcript: a person can run this, see exactly what Forgelore would have
// learned, and disagree with it before anything is written.
//
// Nothing here writes a record or a candidate. The only thing it can write is
// the mark saying a transcript has been read, and only when asked.
func cmdRecap(e env, args []string) error {
	fs := newFlags(e, "recap")
	transcript := fs.String("transcript", "", "the agent transcript to read")
	previous := fs.Bool("previous", false,
		"treat --transcript as the current session and read the one before it")
	settled := fs.Duration("settled", worklog.Settled,
		"how long a transcript must go unwritten to count as finished")
	limit := fs.Int64("max-bytes", worklog.Cap, "how much of a transcript to read")
	mark := fs.Bool("mark", false, "record the transcript as read")
	all := fs.Bool("all", false, "read a transcript even if it is marked as read")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if _, err := parseInterspersed(fs, args); err != nil {
		return err
	}
	if *transcript == "" {
		return fmt.Errorf("recap needs --transcript (a hook payload names it transcript_path)")
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	marksPath := worklog.MarksPath(filepath.Join(s.Root(), "cache"))
	marks := worklog.LoadMarks(marksPath)
	now := time.Now()

	path := *transcript
	if *previous {
		found, err := worklog.Previous(path, now, *settled)
		if err != nil {
			return err
		}
		path = ""
		for _, p := range found {
			if *all || !marks.Seen(p) {
				path = p
				break
			}
		}
		if path == "" {
			if *asJSON {
				return writeJSON(e.stdout, struct {
					Session *worklog.Session `json:"session"`
				}{nil})
			}
			fmt.Fprintln(e.stdout, "no finished transcript left to read")
			return nil
		}
	} else if !*all && marks.Seen(path) {
		if !*asJSON {
			fmt.Fprintln(e.stdout, "already read (pass --all to read it again)")
			return nil
		}
	}

	session, err := worklog.ReadFile(path, *limit)
	if err != nil {
		return err
	}
	if *mark {
		marks.Add(path, now)
		if err := worklog.SaveMarks(marksPath, marks, now); err != nil {
			return err
		}
	}

	if *asJSON {
		return writeJSON(e.stdout, struct {
			Transcript string           `json:"transcript"`
			Session    *worklog.Session `json:"session"`
			Fixed      []string         `json:"fixed,omitempty"`
		}{path, session, session.Fixed()})
	}
	return printRecap(e, path, session)
}

func printRecap(e env, path string, s *worklog.Session) error {
	head := "session"
	if s.ID != "" {
		head += " " + s.ID
	}
	if s.Agent != "" {
		head += "  " + s.Agent
	}
	if !s.Start.IsZero() && !s.End.IsZero() {
		head += "  " + s.End.Sub(s.Start).Round(time.Second).String()
	}
	fmt.Fprintln(e.stdout, head)
	if s.Branch != "" {
		fmt.Fprintf(e.stdout, "  on %s\n", s.Branch)
	}

	if len(s.Edits) > 0 {
		fmt.Fprintf(e.stdout, "\n%d file(s) edited\n", len(s.Edits))
		for _, f := range s.Edits {
			fmt.Fprintf(e.stdout, "  %s\n", f)
		}
	}

	if len(s.Commands) > 0 {
		failed := 0
		for _, c := range s.Commands {
			if c.Failed {
				failed++
			}
		}
		fixed := map[string]bool{}
		for _, f := range s.Fixed() {
			fixed[f] = true
		}
		fmt.Fprintf(e.stdout, "\n%d command(s), %d failed\n", len(s.Commands), failed)
		for _, c := range s.Commands {
			state := "passed"
			switch {
			case c.Failed:
				state = "failed"
			case fixed[c.Text]:
				state = "passed, after failing"
			}
			fmt.Fprintf(e.stdout, "  %-8s %s\n", state, oneLine(c.Text))
		}
	}

	if s.Truncated || s.Unreadable > 0 {
		fmt.Fprintln(e.stdout)
		if s.Truncated {
			fmt.Fprintln(e.stdout, "the transcript was read from its tail, so the "+
				"beginning of the session is not in this")
		}
		if s.Unreadable > 0 {
			fmt.Fprintf(e.stdout, "%d line(s) could not be read\n", s.Unreadable)
		}
	}
	return nil
}

// oneLine keeps a multi-line command from turning one row into several.
func oneLine(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), "\n", " ; ")
	if len(s) > 100 {
		s = s[:97] + "..."
	}
	return s
}
