package agent

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var now = time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)

func mustBuilt(t *testing.T) *Mapping {
	t.Helper()
	m, err := Built("claude-code")
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestBuiltMappingParses(t *testing.T) {
	m := mustBuilt(t)
	if m.Agent != "claude-code" || m.MappingVersion != MappingVersion {
		t.Errorf("agent=%q version=%d", m.Agent, m.MappingVersion)
	}
	if _, err := Built("nope"); err == nil {
		t.Error("an unknown agent returned a mapping")
	}
}

// TestBuiltMappingNamesTheVersionItWasCheckedAgainst: verified_against is a
// claim, and a claim has to have a corpus behind it. Setting it to a version
// with no captured payloads fails here.
func TestBuiltMappingNamesTheVersionItWasCheckedAgainst(t *testing.T) {
	v := mustBuilt(t).VerifiedAgainst
	if v == "" {
		t.Fatal("verified_against is empty; the corpus says otherwise")
	}
	if _, err := os.Stat(filepath.Join(corpusRoot, v)); err != nil {
		t.Errorf("verified_against is %q but there are no payloads for it: %v", v, err)
	}
}

func TestTranslateSessionStart(t *testing.T) {
	payload := `{"hook_event_name":"SessionStart","session_id":"s1","cwd":"/w","source":"startup"}`
	e, ok, err := mustBuilt(t).Translate([]byte(payload), now)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Kind != SessionStarted || e.Session != "s1" || e.Cwd != "/w" {
		t.Errorf("%+v", e)
	}
}

func TestTranslateFailure(t *testing.T) {
	payload := `{
	  "hook_event_name":"PostToolUseFailure","session_id":"s1","cwd":"/w",
	  "scratchpad_dir":"/tmp/scratch",
	  "tool_name":"Bash","tool_input":{"command":"go build ./..."},
	  "tool_response":"./main.go:5:14: undefined: greet"
	}`
	e, ok, err := mustBuilt(t).Translate([]byte(payload), now)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Kind != CommandFailed {
		t.Errorf("kind = %q", e.Kind)
	}
	if e.Command != "go build ./..." || !strings.Contains(e.Output, "undefined: greet") {
		t.Errorf("command=%q output=%q", e.Command, e.Output)
	}
	if e.Scratchpad != "/tmp/scratch" {
		t.Errorf("scratchpad = %q", e.Scratchpad)
	}
}

// TestObjectToolResponseStillYieldsText is why a field path is a list. The
// documentation shows tool_response as a plain string; nothing promises it
// stays one, and a mapping that lists both shapes survives the change.
func TestObjectToolResponseStillYieldsText(t *testing.T) {
	payload := `{
	  "hook_event_name":"PostToolUseFailure","session_id":"s1",
	  "tool_name":"Bash","tool_input":{"command":"go build ./..."},
	  "tool_response":{"stderr":"./main.go:5:14: undefined: greet","exitCode":1}
	}`
	e, ok, err := mustBuilt(t).Translate([]byte(payload), now)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if !strings.Contains(e.Output, "undefined: greet") {
		t.Errorf("output = %q", e.Output)
	}
}

// TestSuccessIsNotMistakenForFailure: with the shape of a failed Bash result
// unverified, the built-in mapping treats only the dedicated failure event as
// a failure. The cost is a missed injection; the alternative is a hint after
// a command that worked.
func TestSuccessIsNotMistakenForFailure(t *testing.T) {
	payload := `{
	  "hook_event_name":"PostToolUse","session_id":"s1",
	  "tool_name":"Bash","tool_input":{"command":"go build ./..."},
	  "tool_response":"error: nothing is wrong, this word just appears"
	}`
	e, ok, err := mustBuilt(t).Translate([]byte(payload), now)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Kind != CommandSucceeded {
		t.Errorf("kind = %q, want %q", e.Kind, CommandSucceeded)
	}
}

func TestToolFilter(t *testing.T) {
	payload := `{"hook_event_name":"PostToolUse","session_id":"s1","tool_name":"Write","tool_input":{}}`
	if _, ok, _ := mustBuilt(t).Translate([]byte(payload), now); ok {
		t.Error("a Write matched a Bash-only event")
	}
}

// TestUnmappedEventIsNotAnError: an agent fires far more events than
// Forgelore acts on, and ignoring them is the normal case.
func TestUnmappedEventIsNotAnError(t *testing.T) {
	payload := `{"hook_event_name":"Notification","session_id":"s1"}`
	_, ok, err := mustBuilt(t).Translate([]byte(payload), now)
	if err != nil {
		t.Errorf("an unmapped event errored: %v", err)
	}
	if ok {
		t.Error("an unmapped event produced a canonical event")
	}
}

func TestTranslateRejectsRubbish(t *testing.T) {
	for _, payload := range []string{"not json", `{"no":"event name"}`} {
		if _, _, err := mustBuilt(t).Translate([]byte(payload), now); err == nil {
			t.Errorf("%q was accepted", payload)
		}
	}
}

func TestParseMappingRejectsBadFiles(t *testing.T) {
	cases := map[string]string{
		"no version":      `{"agent":"a","events":[{"agent_event":"X","kind":"session_started"}]}`,
		"future version":  `{"mapping_version":99,"agent":"a","events":[{"agent_event":"X","kind":"session_started"}]}`,
		"no agent":        `{"mapping_version":1,"events":[{"agent_event":"X","kind":"session_started"}]}`,
		"no events":       `{"mapping_version":1,"agent":"a","events":[]}`,
		"unknown kind":    `{"mapping_version":1,"agent":"a","events":[{"agent_event":"X","kind":"wat"}]}`,
		"no kind":         `{"mapping_version":1,"agent":"a","events":[{"agent_event":"X"}]}`,
		"no agent_event":  `{"mapping_version":1,"agent":"a","events":[{"kind":"session_started"}]}`,
		"bad regexp":      `{"mapping_version":1,"agent":"a","events":[{"agent_event":"X","kind":"command_failed","failure":{"output_matches":"("}}]}`,
		"not json at all": `{`,
	}
	for why, data := range cases {
		if _, err := ParseMapping([]byte(data)); err == nil {
			t.Errorf("%s was accepted", why)
		}
	}
}

// TestAFutureMappingSaysSo: a mapping written for a later Forgelore is
// refused with a sentence rather than half-read.
func TestFutureMappingNamesTheProblem(t *testing.T) {
	_, err := ParseMapping([]byte(`{"mapping_version":99,"agent":"a","events":[{"agent_event":"X","kind":"session_started"}]}`))
	if err == nil || !strings.Contains(err.Error(), "newer than this build") {
		t.Errorf("err = %v", err)
	}
}

func TestPathAcceptsStringOrList(t *testing.T) {
	m, err := ParseMapping([]byte(`{
	  "mapping_version":1,"agent":"a",
	  "response":{"context_path":"additionalContext"},
	  "common":{"event_name":"e","session":["a.b","s"]},
	  "events":[{"agent_event":"X","kind":"session_started"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Common.EventName) != 1 || len(m.Common.Session) != 2 {
		t.Errorf("%+v", m.Common)
	}

	e, ok, err := m.Translate([]byte(`{"e":"X","s":"fallback"}`), now)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Session != "fallback" {
		t.Errorf("the second path was not tried: %q", e.Session)
	}
}

func TestFailureByFieldPresence(t *testing.T) {
	m, err := ParseMapping([]byte(`{
	  "mapping_version":1,"agent":"a",
	  "response":{"context_path":"additionalContext"},
	  "common":{"event_name":"e","session":"s"},
	  "events":[{"agent_event":"X","kind":"command_succeeded","failure":{"field_present":"err"}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	failed, _, _ := m.Translate([]byte(`{"e":"X","s":"1","err":"boom"}`), now)
	if failed.Kind != CommandFailed {
		t.Errorf("kind = %q", failed.Kind)
	}
	ok, _, _ := m.Translate([]byte(`{"e":"X","s":"1"}`), now)
	if ok.Kind != CommandSucceeded {
		t.Errorf("kind = %q", ok.Kind)
	}
}

// TestStatePathIsNotChosenByTheAgent: the session id arrives from outside and
// ends up in a filename, so it is hashed rather than pasted in.
func TestStatePathIsNotChosenByTheAgent(t *testing.T) {
	e := Event{Session: "../../etc/passwd", Scratchpad: "/tmp/scratch"}
	path := StatePath(e, "/cache")
	if strings.Contains(path, "..") {
		t.Errorf("a traversal survived into %q", path)
	}
	if filepath.Dir(path) != "/tmp/scratch" {
		t.Errorf("the file escaped the scratch directory: %q", path)
	}
}

func TestStateFallsBackToTheProjectCache(t *testing.T) {
	path := StatePath(Event{Session: "s1"}, "/cache")
	if filepath.Dir(path) != filepath.Join("/cache", "sessions") {
		t.Errorf("path = %q", path)
	}
}

func TestStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s := LoadState(path)
	if len(s.Failed) != 0 {
		t.Errorf("a missing file was not empty: %+v", s)
	}
	s.Failed["abcd"] = "go build ./..."
	if err := SaveState(path, s); err != nil {
		t.Fatal(err)
	}
	if got := LoadState(path); got.Failed["abcd"] != "go build ./..." {
		t.Errorf("%+v", got)
	}
}

// TestCorruptStateIsEmptyNotFatal: losing it costs a candidate, and failing
// here would cost the hook.
func TestCorruptStateIsEmptyNotFatal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := SaveState(path, &State{Failed: map[string]string{"a": "b"}}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{half"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := LoadState(path); len(got.Failed) != 0 {
		t.Errorf("%+v", got)
	}
}

// TestContextShapeComesFromTheMapping is the gap Copilot CLI opened: Claude
// Code reads context inside a wrapper and Copilot reads it at the top level,
// and neither should need a line of Go.
func TestContextShapeComesFromTheMapping(t *testing.T) {
	claude, err := Built("claude-code")
	if err != nil {
		t.Fatal(err)
	}
	reply, err := claude.Context("PostToolUseFailure", "a hint")
	if err != nil {
		t.Fatal(err)
	}
	var nested struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(reply, &nested); err != nil {
		t.Fatalf("%v\n%s", err, reply)
	}
	if nested.HookSpecificOutput.AdditionalContext != "a hint" {
		t.Errorf("claude-code reply = %s", reply)
	}
	if nested.HookSpecificOutput.HookEventName != "PostToolUseFailure" {
		t.Errorf("the event name was not echoed: %s", reply)
	}

	copilot, err := Built("copilot-cli")
	if err != nil {
		t.Fatal(err)
	}
	reply, err = copilot.Context("postToolUseFailure", "a hint")
	if err != nil {
		t.Fatal(err)
	}
	var flat struct {
		AdditionalContext string `json:"additionalContext"`
	}
	if err := json.Unmarshal(reply, &flat); err != nil {
		t.Fatalf("%v\n%s", err, reply)
	}
	if flat.AdditionalContext != "a hint" {
		t.Errorf("copilot-cli reply = %s", reply)
	}
	if bytes.Contains(reply, []byte("hookSpecificOutput")) {
		t.Errorf("copilot-cli got Claude Code's wrapper: %s", reply)
	}
	if !bytes.HasSuffix(reply, []byte("\n")) {
		t.Error("the reply is not newline terminated")
	}
}

// TestDocumentedPayloadsTranslate pins what the documentation says each
// agent sends, for the two mappings no captured payload backs yet.
//
// Passing here is not verification. It is a record of what was believed on
// the day the mapping was written, so that the first real payload shows up
// as a difference rather than as a mystery.
func TestDocumentedPayloadsTranslate(t *testing.T) {
	cases := []struct {
		agent   string
		payload string
		want    Kind
		output  string
	}{
		{
			agent: "copilot-cli",
			payload: `{"hookEventName":"postToolUseFailure","sessionId":"s1","cwd":"/w",
			           "toolName":"bash","toolArgs":{"command":"go build ./..."},
			           "error":"./main.go:5:14: undefined: greet"}`,
			want:   CommandFailed,
			output: "undefined: greet",
		},
		{
			agent: "copilot-cli",
			payload: `{"hookEventName":"postToolUse","sessionId":"s1","cwd":"/w",
			           "toolName":"bash","toolArgs":{"command":"go version"},
			           "toolResult":{"resultType":"success","textResultForLlm":"go1.27.1"}}`,
			want:   CommandSucceeded,
			output: "go1.27.1",
		},
		{
			agent:   "copilot-cli",
			payload: `{"hook_event_name":"sessionStart","session_id":"s1","cwd":"/w"}`,
			want:    SessionStarted,
		},
		{
			agent: "codex-cli",
			payload: `{"hook_event_name":"PostToolUse","session_id":"s1","cwd":"/w",
			           "tool_name":"shell","tool_input":{"command":"go build ./..."},
			           "error":"./main.go:5:14: undefined: greet"}`,
			want:   CommandFailed,
			output: "undefined: greet",
		},
		{
			agent: "codex-cli",
			payload: `{"hook_event_name":"PostToolUse","session_id":"s1","cwd":"/w",
			           "tool_name":"shell","tool_input":{"command":"go version"},
			           "tool_output":"go1.27.1"}`,
			want:   CommandSucceeded,
			output: "go1.27.1",
		},
	}

	for _, c := range cases {
		m, err := Built(c.agent)
		if err != nil {
			t.Fatal(err)
		}
		e, ok, err := m.Translate([]byte(c.payload), now)
		if err != nil || !ok {
			t.Errorf("%s: ok=%v err=%v", c.agent, ok, err)
			continue
		}
		if e.Kind != c.want {
			t.Errorf("%s: kind = %q, want %q", c.agent, e.Kind, c.want)
		}
		if e.Session != "s1" || e.Cwd != "/w" {
			t.Errorf("%s: session=%q cwd=%q", c.agent, e.Session, e.Cwd)
		}
		if c.output != "" && !strings.Contains(e.Output, c.output) {
			t.Errorf("%s: output = %q, want it to contain %q", c.agent, e.Output, c.output)
		}
	}
}

// TestCodexInterruptIsIgnored: the mapping decides a Codex failure from an
// error field, so a stopped command must be excluded some other way.
func TestCodexInterruptIsIgnored(t *testing.T) {
	m, err := Built("codex-cli")
	if err != nil {
		t.Fatal(err)
	}
	payload := `{"hook_event_name":"PostToolUse","session_id":"s1","cwd":"/w",
	             "tool_name":"shell","tool_input":{"command":"go test ./..."},
	             "error":"interrupted","interrupted":true}`
	if _, ok, _ := m.Translate([]byte(payload), now); ok {
		t.Error("an interrupted command produced an event")
	}
}
