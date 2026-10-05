package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/forgeprint/forgelore/internal/candidate"
	"github.com/forgeprint/forgelore/internal/fingerprint"
	"github.com/forgeprint/forgelore/internal/record"
	"github.com/forgeprint/forgelore/internal/store"
)

// defaultSearchLimit matches the CLI's.
const defaultSearchLimit = 20

// toolCacheTTLMillis is how long a client may hold the tool list. The set
// never changes while a process runs, but a new Forgelore could be installed
// under a long-lived client, so the hint expires rather than being eternal.
const toolCacheTTLMillis = 300000

// The tool set. Four tools, and the shape of them is the access control:
// `get` takes an id, and the only place an id comes from is `search` or
// `recall_error`, so no call can return the whole store (ADR-0006).
//
// Descriptions lead with what the tool does, because a client is free to
// truncate them and the first clause is the part that always survives.
func (s *Server) listTools() map[string]any {
	object := func(props map[string]any, required ...string) map[string]any {
		schema := map[string]any{"type": "object", "properties": props}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema
	}
	str := func(desc string) map[string]any {
		return map[string]any{"type": "string", "description": desc}
	}

	return map[string]any{
		"tools": []map[string]any{
			{
				"name":        "search",
				"title":       "Search project memory",
				"description": "Search this project's memory. Returns ids and titles only; pass an id to `get` for the detail.",
				"inputSchema": object(map[string]any{
					"query": str("Words to look for in titles and bodies."),
					"limit": map[string]any{"type": "integer", "description": "At most this many results (default 20)."},
				}, "query"),
			},
			{
				"name":        "get",
				"title":       "Read one memory",
				"description": "Read one memory in full, by an id that came from `search` or `recall_error`.",
				"inputSchema": object(map[string]any{
					"id": str("The record id."),
				}, "id"),
			},
			{
				"name":        "recall_error",
				"title":       "Look up a failed command",
				"description": "Look up what is known about a command's error output. Fingerprints the error and returns any fix or dead end recorded against it. Nothing known means nothing is returned.",
				"inputSchema": object(map[string]any{
					"output":  str("The failing command's output, verbatim."),
					"command": str("The command that produced it. Pass it whenever you have it: the tool it names is part of the fingerprint, so leaving it out can miss a memory recorded with it."),
				}, "output"),
			},
			{
				"name":        "propose",
				"title":       "Propose a memory",
				"description": "Propose a fix for review. Nothing is recorded until a person accepts it with `forgelore review`.",
				"inputSchema": object(map[string]any{
					"title":       str("One line stating the fix, not the problem. This is what gets injected later."),
					"fingerprint": str("The fingerprint from `recall_error`, when the proposal answers a specific error."),
					"command":     str("The command this concerns."),
					"body":        str("Longer explanation, fetched only on request."),
					"tainted":     map[string]any{"type": "boolean", "description": "True when this came from outside the project, such as a web page."},
				}, "title"),
			},
		},
		// The set is fixed at build time, so it is worth caching. The two
		// hints travel together: a client that validates its input rejects
		// a cacheScope with no ttlMs, which is how this was found.
		"cacheScope": "public",
		"ttlMs":      toolCacheTTLMillis,
	}
}

type callParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (s *Server) callTool(raw json.RawMessage) (map[string]any, *rpcError) {
	var p callParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, errf(codeInvalidParams, "params is not a tools/call request")
	}

	var (
		text       string
		structured any
		err        error
	)
	switch p.Name {
	case "search":
		text, structured, err = s.toolSearch(p.Arguments)
	case "get":
		text, structured, err = s.toolGet(p.Arguments)
	case "recall_error":
		text, structured, err = s.toolRecallError(p.Arguments)
	case "propose":
		text, structured, err = s.toolPropose(p.Arguments)
	default:
		// A protocol error, not a tool error: the model cannot fix a tool
		// that does not exist by trying different arguments.
		return nil, errf(codeInvalidParams, "Unknown tool: %s", p.Name)
	}

	if err != nil {
		// A tool error, deliberately: these are things a model can act on,
		// and the specification wants them in the result so it can.
		return toolResult(err.Error(), nil, true), nil
	}
	return toolResult(text, structured, false), nil
}

func toolResult(text string, structured any, isError bool) map[string]any {
	result := map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
		"isError": isError,
	}
	if structured != nil {
		result["structuredContent"] = structured
	}
	return result
}

type hit struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Scope      string `json:"scope"`
	Superseded bool   `json:"superseded,omitempty"`
}

func (s *Server) toolSearch(raw json.RawMessage) (string, any, error) {
	var args struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	_ = json.Unmarshal(raw, &args)
	if strings.TrimSpace(args.Query) == "" {
		return "", nil, fmt.Errorf("search needs a query")
	}
	if args.Limit <= 0 {
		args.Limit = defaultSearchLimit
	}

	idx, err := s.index()
	if err != nil {
		return "", nil, err
	}
	defer idx.Close()

	found, err := idx.Search(args.Query, args.Limit)
	if err != nil {
		return "", nil, err
	}
	hits := toHits(found)
	if len(hits) == 0 {
		return fmt.Sprintf("Nothing in this project's memory matches %q.", args.Query), hits, nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d result(s). Use `get` with an id for the detail.\n", len(hits))
	for _, h := range hits {
		fmt.Fprintf(&b, "%s  %s  %s\n", h.ID, h.Type, h.Title)
	}
	return b.String(), hits, nil
}

func (s *Server) toolGet(raw json.RawMessage) (string, any, error) {
	var args struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(raw, &args)
	id, ok := record.NormalizeULID(args.ID)
	if !ok {
		return "", nil, fmt.Errorf("%q is not a record id; ids come from `search` or `recall_error`", args.ID)
	}

	r, err := s.store.Get(id)
	if err != nil {
		return "", nil, err
	}
	encoded, err := r.Encode()
	if err != nil {
		return "", nil, err
	}
	return string(encoded), map[string]any{
		"id": r.ID, "type": r.Type, "scope": string(r.Scope), "title": r.Title,
		"fingerprint": r.Fingerprint, "tainted": r.Tainted, "body": r.Body,
	}, nil
}

type recalled struct {
	Fingerprint string `json:"fingerprint"`
	Message     string `json:"message"`
	Matches     []hit  `json:"matches"`
}

func (s *Server) toolRecallError(raw json.RawMessage) (string, any, error) {
	var args struct {
		Output  string `json:"output"`
		Command string `json:"command"`
	}
	_ = json.Unmarshal(raw, &args)
	if strings.TrimSpace(args.Output) == "" {
		return "", nil, fmt.Errorf("recall_error needs the command's output")
	}

	events := fingerprint.Scan(args.Command, args.Output)
	if len(events) == 0 {
		return "No error recognised in that output.", []recalled{}, nil
	}

	idx, err := s.index()
	if err != nil {
		return "", nil, err
	}
	defer idx.Close()

	out := make([]recalled, 0, len(events))
	known := 0
	for _, ev := range events {
		found, err := idx.ByFingerprint(ev.Sum)
		if err != nil {
			continue
		}
		hits := toHits(found)
		if len(hits) > 0 {
			known++
		}
		out = append(out, recalled{Fingerprint: ev.Sum, Message: ev.Message, Matches: hits})
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d error(s), %d with something recorded.\n", len(out), known)
	for _, r := range out {
		fmt.Fprintf(&b, "\n%s  %s\n", r.Fingerprint, r.Message)
		if len(r.Matches) == 0 {
			b.WriteString("  nothing recorded\n")
			continue
		}
		for _, m := range r.Matches {
			fmt.Fprintf(&b, "  %s: %s (%s)\n", m.Type, m.Title, m.ID)
		}
	}
	return b.String(), out, nil
}

func (s *Server) toolPropose(raw json.RawMessage) (string, any, error) {
	var args struct {
		Title       string `json:"title"`
		Fingerprint string `json:"fingerprint"`
		Command     string `json:"command"`
		Body        string `json:"body"`
		Tainted     bool   `json:"tainted"`
	}
	_ = json.Unmarshal(raw, &args)
	if strings.TrimSpace(args.Title) == "" {
		return "", nil, fmt.Errorf("propose needs a title: one line stating the fix")
	}

	c := candidate.Candidate{
		Time:        time.Now().UTC().Truncate(time.Second),
		Fingerprint: args.Fingerprint,
		Command:     args.Command,
		Title:       args.Title,
		Body:        args.Body,
		Tainted:     args.Tainted,
	}
	if err := candidate.Append(s.store.Root(), c); err != nil {
		return "", nil, err
	}
	return "Proposed, and waiting for review. It is not in memory yet and will not be injected; a person accepts it with `forgelore review`.",
		map[string]any{"proposed": true, "fingerprint": c.Fingerprint}, nil
}

// index opens the store's index and brings it up to date.
func (s *Server) index() (*store.Index, error) {
	idx, err := s.store.OpenIndex()
	if err != nil {
		return nil, err
	}
	if _, err := idx.Sync(); err != nil {
		idx.Close()
		return nil, err
	}
	return idx, nil
}

func toHits(found []store.Hit) []hit {
	out := make([]hit, 0, len(found))
	for _, h := range found {
		out = append(out, hit{
			ID: h.ID, Type: h.Type, Title: h.Title,
			Scope: string(h.Scope), Superseded: h.Superseded,
		})
	}
	return out
}
