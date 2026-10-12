package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/forgeprint/forgelore/internal/agent"
	"github.com/forgeprint/forgelore/internal/candidate"
	"github.com/forgeprint/forgelore/internal/config"
	"github.com/forgeprint/forgelore/internal/fingerprint"
	"github.com/forgeprint/forgelore/internal/measure"
	"github.com/forgeprint/forgelore/internal/record"
	"github.com/forgeprint/forgelore/internal/store"
)

// cmdHook is the agent-facing entry point.
//
// It never fails. K7 is not a preference here: a hook that exits non-zero
// puts an error in the user's session, and one that exits 2 blocks their
// agent outright. Every path below ends in a successful exit and, at worst,
// an empty response. The costs of that are real — a broken store is silent
// rather than loud — which is why `doctor` exists and why the ledger records
// what happened.
func cmdHook(e env, args []string) error {
	fs := newFlags(e, "hook")
	adapter := fs.String("adapter", "claude-code", "which agent is calling: "+strings.Join(agent.BuiltIn(), ", "))
	mappingFile := fs.String("mapping", "", "use this mapping file instead of the built-in one")
	event := fs.String("event", "", "the agent's name for this event, for agents whose payload does not carry it")
	dir := fs.String("dir", "", "work on the project containing this directory")
	if err := fs.Parse(args); err != nil {
		// Even a bad flag must not break the session.
		return nil
	}

	deadline := time.Now()
	response, err := runHook(e, *adapter, *mappingFile, *event, *dir, &deadline)
	if err != nil {
		// The reason goes to stderr, where Claude Code shows it only if the
		// exit code says something went wrong. It does not.
		fmt.Fprintln(e.stderr, "forgelore hook:", err)
		return nil
	}
	if response == "" {
		return nil
	}
	_, _ = io.WriteString(e.stdout, response)
	return nil
}

// runHook does the work, with every failure turned into a quiet nothing by
// its caller.
func runHook(e env, adapter, mappingFile, callerEvent, dir string, started *time.Time) (resp string, err error) {
	defer func() {
		// A panic in a hook is a crashed process and a visible error in the
		// user's session. Recovering turns the worst outcome into silence.
		if r := recover(); r != nil {
			resp, err = "", fmt.Errorf("recovered from a panic: %v", r)
		}
	}()

	payload, err := io.ReadAll(e.stdin)
	if err != nil {
		return "", err
	}

	mapping, err := loadMapping(adapter, mappingFile)
	if err != nil {
		return "", err
	}

	event, ok, err := mapping.Translate(payload, callerEvent, time.Now().UTC().Truncate(time.Second))
	if err != nil {
		return "", err
	}
	if !ok {
		// An event this mapping does not cover. Most of them are.
		return "", nil
	}

	s, err := openStore(e, dir)
	if err != nil {
		return "", err
	}
	cfg, _ := loadConfig(e, s, nil)

	deadline := started.Add(time.Duration(cfg.Int("hook.deadline_ms")) * time.Millisecond)

	var text string
	switch event.Kind {
	case agent.SessionStarted:
		// Reading the previous session's transcript happens here, not at
		// its own end: the file is written asynchronously and is
		// incomplete while the session that produced it is closing
		// (ADR-0026). The index is built first, because it is what this
		// event owes the agent and it is cheap.
		text, err = sessionIndex(s, cfg, deadline)
		if cfg.Bool("work.harvest") {
			harvest(s, event.TranscriptFormat, event.Transcript, deadline, *started)
		}
	case agent.CommandFailed:
		text, err = onCommandFailed(s, cfg, event, deadline, *started)
	case agent.CommandSucceeded:
		err = onCommandSucceeded(s, event)
	case agent.SessionEnded:
		text, err = onSessionEnded(s, event)
	}
	if err != nil || text == "" {
		return "", err
	}

	// A notice goes to the person, and on an event that cannot reach the
	// model it is the only channel there is. Claude Code's SessionEnd is
	// such an event: text put in the context path there used to reach
	// nobody at all.
	var reply []byte
	if event.Reply == "notice" {
		reply, err = mapping.Notice(text)
	} else {
		reply, err = mapping.Context(event.AgentEvent, text)
	}
	if err != nil {
		return "", err
	}
	return string(reply), nil
}

func loadMapping(adapter, file string) (*agent.Mapping, error) {
	if file == "" {
		return agent.Built(adapter)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return agent.ParseMapping(data)
}

// sessionIndex is the budgeted list handed over once per session: the titles
// of commands and decisions, and of the last few sessions' work (ADR-0016).
// Fixes and dead ends are deliberately absent — they cost nothing until an
// error matches them.
//
// Work has a budget of its own rather than a share of this one. It arrives
// once per session where a decision arrives once per decision, so one budget
// would mean a busy fortnight quietly pushing every decision out of the index.
func sessionIndex(s *store.Store, cfg *config.Config, deadline time.Time) (string, error) {
	records, _, err := s.All()
	if err != nil {
		return "", err
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID > records[j].ID })

	// Newest first in both halves, which for work is the point: what the
	// last session learned is worth more than what the tenth-last did.
	known := titlesWithin(records, deadline,
		int(cfg.Int("inject.budget_tokens")), (*record.Record).InSessionIndex)
	work := titlesWithin(records, deadline,
		int(cfg.Int("inject.work_budget_tokens")), (*record.Record).InWorkIndex)
	if len(known) == 0 && len(work) == 0 {
		return "", nil
	}

	out := "Project memory (forgelore). Ask for detail with: forgelore show <id>"
	if len(known) > 0 {
		out += "\n" + strings.Join(known, "\n")
	}
	if len(work) > 0 {
		out += "\nWhat recent sessions learned:\n" + strings.Join(work, "\n")
	}
	return out, nil
}

// titlesWithin collects the titles a predicate admits, up to a budget.
func titlesWithin(records []*record.Record, deadline time.Time, budget int,
	wanted func(*record.Record) bool) []string {
	var lines []string
	used := 0
	for _, r := range records {
		if time.Now().After(deadline) {
			break
		}
		if !wanted(r) {
			continue
		}
		line := "- " + r.Title + " (" + r.ID + ")"
		cost := measure.EstimateTokens(line)
		if used+cost > budget {
			break
		}
		used += cost
		lines = append(lines, line)
	}
	return lines
}

// onCommandFailed is the injection path.
func onCommandFailed(s *store.Store, cfg *config.Config, e agent.Event, deadline, startedAt time.Time) (string, error) {
	events := fingerprint.Scan(e.Command, e.Output)
	if len(events) == 0 {
		return "", nil
	}

	group := measure.Assign(cfg.String("measure.ab.salt"), e.Session,
		int(cfg.Int("measure.ab.control_percent")))

	state := agent.LoadState(agent.StatePath(e, filepath.Join(s.Root(), "cache")))

	idx, err := openIndex(s)
	if err != nil {
		return "", err
	}
	defer idx.Close()

	var lines []string
	for i, ev := range events {
		// The budget limits how many lookups are made, not whether any
		// is. Opening the index is the expensive part and it has already
		// happened by the time this loop starts; breaking here would
		// throw away work already paid for and, on a machine slow enough
		// to blow the budget before the first iteration, record nothing
		// at all — so `report` would show fewer errors than really
		// happened and say nothing about the difference.
		//
		// Found by a CI runner slower than any laptop here: the test for
		// this wrote no ledger line, and a one-millisecond deadline
		// reproduces it exactly.
		if i > 0 && time.Now().After(deadline) {
			break
		}

		hits, err := idx.ByFingerprint(ev.Sum)
		if err != nil {
			continue
		}

		// An inferred failure that matches nothing leaves no trace. The
		// agent never said this command failed; the mapping guessed it
		// from the shape of the output, and output that merely prints an
		// error looks the same as output that is one. Recording the guess
		// would inflate "nothing known" with commands that never failed
		// and fill the session state with them — measured over one
		// session, every inferred failure was of exactly that kind.
		//
		// A match is different: something was known about it, and that is
		// worth both the injection and the line in the ledger.
		if e.Inferred && len(hits) == 0 {
			continue
		}
		state.Failed[ev.Sum] = e.Command

		entry := measure.Entry{
			Time: e.Time, Session: e.Session, Group: group,
			Event: measure.EventMiss, Fingerprint: ev.Sum,
		}
		if len(hits) > 0 {
			entry.Event = measure.EventInject
			if group == measure.GroupControl {
				entry.Event = measure.EventControl
			}
		}
		for _, h := range hits {
			if group == measure.GroupControl {
				break
			}
			line := "- " + h.Type + ": " + h.Title + " (" + h.ID + ")"
			lines = append(lines, line)
			entry.RecordID = h.ID
			entry.Bytes += len(line)
		}
		entry.EstTokens = measure.EstimateTokensFromBytes(entry.Bytes)
		entry.HookMS = int(time.Since(startedAt) / time.Millisecond)
		if cfg.Bool("measure.ledger") {
			_ = measure.AppendEntry(s.Root(), entry)
		}
	}

	_ = agent.SaveState(agent.StatePath(e, filepath.Join(s.Root(), "cache")), state)
	if len(lines) == 0 {
		return "", nil
	}
	return "forgelore has seen this error before:\n" + strings.Join(lines, "\n"), nil
}

// onCommandSucceeded proposes a fix when a command that failed earlier in
// this session now works.
//
// The proposal is never a record. It goes to a candidates file in the local
// scope, unapproved, because a command that starts working is not evidence
// that anybody understood why (ADR-0013 for the tainted case, and plain
// caution otherwise). `forgelore review` is where a human turns one into a
// memory.
func onCommandSucceeded(s *store.Store, e agent.Event) error {
	path := agent.StatePath(e, filepath.Join(s.Root(), "cache"))
	state := agent.LoadState(path)
	if len(state.Failed) == 0 {
		return nil
	}

	var proposed []string
	for sum, command := range state.Failed {
		if command != e.Command {
			continue
		}
		proposed = append(proposed, sum)
	}
	if len(proposed) == 0 {
		return nil
	}
	sort.Strings(proposed)

	for _, sum := range proposed {
		if err := candidate.Append(s.Root(), candidate.Candidate{
			Time:        e.Time,
			Session:     e.Session,
			Fingerprint: sum,
			Command:     e.Command,
			Tainted:     e.Untrusted,
		}); err != nil {
			return err
		}
		delete(state.Failed, sum)
	}
	return agent.SaveState(path, state)
}

// onSessionEnded reminds the user what is waiting for them. Claude Code gives
// every SessionEnd hook a 1.5 second budget between them, so this reads one
// small file and prints one line.
func onSessionEnded(s *store.Store, e agent.Event) (string, error) {
	pending, err := candidate.Read(s.Root())
	if err != nil || len(pending) == 0 {
		return "", err
	}
	return fmt.Sprintf(
		"forgelore: %d candidate(s) are waiting. Review with: forgelore review",
		len(pending)), nil
}

// firstLine is how a body with no title introduces itself in a list.
func firstLine(s string) string {
	line := strings.TrimSpace(s)
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	if len(line) > 72 {
		line = line[:69] + "..."
	}
	if line == "" {
		return "(nothing said)"
	}
	return line
}

// cmdReview is where a candidate becomes a memory, or stops being one.
// Nothing here happens without the title a person writes: the whole point of
// a candidate is that Forgelore noticed something and does not know what it
// means.
func cmdReview(e env, args []string) error {
	fs := newFlags(e, "review")
	accept := fs.String("accept", "", "the key of the candidate to record")
	title := fs.String("title", "", "the one line that will be injected")
	drop := fs.String("drop", "", "the key of a candidate to discard")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	pending, err := candidate.Read(s.Root())
	if err != nil {
		return err
	}

	switch {
	case *accept != "" && *drop != "":
		return fmt.Errorf("--accept and --drop ask for opposite things")
	case *accept != "":
		return acceptCandidate(e, s, pending, *accept, *title)
	case *drop != "":
		dropped, kept, err := candidate.Take(pending, *drop)
		if err != nil {
			return err
		}
		if err := candidate.Write(s.Root(), kept); err != nil {
			return err
		}
		// Say so. Accepting prints the new record's id, and dropping used
		// to print nothing at all, which left running review again as the
		// only way to find out whether it had worked.
		fmt.Fprintf(e.stdout, "%s  dropped, %d still waiting\n", dropped.Key(), len(kept))
		return nil
	}

	if *asJSON {
		return writeJSON(e.stdout, struct {
			Candidates []candidate.Candidate `json:"candidates"`
		}{pending})
	}
	if len(pending) == 0 {
		fmt.Fprintln(e.stdout, "nothing waiting")
		return nil
	}
	for _, c := range pending {
		what := c.Title
		if what == "" {
			what = c.Command
		}
		if c.Kind == candidate.KindWork {
			what = firstLine(c.Body)
		}
		fmt.Fprintf(e.stdout, "%s  %s\n", c.Key(), what)
	}
	fmt.Fprintf(e.stdout, "\n%d waiting. Record one with:\n  forgelore review --accept <key> --title \"what it was\"\n", len(pending))
	return nil
}

func acceptCandidate(e env, s *store.Store, pending []candidate.Candidate, sum, title string) error {
	chosen, kept, err := candidate.Take(pending, sum)
	if err != nil {
		return err
	}

	// A hook's candidate has no title, because a hook saw an error stop and
	// not a reason. One proposed through MCP usually does. Either way the
	// person reviewing can override it.
	if strings.TrimSpace(title) == "" {
		title = chosen.Title
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("--accept needs --title: nothing proposed one, and a title is what gets injected")
	}

	body := chosen.Body
	if body == "" && chosen.Command != "" {
		body = "Noticed when `" + chosen.Command + "` started working again."
	}

	// Work memory has no fingerprint to answer. Its title reaches the
	// session-start index under the work budget, and nothing else about it
	// is injected (ADR-0016).
	kind, fingerprint := record.TypeFix, chosen.Fingerprint
	if chosen.Kind == candidate.KindWork {
		kind, fingerprint = record.TypeWork, ""
	}

	id, err := record.NewULID()
	if err != nil {
		return err
	}
	r := &record.Record{
		Schema:      record.CurrentSchema,
		ID:          id,
		Type:        kind,
		Scope:       record.ScopeLocal,
		Title:       title,
		Created:     time.Now().UTC().Truncate(time.Second),
		Source:      record.SourceHook,
		Tainted:     chosen.Tainted,
		Fingerprint: fingerprint,
		Body:        body,
	}
	if _, err := s.Put(r); err != nil {
		return err
	}
	if idx, err := openIndex(s); err == nil {
		idx.Close()
	}

	if err := candidate.Write(s.Root(), kept); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "%s  recorded in the local scope\n", r.ID)
	return nil
}
