package agent

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// MappingVersion is the format this package understands. A mapping file
// carries its own number so that a file written for a later Forgelore is
// refused with a sentence rather than misread (K5, ADR-0011).
const MappingVersion = 1

// Mapping is one agent's hook format, as data.
//
// It is JSON and not the restricted YAML of ADR-0017: that dialect has no
// nested maps on purpose, and a mapping is nested by nature — an event holds
// field paths, which hold alternatives.
type Mapping struct {
	MappingVersion int    `json:"mapping_version"`
	Agent          string `json:"agent"`

	// VerifiedAgainst names the agent version whose real captured payloads
	// this mapping was tested against. Empty means nobody has checked it
	// against anything but documentation, and `doctor` says so.
	VerifiedAgainst string `json:"verified_against"`

	Common   Common   `json:"common"`
	Events   []Match  `json:"events"`
	Response Response `json:"response"`
}

// Response is how this agent wants context handed back.
//
// The shape differs more than the input does. Claude Code reads
// `hookSpecificOutput.additionalContext`; Copilot CLI reads
// `additionalContext` at the top level. Hard-coding either one would have
// meant a code change per agent, which is what K5 exists to prevent.
type Response struct {
	// ContextPath is where the text goes, as dotted keys. The object is
	// built from it.
	ContextPath string `json:"context_path"`

	// EventNamePath, when set, is where the agent's own event name is
	// echoed back. Claude Code requires it; nothing else seen so far does.
	EventNamePath string `json:"event_name_path,omitempty"`
}

// Common are the fields every payload of this agent carries.
type Common struct {
	// EventName is where the payload names its own event. It may be empty:
	// Copilot CLI sends no such field, and the event is known only from
	// which hook entry fired. The caller then supplies the name.
	EventName  Path `json:"event_name"`
	Session    Path `json:"session"`
	Cwd        Path `json:"cwd"`
	Scratchpad Path `json:"scratchpad"`
}

// Match turns one of the agent's events into a canonical one.
type Match struct {
	AgentEvent string `json:"agent_event"`
	Kind       Kind   `json:"kind"`

	// Tools limits the match to these tool names. Empty matches any.
	Tools []string `json:"tools,omitempty"`

	Tool    Path `json:"tool,omitempty"`
	Command Path `json:"command,omitempty"`
	Output  Path `json:"output,omitempty"`

	// Failure decides whether a finished command failed. An agent that has
	// a dedicated failure event sets Always; one that reports the outcome
	// inside the payload needs a test.
	Failure Failure `json:"failure,omitempty"`

	// SkipWhen names a field that, when present and not false, means this
	// event is not worth acting on. Claude Code sets is_interrupt on a
	// command the user stopped; that is somebody pressing Ctrl-C, not an
	// error anyone wants remembered.
	SkipWhen Path `json:"skip_when,omitempty"`

	// Untrusted marks every event from this match as coming from outside
	// the project (ADR-0013).
	Untrusted bool `json:"untrusted,omitempty"`
}

// Failure is how an agent says a command did not work.
type Failure struct {
	// Always: reaching this event is itself the failure.
	Always bool `json:"always,omitempty"`
	// OutputMatches: the output says so. The pattern is the uncertain part
	// of any mapping and the reason this is data.
	OutputMatches string `json:"output_matches,omitempty"`
	// FieldPresent: a field exists and is non-empty.
	FieldPresent Path `json:"field_present,omitempty"`
}

// Path is where a value lives in the payload, as dotted keys.
//
// It is a list rather than one path because an agent can report the same
// thing in more than one shape — a tool result may be a plain string in one
// version and an object in the next — and a mapping that lists both keeps
// working across the change. The first path that resolves wins.
type Path []string

// UnmarshalJSON accepts a single string as well as a list, because a mapping
// with one path per field is the common case and should read like one.
func (p *Path) UnmarshalJSON(data []byte) error {
	var one string
	if err := json.Unmarshal(data, &one); err == nil {
		*p = Path{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return fmt.Errorf("a field path must be a string or a list of strings")
	}
	*p = many
	return nil
}

// ParseMapping reads a mapping file.
func ParseMapping(data []byte) (*Mapping, error) {
	var m Mapping
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("agent: reading the mapping: %w", err)
	}
	switch {
	case m.MappingVersion == 0:
		return nil, fmt.Errorf("agent: the mapping has no mapping_version")
	case m.MappingVersion > MappingVersion:
		return nil, fmt.Errorf("agent: mapping version %d is newer than this build understands (%d)",
			m.MappingVersion, MappingVersion)
	case m.Agent == "":
		return nil, fmt.Errorf("agent: the mapping does not say which agent it is for")
	case len(m.Events) == 0:
		return nil, fmt.Errorf("agent: the mapping maps no events")
	case m.Response.ContextPath == "":
		return nil, fmt.Errorf("agent: the mapping does not say where context goes")
	}
	for i, ev := range m.Events {
		if ev.AgentEvent == "" {
			return nil, fmt.Errorf("agent: event %d has no agent_event", i)
		}
		switch ev.Kind {
		case SessionStarted, CommandFailed, CommandSucceeded, SessionEnded:
		case "":
			return nil, fmt.Errorf("agent: %s has no kind", ev.AgentEvent)
		default:
			return nil, fmt.Errorf("agent: %s maps to unknown kind %q", ev.AgentEvent, ev.Kind)
		}
		if ev.Failure.OutputMatches != "" {
			if _, err := regexp.Compile(ev.Failure.OutputMatches); err != nil {
				return nil, fmt.Errorf("agent: %s has an unusable failure pattern: %w", ev.AgentEvent, err)
			}
		}
	}
	return &m, nil
}

// Translate turns one payload into a canonical event.
//
// A payload the mapping does not cover is not an error: an agent fires many
// more events than Forgelore acts on, and the ones it ignores are the normal
// case. ok is false and the caller does nothing.
// callerEvent is what invoked the hook, used when the payload does not name
// its own event.
func (m *Mapping) Translate(payload []byte, callerEvent string, now time.Time) (Event, bool, error) {
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return Event{}, false, fmt.Errorf("agent: that is not a hook payload: %w", err)
	}

	name, _ := lookup(raw, m.Common.EventName)
	if name == "" {
		name = callerEvent
	}
	if name == "" {
		return Event{}, false, fmt.Errorf(
			"agent: the payload does not name its event, and none was given (--event)")
	}

	for _, match := range m.Events {
		if match.AgentEvent != name {
			continue
		}
		tool, _ := lookup(raw, match.Tool)
		if len(match.Tools) > 0 && !contains(match.Tools, tool) {
			continue
		}
		if v, ok := lookup(raw, match.SkipWhen); ok && v != "false" {
			return Event{}, false, nil
		}

		e := Event{Kind: match.Kind, Time: now, Tool: tool, AgentEvent: name, Untrusted: match.Untrusted}
		e.Session, _ = lookup(raw, m.Common.Session)
		e.Cwd, _ = lookup(raw, m.Common.Cwd)
		e.Scratchpad, _ = lookup(raw, m.Common.Scratchpad)
		e.Command, _ = lookup(raw, match.Command)
		e.Output, _ = lookup(raw, match.Output)

		// A match that maps to one outcome says so; one that has to read
		// the payload decides here.
		if match.Kind == CommandFailed || match.Kind == CommandSucceeded {
			if match.failed(raw, e.Output) {
				e.Kind = CommandFailed
			} else {
				e.Kind = CommandSucceeded
			}
		}
		return e, true, nil
	}
	return Event{}, false, nil
}

func (match Match) failed(raw map[string]any, output string) bool {
	if match.Failure.Always {
		return true
	}
	if len(match.Failure.FieldPresent) > 0 {
		if v, ok := lookup(raw, match.Failure.FieldPresent); ok && v != "" {
			return true
		}
	}
	if match.Failure.OutputMatches != "" {
		// Already compiled once in ParseMapping.
		if re, err := regexp.Compile(match.Failure.OutputMatches); err == nil && re.MatchString(output) {
			return true
		}
	}
	return false
}

// lookup walks dotted keys and returns the first path that resolves.
//
// A value that is not a string is returned as its JSON, which is what keeps a
// mapping working when an agent turns a plain string field into an object:
// the text is still there to be fingerprinted.
func lookup(raw map[string]any, paths Path) (string, bool) {
	for _, path := range paths {
		var cur any = raw
		ok := true
		for _, key := range strings.Split(path, ".") {
			node, isMap := cur.(map[string]any)
			if !isMap {
				ok = false
				break
			}
			cur, ok = node[key]
			if !ok {
				break
			}
		}
		if !ok || cur == nil {
			continue
		}
		switch v := cur.(type) {
		case string:
			if v != "" {
				return v, true
			}
		default:
			if encoded, err := json.Marshal(v); err == nil {
				return string(encoded), true
			}
		}
	}
	return "", false
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// Context builds the reply that hands text back to the agent, in whatever
// shape its mapping says it reads.
func (m *Mapping) Context(agentEvent, text string) ([]byte, error) {
	reply := map[string]any{}
	setPath(reply, m.Response.ContextPath, text)
	if m.Response.EventNamePath != "" && agentEvent != "" {
		setPath(reply, m.Response.EventNamePath, agentEvent)
	}
	data, err := json.Marshal(reply)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// setPath writes a value at dotted keys, creating the objects on the way.
func setPath(root map[string]any, path string, value any) {
	keys := strings.Split(path, ".")
	node := root
	for _, key := range keys[:len(keys)-1] {
		next, ok := node[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			node[key] = next
		}
		node = next
	}
	node[keys[len(keys)-1]] = value
}
