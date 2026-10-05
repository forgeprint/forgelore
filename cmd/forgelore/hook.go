package main

import (
	"encoding/json"
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
	adapter := fs.String("adapter", "claude-code", "which agent is calling")
	mappingFile := fs.String("mapping", "", "use this mapping file instead of the built-in one")
	dir := fs.String("dir", "", "work on the project containing this directory")
	if err := fs.Parse(args); err != nil {
		// Even a bad flag must not break the session.
		return nil
	}

	deadline := time.Now()
	response, err := runHook(e, *adapter, *mappingFile, *dir, &deadline)
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
func runHook(e env, adapter, mappingFile, dir string, started *time.Time) (resp string, err error) {
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

	event, ok, err := mapping.Translate(payload, time.Now().UTC().Truncate(time.Second))
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
	switch event.Kind {
	case agent.SessionStarted:
		return sessionIndex(s, cfg, deadline)
	case agent.CommandFailed:
		return onCommandFailed(s, cfg, event, deadline, *started)
	case agent.CommandSucceeded:
		return "", onCommandSucceeded(s, event)
	case agent.SessionEnded:
		return onSessionEnded(s, event)
	}
	return "", nil
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

// claudeHookResponse is the shape Claude Code reads from a hook's stdout.
// Verified against https://code.claude.com/docs/en/hooks on 2026-10-05.
type claudeHookResponse struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

func contextResponse(eventName, text string) (string, error) {
	var r claudeHookResponse
	r.HookSpecificOutput.HookEventName = eventName
	r.HookSpecificOutput.AdditionalContext = text
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

// sessionIndex is the budgeted list handed over once per session: the titles
// of commands and decisions, and nothing else (ADR-0016). Fixes and dead ends
// are deliberately absent — they cost nothing until an error matches them.
func sessionIndex(s *store.Store, cfg *config.Config, deadline time.Time) (string, error) {
	records, _, err := s.All()
	if err != nil {
		return "", err
	}

	budget := int(cfg.Int("inject.budget_tokens"))
	var lines []string
	used := 0
	sort.Slice(records, func(i, j int) bool { return records[i].ID > records[j].ID })
	for _, r := range records {
		if time.Now().After(deadline) {
			break
		}
		if !r.InSessionIndex() {
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
	if len(lines) == 0 {
		return "", nil
	}

	text := "Project memory (forgelore). Ask for detail with: forgelore show <id>\n" +
		strings.Join(lines, "\n")
	return contextResponse("SessionStart", text)
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
	for _, ev := range events {
		state.Failed[ev.Sum] = e.Command
		if time.Now().After(deadline) {
			break
		}

		hits, err := idx.ByFingerprint(ev.Sum)
		if err != nil {
			continue
		}

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
	return contextResponse("PostToolUseFailure",
		"forgelore has seen this error before:\n"+strings.Join(lines, "\n"))
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
	return contextResponse("SessionEnd",
		fmt.Sprintf("forgelore: %d fix candidate(s) from this session are waiting. Review with: forgelore review", len(pending)))
}

// cmdReview is where a candidate becomes a memory, or stops being one.
// Nothing here happens without the title a person writes: the whole point of
// a candidate is that Forgelore noticed something and does not know what it
// means.
func cmdReview(e env, args []string) error {
	fs := newFlags(e, "review")
	accept := fs.String("accept", "", "the fingerprint of the candidate to record")
	title := fs.String("title", "", "the one line that will be injected")
	drop := fs.String("drop", "", "the fingerprint of a candidate to discard")
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
		_, kept, err := candidate.Take(pending, *drop)
		if err != nil {
			return err
		}
		return candidate.Write(s.Root(), kept)
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
		fmt.Fprintf(e.stdout, "%s  %s\n", c.Fingerprint, what)
	}
	fmt.Fprintf(e.stdout, "\n%d waiting. Record one with:\n  forgelore review --accept <fingerprint> --title \"what fixed it\"\n", len(pending))
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
		return fmt.Errorf("--accept needs --title: nothing proposed one, and Forgelore saw the error stop, not why")
	}

	body := chosen.Body
	if body == "" && chosen.Command != "" {
		body = "Noticed when `" + chosen.Command + "` started working again."
	}

	id, err := record.NewULID()
	if err != nil {
		return err
	}
	r := &record.Record{
		Schema:      record.CurrentSchema,
		ID:          id,
		Type:        record.TypeFix,
		Scope:       record.ScopeLocal,
		Title:       title,
		Created:     time.Now().UTC().Truncate(time.Second),
		Source:      record.SourceHook,
		Tainted:     chosen.Tainted,
		Fingerprint: chosen.Fingerprint,
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
