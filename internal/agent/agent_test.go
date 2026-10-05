package agent

import (
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

// TestBuiltMappingIsHonestAboutVerification: the mapping was written from
// documentation, not from captured payloads, and has to say so until real
// samples exist.
func TestBuiltMappingIsHonestAboutVerification(t *testing.T) {
	if v := mustBuilt(t).VerifiedAgainst; v != "" {
		t.Errorf("verified_against = %q; set it only when real payloads back it", v)
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
