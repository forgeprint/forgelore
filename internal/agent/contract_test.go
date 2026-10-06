package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/forgeprint/forgelore/internal/fingerprint"
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

// deliberatelyUnmapped is the other direction: an event the corpus holds
// and the mapping ignores on purpose. Without this the replay reads an
// unmapped payload as a mapping that has rotted.
var deliberatelyUnmapped = map[string]string{
	// Cursor fires both for the same command: afterShellExecution carries
	// the command and its output, and postToolUse/postToolUseFailure carry
	// the same thing with a tool name, a cwd and an exit code. Mapping
	// both would look one error up twice — two ledger lines and two
	// injections for one failure. The payloads stay in the corpus because
	// the duplication is the finding.
	"cursor/afterShellExecution": "duplicates postToolUse for the same command",
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

		// The event name is supplied only for an agent whose mapping says
		// its payloads do not carry one. Supplying it unconditionally
		// would mask a broken event_name path: the fallback would answer
		// for a lookup that should have failed, and the mapping would
		// pass this test while failing in front of the agent, which does
		// not pass --event.
		caller := ""
		if len(mapping.Common.EventName) == 0 || mapping.Common.EventName[0] == "" {
			caller = event
		}
		e, ok, err := mapping.Translate(payload, caller, now)
		if err != nil {
			t.Errorf("%s/%s: %v", dir, f.Name(), err)
			continue
		}
		if !ok {
			if why, listed := deliberatelyUnmapped[mapping.Agent+"/"+event]; listed {
				t.Logf("%s/%s not mapped on purpose: %s", mapping.Agent, event, why)
				continue
			}
			t.Errorf("%s/%s: the mapping no longer covers %s", dir, f.Name(), event)
			continue
		}
		if e.Session == "" {
			t.Errorf("%s/%s: the session id did not resolve", dir, f.Name())
		}
		// A working directory is required of a command event and not of a
		// session one. Cursor's session payloads carry only
		// workspace_roots, a list, and Event.Cwd is informational — it is
		// set and never read. Demanding it everywhere would mean putting
		// a JSON array into a field named Cwd to satisfy a test.
		if e.Cwd == "" && (e.Kind == CommandFailed || e.Kind == CommandSucceeded) {
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
		// Cursor puts the signed-in user's address in every payload, and
		// an address is the one thing in a hook payload that identifies a
		// person outright. The scrubbing replaces it with
		// dev@example.com; anything else reaching the corpus is a leak.
		for _, m := range emailish.FindAllString(string(data), -1) {
			if m != "dev@example.com" {
				t.Errorf("%s carries an email address: %s", path, m)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// emailish is deliberately loose. It is looking for something to refuse,
// not something to parse.
var emailish = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)

// TestTheScratchpadCameAndWent records what two corpora settle between
// them, and it has been wrong twice.
//
// First reading: an early capture carried no scratchpad_dir anywhere, and
// the conclusion — that the fallback was the ordinary path — was wrong,
// because that run never reached a model. Second reading: 2.1.289 carried
// it on all four events, so it was called the rule. Then 2.1.290 dropped it
// from every event again.
//
// So the field is optional in practice and not only on paper, and the
// fallback is load-bearing rather than a corner. Nothing breaks either way:
// without it the session state goes to the store's own cache. This test
// holds both shapes so that neither reading can quietly become an
// assumption again.
func TestTheScratchpadCameAndWent(t *testing.T) {
	events := []string{
		"PostToolUseFailure-1.json", "PostToolUse-1.json",
		"SessionStart-1.json", "SessionEnd-1.json",
	}
	for _, tc := range []struct {
		version string
		want    bool
	}{
		{"2.1.289", true},
		{"2.1.290", false},
	} {
		for _, name := range events {
			payload, err := os.ReadFile(filepath.Join(corpusRoot, tc.version, name))
			if err != nil {
				t.Fatal(err)
			}
			e, ok, err := mustBuilt(t).Translate(payload, "", now)
			if err != nil || !ok {
				t.Fatalf("%s/%s: ok=%v err=%v", tc.version, name, ok, err)
			}
			if got := e.Scratchpad != ""; got != tc.want {
				t.Errorf("%s/%s: scratchpad present = %v, want %v", tc.version, name, got, tc.want)
				continue
			}
			dir := filepath.Dir(StatePath(e, "/cache"))
			if tc.want && dir != e.Scratchpad {
				t.Errorf("%s/%s: the state file left the scratchpad", tc.version, name)
			}
			if !tc.want && dir != filepath.Join("/cache", "sessions") {
				t.Errorf("%s/%s: without a scratchpad the state went to %q", tc.version, name, dir)
			}
		}
	}

	// And with no session at all, it still does not write to the root.
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

// TestAPipedBuildStillCountsAsAFailure is the second finding this corpus
// paid for, and it is the opposite shape to the first.
//
// A Bash command's exit status reaches Forgelore only through which event
// fires. When the agent writes `go build ./... 2>&1 | head -40` the pipeline
// exits with head's status, which is zero, so Claude Code sends PostToolUse
// — correctly — and no field anywhere in the payload says the build failed.
// Until output_has_diagnostic the mapping called that a success and the
// error was never looked up. Silently, and indistinguishably from a session
// that met no errors.
//
// The two payloads below are the whole behaviour: one succeeded, one did
// not, and only the output tells them apart.
func TestAPipedBuildStillCountsAsAFailure(t *testing.T) {
	for _, tc := range []struct {
		file string
		want Kind
	}{
		{"PostToolUse-2.json", CommandFailed},
		{"PostToolUse-1.json", CommandSucceeded},
	} {
		payload, err := os.ReadFile(filepath.Join(corpusRoot, "2.1.289", tc.file))
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]any
		if err := json.Unmarshal(payload, &raw); err != nil {
			t.Fatal(err)
		}
		if _, present := raw["error"]; present {
			t.Errorf("%s: the capture now has an error field, so the event decides after all", tc.file)
		}

		e, ok, err := mustBuilt(t).Translate(payload, "", now)
		if err != nil || !ok {
			t.Fatalf("%s: ok=%v err=%v", tc.file, ok, err)
		}
		if e.Kind != tc.want {
			t.Errorf("%s: kind = %q, want %q", tc.file, e.Kind, tc.want)
		}
	}
}

// TestOutputHasDiagnosticNeedsItsFormatVersion: the version number is what
// stops an older binary reading a mapping it would get wrong. A mapping that
// uses the test while claiming version 1 would be read by that binary with
// the test quietly dropped, which is the silent failure this whole file
// exists to prevent.
func TestOutputHasDiagnosticNeedsItsFormatVersion(t *testing.T) {
	const m = `{"mapping_version":1,"agent":"a","response":{"context_path":"c"},
	            "events":[{"agent_event":"X","kind":"command_succeeded",
	                       "failure":{"output_has_diagnostic":true}}]}`
	_, err := ParseMapping([]byte(m))
	if err == nil {
		t.Fatal("a version 1 mapping used output_has_diagnostic and was accepted")
	}
	if !strings.Contains(err.Error(), "output_has_diagnostic") {
		t.Errorf("the error does not say which field is at fault: %v", err)
	}
}

// TestGeminiReportsItsExitCodeInTheOutput is the third agent in a row whose
// documented failure marker was not the one it sends.
//
// The hooks reference says a tool result carries an optional `error`. A
// shell command that exits non-zero sets no such field — the mapping built
// on it would have called every failing build a success. What the payload
// does carry is the exit code, as a line inside `llmContent`, which is the
// same shape Copilot CLI turned out to use.
func TestGeminiReportsItsExitCodeInTheOutput(t *testing.T) {
	m, err := Built("gemini-cli")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		file string
		want Kind
	}{
		{"AfterTool-1.json", CommandFailed},
		{"AfterTool-2.json", CommandSucceeded},
	} {
		payload, err := os.ReadFile(filepath.Join("../../testdata/agents/gemini-cli/0.62.0", tc.file))
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]any
		if err := json.Unmarshal(payload, &raw); err != nil {
			t.Fatal(err)
		}
		if resp, ok := raw["tool_response"].(map[string]any); ok {
			if _, present := resp["error"]; present {
				t.Errorf("%s: the capture now has tool_response.error, so the documented field does fire after all", tc.file)
			}
		}

		e, ok, err := m.Translate(payload, "", now)
		if err != nil || !ok {
			t.Fatalf("%s: ok=%v err=%v", tc.file, ok, err)
		}
		if e.Kind != tc.want {
			t.Errorf("%s: kind = %q, want %q", tc.file, e.Kind, tc.want)
		}
	}
}

// TestGeminiWrappingDoesNotReachTheFingerprint: Gemini hands the output
// back inside an <untrusted_context> wrapper, with an "Output:" prefix, an
// "Exit Code:" line and a process group id that is different on every run.
//
// If any of that reached the fingerprint, the same error would hash
// differently each time and memory would never match anything — silently,
// because every individual lookup still succeeds. So this checks the
// stronger property the whole design rests on: one error, one fingerprint,
// whichever agent reported it.
func TestGeminiWrappingDoesNotReachTheFingerprint(t *testing.T) {
	const command = "go build ./..."
	const wrapped = "<untrusted_context>\nOutput: # example.com/broken/cmd/app\n" +
		"cmd/app/main.go:4:2: undefined: greet\nExit Code: 1\n" +
		"Process Group PGID: %s\n</untrusted_context>"

	// The same error as Claude Code reports it, with no wrapper at all.
	plain := fingerprint.Scan(command, "# example.com/broken/cmd/app\ncmd/app/main.go:4:2: undefined: greet")
	if len(plain) != 1 {
		t.Fatalf("the plain output produced %d events", len(plain))
	}

	for _, pgid := range []string{"72958", "99999"} {
		got := fingerprint.Scan(command, strings.Replace(wrapped, "%s", pgid, 1))
		if len(got) != 1 {
			t.Fatalf("pgid %s produced %d events", pgid, len(got))
		}
		if got[0].Sum != plain[0].Sum {
			t.Errorf("pgid %s: %s, but the same error unwrapped is %s", pgid, got[0].Sum, plain[0].Sum)
		}
	}
}

// TestCursorSendsTheSameCommandTwice is the finding its capture produced,
// and the reason one of its events is deliberately unmapped.
//
// A single failing build fires both afterShellExecution, which carries the
// command and its output, and postToolUseFailure, which carries the same
// failure with a tool name, a working directory and an error message.
// Mapping both would look one error up twice: two ledger lines, two
// injections, one failure.
//
// The exit code is also worth pinning. postToolUse hands the result back as
// a JSON *string* — `{"output":"…","exitCode":0}` — so a command whose
// status was masked can be spotted by reading the code out of the text,
// which is what the mapping does.
func TestCursorSendsTheSameCommandTwice(t *testing.T) {
	const dir = "../../testdata/agents/cursor/2026.10.01"
	m, err := Built("cursor")
	if err != nil {
		t.Fatal(err)
	}

	// The two events describing the same failing `go build`.
	shell, err := os.ReadFile(filepath.Join(dir, "afterShellExecution-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := m.Translate(shell, "", now); ok {
		t.Error("afterShellExecution produced an event; the same failure is already covered by postToolUseFailure")
	}

	failure, err := os.ReadFile(filepath.Join(dir, "postToolUseFailure-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	e, ok, err := m.Translate(failure, "", now)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Kind != CommandFailed {
		t.Errorf("kind = %q", e.Kind)
	}
	if !strings.Contains(e.Output, "undefined: greet") {
		t.Errorf("the error text did not reach the event: %q", e.Output)
	}

	// And the succeeding command stays a success, exit code zero.
	success, err := os.ReadFile(filepath.Join(dir, "postToolUse-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	e, ok, err = m.Translate(success, "", now)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Kind != CommandSucceeded {
		t.Errorf("kind = %q, want %q", e.Kind, CommandSucceeded)
	}
	if !strings.Contains(e.Output, `"exitCode":0`) {
		t.Errorf("the result is no longer a JSON string carrying the code: %q", e.Output)
	}
}

// TestCursorRepliesWhereCursorReads: verified against a running agent on
// 2026-10-06, not taken from the documentation. A hook returned a probe
// token in additional_context and the model repeated it verbatim, which is
// the only way to know a reply path works — Codex's documented one turned
// out to be wrong, and would have been silent.
func TestCursorRepliesWhereCursorReads(t *testing.T) {
	m, err := Built("cursor")
	if err != nil {
		t.Fatal(err)
	}
	reply, err := m.Context("postToolUseFailure", "a hint")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(reply, &got); err != nil {
		t.Fatalf("%v\n%s", err, reply)
	}
	if got["additional_context"] != "a hint" {
		t.Errorf("the context is not where Cursor reads it:\n%s", reply)
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
