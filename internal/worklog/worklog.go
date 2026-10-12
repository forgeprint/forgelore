// Package worklog reads what an agent wrote about a session and keeps the
// facts.
//
// An agent writes a transcript of its own session. No agent documents that
// file's format, and most of it is the agent's internal bookkeeping, so this
// package reads one layer of it and ignores the rest: the entries shaped like
// the Messages API wire format, which is the only part not invented by a
// single agent. A tool call is there with its input, a tool result is there
// with `is_error`, and that is enough to say which files were edited and
// which commands passed. Nothing here calls a model, and nothing here writes
// anything (ADR-0026).
package worklog

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"
)

// Cap is how much of a transcript is read by default. One measured session
// was 17 MB, so a cap is not a corner case; past it the tail is read, because
// how a session ended is worth more than how it opened.
const Cap = 8 << 20

// maxLine is the longest line kept. A transcript line holding the agent's
// whole system prompt is not a fact about the session, and a line longer than
// this is skipped rather than allowed to decide how much memory this takes.
const maxLine = 4 << 20

// A Session is what one transcript says happened.
type Session struct {
	ID     string    `json:"session,omitempty"`
	Agent  string    `json:"agent_version,omitempty"`
	Cwd    string    `json:"cwd,omitempty"`
	Branch string    `json:"branch,omitempty"`
	Start  time.Time `json:"start,omitempty"`
	End    time.Time `json:"end,omitempty"`

	// Edits are file paths in the order they were first touched.
	Edits []string `json:"edits,omitempty"`

	// Commands are shell commands in the order they ran.
	Commands []Command `json:"commands,omitempty"`

	// Discovery is how many of this session's tool calls were the agent
	// finding its way around: reading, searching and listing. It is the
	// proxy work memory is measured by — a session that already knows
	// where things are should not have to look (ADR-0019).
	Discovery int `json:"discovery,omitempty"`

	// Calls is every tool call, so that Discovery has a denominator.
	Calls int `json:"calls,omitempty"`

	// CallsBeforeFirstEdit is how long it took to start changing
	// something. Reading for twenty calls before the first edit is the
	// shape of a session that is rediscovering a codebase.
	CallsBeforeFirstEdit int `json:"calls_before_first_edit,omitempty"`

	// Closing is the last thing the agent said in its own words.
	//
	// The agent's account of a session is worth more than any summary this
	// package could assemble, and it is already written: the final text
	// block of the final assistant entry. Claude Code documents a
	// `last_assistant_message` field on its hook payloads for the same
	// thing, but the payload captured from 2.1.291 does not carry it, and
	// the transcript does — verified against a real one rather than taken
	// from documentation (ADR-0026).
	Closing string `json:"closing,omitempty"`

	// Truncated says the cap was reached and the beginning was not read, so
	// an absent edit is not evidence that nothing was edited.
	Truncated bool `json:"truncated,omitempty"`

	// Unreadable counts lines that were not JSON or were too long. A
	// transcript is an undocumented format; a count is more honest than
	// silence and cheaper than a guess at what was in them.
	Unreadable int `json:"unreadable,omitempty"`

	// calls maps a tool call's id to the command it started, so that the
	// result arriving in a later entry can be attributed to it.
	calls map[string]int
}

// A Command is one shell command and how it ended.
type Command struct {
	Text   string    `json:"text"`
	Failed bool      `json:"failed,omitempty"`
	At     time.Time `json:"at,omitempty"`
}

// Fixed returns commands that failed and later ran again without failing.
// This is the one inference in this package, and it is the cheapest useful
// one: the trace of a fix, without any claim about what the fix was.
func (s *Session) Fixed() []string {
	failed := map[string]bool{}
	var out []string
	seen := map[string]bool{}
	for _, c := range s.Commands {
		switch {
		case c.Failed:
			failed[c.Text] = true
		case failed[c.Text] && !seen[c.Text]:
			seen[c.Text] = true
			out = append(out, c.Text)
		}
	}
	return out
}

// ReadFile reads a transcript, taking the last limit bytes of a file larger
// than that. A limit of zero means Cap.
func ReadFile(path string, limit int64) (*Session, error) {
	if limit <= 0 {
		limit = Cap
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	truncated := false
	if info.Size() > limit {
		if _, err := f.Seek(info.Size()-limit, io.SeekStart); err != nil {
			return nil, err
		}
		truncated = true
	}

	s, err := Parse(f, truncated)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Parse reads a transcript from r. When partial is true the reader starts
// mid-file, so the first line is a fragment and is dropped.
func Parse(r io.Reader, partial bool) (*Session, error) {
	s := &Session{Truncated: partial}
	br := bufio.NewReaderSize(r, 64<<10)
	first := true

	for {
		line, long, err := readLine(br)
		if len(line) == 0 && err != nil {
			if err == io.EOF {
				return s, nil
			}
			return s, err
		}
		if long {
			s.Unreadable++
			if err != nil {
				return s, nil
			}
			continue
		}
		if first {
			first = false
			if partial {
				// Half a line is not a fact.
				if err != nil {
					return s, nil
				}
				continue
			}
		}
		if len(line) > 0 {
			s.read(line)
		}
		if err != nil {
			return s, nil
		}
	}
}

// entry is the part of a transcript line this package admits to knowing
// about. Everything else in the file is the agent's own bookkeeping.
type entry struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId"`
	Version   string `json:"version"`
	Cwd       string `json:"cwd"`
	Branch    string `json:"gitBranch"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// block is one Messages API content block.
type block struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Name  string `json:"name"`
	ID    string `json:"id"`
	Input struct {
		FilePath string `json:"file_path"`
		Command  string `json:"command"`
	} `json:"input"`
	ToolUseID string `json:"tool_use_id"`
	IsError   *bool  `json:"is_error"`
}

func (s *Session) read(line []byte) {
	var e entry
	if json.Unmarshal(line, &e) != nil {
		s.Unreadable++
		return
	}
	if e.Type != "user" && e.Type != "assistant" {
		return
	}
	s.note(e)

	var blocks []block
	// A plain prompt is a string rather than a list of blocks, and a string
	// holds nothing this package wants.
	if json.Unmarshal(e.Message.Content, &blocks) != nil {
		return
	}
	at, _ := time.Parse(time.RFC3339, e.Timestamp)
	for _, b := range blocks {
		switch b.Type {
		case "tool_use":
			s.tool(b, at)
		case "tool_result":
			s.result(b)
		case "text":
			if e.Type == "assistant" && strings.TrimSpace(b.Text) != "" {
				s.Closing = strings.TrimSpace(b.Text)
			}
		}
	}
}

// note records what every entry agrees on. The first entry seen wins for the
// session's identity, because a transcript read from its tail has no first
// entry to trust.
func (s *Session) note(e entry) {
	if s.ID == "" {
		s.ID = e.SessionID
	}
	if e.Version != "" {
		s.Agent = e.Version
	}
	if s.Cwd == "" {
		s.Cwd = e.Cwd
	}
	if s.Branch == "" {
		s.Branch = e.Branch
	}
	if at, err := time.Parse(time.RFC3339, e.Timestamp); err == nil {
		if s.Start.IsZero() || at.Before(s.Start) {
			s.Start = at
		}
		if at.After(s.End) {
			s.End = at
		}
	}
}

// tool records an edit or a command.
//
// An edit is decided by the tool's name and not by the presence of a
// file_path: Read carries one too, and counting a file somebody looked at as a
// file somebody changed was wrong in both directions — it inflated what a
// session did and it hid how much of the session was looking. The names are
// Claude Code's, which is the only transcript format with a parser here.
func (s *Session) tool(b block, at time.Time) {
	s.Calls++
	edit := editTools[b.Name]
	if len(s.Edits) == 0 && !edit {
		s.CallsBeforeFirstEdit = s.Calls
	}
	if discovering(b.Name, b.Input.Command) {
		s.Discovery++
	}
	switch {
	case edit && b.Input.FilePath != "":
		for _, seen := range s.Edits {
			if seen == b.Input.FilePath {
				return
			}
		}
		s.Edits = append(s.Edits, b.Input.FilePath)
	case b.Input.Command != "":
		s.Commands = append(s.Commands, Command{Text: b.Input.Command, At: at})
		if s.calls == nil {
			s.calls = map[string]int{}
		}
		s.calls[b.ID] = len(s.Commands) - 1
	}
}

func (s *Session) result(b block) {
	i, ok := s.calls[b.ToolUseID]
	if !ok || b.IsError == nil {
		return
	}
	s.Commands[i].Failed = *b.IsError
}

// editTools are the tools that change a file.
var editTools = map[string]bool{
	"Edit": true, "Write": true, "MultiEdit": true, "NotebookEdit": true,
}

// discoveryTools are the tools whose whole job is finding out where
// something is.
var discoveryTools = map[string]bool{
	"Read": true, "Grep": true, "Glob": true, "NotebookRead": true,
}

// discoveryCommands are the shell commands that do the same job. Counting
// only the tools would undercount badly: in this project's own transcripts
// the agent reads files with `sed -n` and searches with `grep` far more often
// than it calls Read, so a measurement that ignored the shell would make work
// memory look effective by missing the looking.
//
// The cost is that this is pattern matching on a command line, and a command
// matching one of these words is not always looking something up — `grep -c`
// inside a test is counted too. It is a proxy and the report says so.
var discoveryCommands = []string{
	"cat ", "head ", "tail ", "sed -n", "less ", "grep", "rg ", "ag ",
	"find ", "ls ", "tree ", "wc -l", "awk ", "git log", "git show",
	"git diff", "git blame", "git grep",
}

// discovering reports whether one tool call was the agent looking rather than
// doing.
func discovering(tool, command string) bool {
	if discoveryTools[tool] {
		return true
	}
	if command == "" {
		return false
	}
	// Only the first command in a chain decides. A chain that starts by
	// looking and goes on to build is one call either way, and reading it
	// as both would count a build as discovery.
	first := strings.TrimSpace(command)
	for _, sep := range []string{"&&", "||", ";", "|"} {
		if i := strings.Index(first, sep); i >= 0 {
			first = strings.TrimSpace(first[:i])
		}
	}
	for _, c := range discoveryCommands {
		if strings.HasPrefix(first, c) || strings.HasPrefix(first, strings.TrimSpace(c)+"\n") {
			return true
		}
	}
	return false
}

// readLine reads one line without letting its length decide how much memory
// this costs. A line longer than maxLine is consumed and reported as long,
// not returned.
func readLine(br *bufio.Reader) (line []byte, long bool, err error) {
	for {
		chunk, more, err := br.ReadLine()
		if len(line)+len(chunk) <= maxLine {
			line = append(line, chunk...)
		} else {
			long = true
		}
		if !more {
			return line, long, err
		}
		if err != nil {
			return line, long, err
		}
	}
}
