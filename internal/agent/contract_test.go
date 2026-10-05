package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// corpusRoot holds real hook payloads, captured from a running agent by
// scripts/capture-agent-events.sh. Documentation says what an agent intends
// to send; these files are what it sent.
const corpusRoot = "../../testdata/agents/claude-code"

// withoutSamples lists the mapped events no captured payload covers yet.
//
// It is empty, and keeping it empty is the point: every event the mapping
// claims to handle is held against a payload the agent really sent. An event
// added to the mapping without a sample fails the test below rather than
// resting on documentation, which is how the two tool events were wrong in
// the first place.
var withoutSamples = map[string]string{
	// Copilot CLI fires postToolUseFailure when the tool itself fails, not
	// when a command it ran exits non-zero: a failing build arrives as
	// postToolUse with "exit code 1" in the result text. Provoking a real
	// tool failure needs something other than a broken build.
	"copilot-cli/postToolUseFailure": "a failing command is not a failing tool; needs a tool error to provoke",
}

// TestEveryBuiltInMappingParses: a mapping that ships broken is a feature
// that silently does nothing.
func TestEveryBuiltInMappingParses(t *testing.T) {
	for _, name := range BuiltIn() {
		m, err := Built(name)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if m.Agent != name {
			t.Errorf("%s is named %q inside", name, m.Agent)
		}
		if m.Response.ContextPath == "" {
			t.Errorf("%s says nothing about where context goes", name)
		}
	}
}

// TestVerifiedMappingsHaveACorpus: `verified_against` is a claim, and a
// claim needs payloads behind it. A mapping that names a version with no
// captured payloads fails here; one that claims nothing is free to.
func TestVerifiedMappingsHaveACorpus(t *testing.T) {
	verified := 0
	for _, name := range BuiltIn() {
		m, err := Built(name)
		if err != nil {
			t.Fatal(err)
		}
		if m.VerifiedAgainst == "" {
			t.Logf("%-14s not verified against any captured payload yet", name)
			continue
		}
		verified++
		if _, err := os.Stat(filepath.Join("../../testdata/agents", name, m.VerifiedAgainst)); err != nil {
			t.Errorf("%s claims %q but there are no payloads for it: %v", name, m.VerifiedAgainst, err)
		}
	}
	if verified == 0 {
		t.Error("no mapping is verified against anything; the corpus has stopped being used")
	}
}

// TestContractAgainstCapturedPayloads replays what a real client sent
// against the mapping that ships for it.
func TestContractAgainstCapturedPayloads(t *testing.T) {
	for _, name := range BuiltIn() {
		mapping, err := Built(name)
		if err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join("../../testdata/agents", name)
		versions, err := os.ReadDir(dir)
		if err != nil {
			continue // No payloads for this agent yet; the test above says so.
		}
		for _, version := range versions {
			if version.IsDir() {
				replayAgent(t, mapping, filepath.Join(dir, version.Name()))
			}
		}
	}
}

// replayAgent checks one captured version of one agent.
func replayAgent(t *testing.T, mapping *Mapping, dir string) {
	t.Helper()
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	seen := map[string]bool{}
	total := 0
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}
		total++
		payload, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Fatal(err)
		}
		event := strings.SplitN(strings.TrimSuffix(f.Name(), ".json"), "-", 2)[0]
		seen[event] = true

		// The file is named after the event, which is what the hook entry
		// passes on the command line for an agent whose payload does not
		// carry it.
		e, ok, err := mapping.Translate(payload, event, now)
		if err != nil {
			t.Errorf("%s/%s: %v", dir, f.Name(), err)
			continue
		}
		if !ok {
			t.Errorf("%s/%s: the mapping no longer covers %s", dir, f.Name(), event)
			continue
		}
		if e.Session == "" {
			t.Errorf("%s/%s: the session id did not resolve", dir, f.Name())
		}
		if e.Cwd == "" {
			t.Errorf("%s/%s: the working directory did not resolve", dir, f.Name())
		}

		// A session event must map to exactly one kind. A tool event maps
		// to either command kind, because which one it is depends on the
		// payload and not on the event's name: Copilot CLI reports a
		// command that exited non-zero as a successful postToolUse with
		// the exit code buried in the result text.
		switch event {
		case "SessionStart", "sessionStart":
			if e.Kind != SessionStarted {
				t.Errorf("%s/%s: kind = %q, want %q", dir, f.Name(), e.Kind, SessionStarted)
			}
		case "SessionEnd", "sessionEnd":
			if e.Kind != SessionEnded {
				t.Errorf("%s/%s: kind = %q, want %q", dir, f.Name(), e.Kind, SessionEnded)
			}
		case "PostToolUse", "postToolUse", "PostToolUseFailure", "postToolUseFailure":
			if e.Kind != CommandFailed && e.Kind != CommandSucceeded {
				t.Errorf("%s/%s: kind = %q, want a command event", dir, f.Name(), e.Kind)
			}
			if e.Command == "" {
				t.Errorf("%s/%s: the command did not resolve", dir, f.Name())
			}
			if e.Output == "" {
				t.Errorf("%s/%s: the output did not resolve", dir, f.Name())
			}
		}
	}
	if total == 0 {
		t.Errorf("%s holds no payloads", dir)
		return
	}
	t.Logf("%-40s %d payload(s) check out", dir, total)

	for _, m := range mapping.Events {
		if seen[m.AgentEvent] {
			continue
		}
		if why, listed := withoutSamples[mapping.Agent+"/"+m.AgentEvent]; listed {
			t.Logf("%s/%s no sample yet: %s", mapping.Agent, m.AgentEvent, why)
			continue
		}
		t.Errorf("%s maps %s but has no captured payload for it, and it is not listed in withoutSamples",
			mapping.Agent, m.AgentEvent)
	}
}

// TestCapturedPayloadsCarryNoAccountName: these files are committed, and the
// scrubbing that made them safe has to keep holding.
func TestCapturedPayloadsAreScrubbed(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory to compare against")
	}
	account := filepath.Base(home)
	if account == "" || account == "." || account == "dev" {
		t.Skip("the account name is not distinctive enough to search for")
	}

	err = filepath.WalkDir("../../testdata/agents", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".json" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), account) {
			t.Errorf("%s carries the account name", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestEveryCapturedPayloadHasAScratchpad records what the corpus settles.
//
// An earlier capture had no scratchpad_dir on any event, and the conclusion
// drawn from it — that the fallback was the ordinary path — was wrong: that
// run never reached a model, because the CLI was not logged in. In a real
// session all four events carry it. The fallback still has to work, because
// the documentation says the field is absent when a session has no
// scratchpad, but it is the exception and not the rule.
func TestEveryCapturedPayloadHasAScratchpad(t *testing.T) {
	for _, name := range []string{
		"PostToolUseFailure-1.json", "PostToolUse-1.json",
		"SessionStart-1.json", "SessionEnd-1.json",
	} {
		payload, err := os.ReadFile(filepath.Join(corpusRoot, "2.1.289", name))
		if err != nil {
			t.Fatal(err)
		}
		e, ok, err := mustBuilt(t).Translate(payload, "", now)
		if err != nil || !ok {
			t.Fatalf("%s: ok=%v err=%v", name, ok, err)
		}
		if e.Scratchpad == "" {
			t.Errorf("%s: no scratchpad_dir", name)
			continue
		}
		if filepath.Dir(StatePath(e, "/cache")) != e.Scratchpad {
			t.Errorf("%s: the state file left the scratchpad", name)
		}
	}

	// And without one, it falls back rather than writing to the root.
	bare := Event{Session: "s1"}
	if dir := filepath.Dir(StatePath(bare, "/cache")); dir != filepath.Join("/cache", "sessions") {
		t.Errorf("the fallback went to %q", dir)
	}
}

// TestTheRealFailurePayloadCarriesItsError is the finding that made this
// corpus worth collecting. The documentation shows a tool result in
// tool_response; a failed Bash command has no tool_response at all, and puts
// its output in a top-level error field. The mapping built from the
// documentation alone found nothing, and the injection path would have been
// silent forever.
func TestTheRealFailurePayloadCarriesItsError(t *testing.T) {
	payload, err := os.ReadFile(filepath.Join(corpusRoot, "2.1.289", "PostToolUseFailure-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	if _, present := raw["tool_response"]; present {
		t.Error("the capture now has tool_response; the mapping can be simplified")
	}
	if _, present := raw["error"]; !present {
		t.Fatal("the capture no longer has a top-level error field")
	}

	e, ok, err := mustBuilt(t).Translate(payload, "", now)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Kind != CommandFailed {
		t.Errorf("kind = %q", e.Kind)
	}
	if !strings.Contains(e.Output, "undefined: greet") {
		t.Errorf("the error text did not reach the event: %q", e.Output)
	}
	if e.Command != "go build ./..." {
		t.Errorf("command = %q", e.Command)
	}
}

// TestAnInterruptedCommandIsNotAnError: Ctrl-C is the user changing their
// mind, and nothing about it belongs in memory.
func TestAnInterruptedCommandIsIgnored(t *testing.T) {
	payload := `{"hook_event_name":"PostToolUseFailure","session_id":"s1","cwd":"/w",
	             "tool_name":"Bash","tool_input":{"command":"go test ./..."},
	             "error":"Exit code 130","is_interrupt":true}`
	if _, ok, _ := mustBuilt(t).Translate([]byte(payload), "", now); ok {
		t.Error("an interrupted command produced an event")
	}

	notInterrupted := strings.Replace(payload, "true", "false", 1)
	if _, ok, _ := mustBuilt(t).Translate([]byte(notInterrupted), "", now); !ok {
		t.Error("is_interrupt false suppressed a real failure")
	}
}
