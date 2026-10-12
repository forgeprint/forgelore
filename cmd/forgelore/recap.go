package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/forgeprint/forgelore/internal/candidate"
	"github.com/forgeprint/forgelore/internal/config"
	"github.com/forgeprint/forgelore/internal/measure"
	"github.com/forgeprint/forgelore/internal/redact"
	"github.com/forgeprint/forgelore/internal/store"
	"github.com/forgeprint/forgelore/internal/worklog"
)

// cmdRecap prints what a session's transcript says happened. It is the CLI
// half of work memory, and K6 is why it exists before any hook reads a
// transcript: a person can run this, see exactly what Forgelore would have
// learned, and disagree with it before anything is written.
//
// Nothing here becomes a record. --propose puts what it found in the review
// queue, which is where a person turns it into memory or does not (ADR-0026);
// --mark says a transcript has been read. Without either flag it writes
// nothing at all.
func cmdRecap(e env, args []string) error {
	fs := newFlags(e, "recap")
	transcript := fs.String("transcript", "", "the agent transcript to read")
	previous := fs.Bool("previous", false,
		"treat --transcript as the current session and read the one before it")
	settled := fs.Duration("settled", worklog.Settled,
		"how long a transcript must go unwritten to count as finished")
	limit := fs.Int64("max-bytes", worklog.Cap, "how much of a transcript to read")
	propose := fs.Bool("propose", false, "write what it found as a candidate for review")
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
	if *propose {
		if err := proposeWork(s, session); err != nil {
			return err
		}
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

// harvest is the automatic half of work memory: at the start of a session,
// read the transcript of the one before it and put what it found in the review
// queue.
//
// Every way this can fail is a way for it to do nothing. A format nothing here
// reads, a directory that is gone, a transcript still being written, one
// already harvested, a deadline already passed — each returns silently, because
// a hook that cannot be silent is a hook that can break an agent (K7).
func harvest(s *store.Store, cfg *config.Config, format, transcript string,
	deadline time.Time, now time.Time) {
	if format != formatClaudeCode || transcript == "" || !now.Before(deadline) {
		return
	}
	marksPath := worklog.MarksPath(filepath.Join(s.Root(), "cache"))
	marks := worklog.LoadMarks(marksPath)

	found, err := worklog.Previous(transcript, now, worklog.Settled)
	if err != nil {
		return
	}
	for _, path := range found {
		if marks.Seen(path) {
			continue
		}
		if !time.Now().Before(deadline) {
			return
		}
		session, err := worklog.ReadFile(path, 0)
		if err != nil {
			// Marked anyway: a transcript that cannot be read now will not
			// read better at the next session start, and retrying it every
			// time would spend the deadline on it forever.
			marks.Add(path, now)
			continue
		}
		if err := proposeWork(s, session); err != nil {
			return
		}
		recordSession(s, cfg, session)
		if cfg.Bool("distill.enabled") {
			startDistill(s, session.ID)
		}
		marks.Add(path, now)
		// One per session start. The next one is still there next time, and
		// a backlog of ten transcripts must not become ten parses inside
		// one hook's deadline.
		break
	}
	_ = worklog.SaveMarks(marksPath, marks, now)
}

// recordSession puts a finished session's shape in the ledger: how much of it
// was the agent looking around, which is the outcome work memory is measured
// by (ADR-0019).
//
// The arm is computed from the session's own id, not from the arm of the
// session doing the harvesting. Assignment is deterministic, so the session
// that was measured is the session whose arm this is — and it has to be, since
// a control session is the baseline the treatment is compared against.
func recordSession(s *store.Store, cfg *config.Config, session *worklog.Session) {
	if !cfg.Bool("measure.ledger") || session.ID == "" || session.Calls == 0 {
		return
	}
	_ = measure.AppendEntry(s.Root(), measure.Entry{
		Time:    session.End,
		Session: session.ID,
		Event:   measure.EventSession,
		Group: measure.Assign(cfg.String("measure.ab.salt"), session.ID,
			int(cfg.Int("measure.ab.control_percent"))),
		Discovery: session.Discovery,
		Calls:     session.Calls,
	})
}

// formatClaudeCode is the only transcript format with a parser. A mapping that
// names another one gets nothing until there is code to read it.
const formatClaudeCode = "claude-code-jsonl"

// proposeWork puts a session's account in the review queue.
//
// The body is the agent's own closing words first and the facts after them,
// because the words say why and the facts only say what. No title: the one
// line that would be injected is the person's to write, exactly as it is for
// a fix (ADR-0026). Redaction runs before anything is written — a transcript
// carries whole prompts, which is the hardest thing this project asks of it
// (ADR-0013).
func proposeWork(s *store.Store, session *worklog.Session) error {
	var b strings.Builder
	if session.Closing != "" {
		b.WriteString(session.Closing)
		b.WriteString("\n")
	}
	if len(session.Edits) > 0 {
		b.WriteString("\nFiles touched:\n")
		for _, f := range session.Edits {
			b.WriteString("- " + f + "\n")
		}
	}
	if fixed := session.Fixed(); len(fixed) > 0 {
		b.WriteString("\nStarted working again:\n")
		for _, c := range fixed {
			b.WriteString("- `" + oneLine(c) + "`\n")
		}
	}
	body, _ := redact.Text(strings.TrimSpace(b.String()))
	if body == "" {
		return nil
	}

	return candidate.Upsert(s.Root(), candidate.Candidate{
		Time:    time.Now().UTC().Truncate(time.Second),
		Session: session.ID,
		Kind:    candidate.KindWork,
		Body:    body,
	})
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

	if s.Closing != "" {
		fmt.Fprintf(e.stdout, "\nwhat the agent said last:\n%s\n", indent(firstLines(s.Closing, 6), "  "))
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

// firstLines keeps a closing message from filling a terminal. The whole of it
// is what a candidate carries; this is only the look at it.
func firstLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) <= n {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[:n], "\n") + "\n..."
}

// oneLine keeps a multi-line command from turning one row into several.
func oneLine(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), "\n", " ; ")
	if len(s) > 100 {
		s = s[:97] + "..."
	}
	return s
}
