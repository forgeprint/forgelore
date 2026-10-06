package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func hookPayload(event, session, command, output string) string {
	p := map[string]any{
		"hook_event_name": event,
		"session_id":      session,
		"cwd":             "/w",
		"tool_name":       "Bash",
		"tool_input":      map[string]any{"command": command},
		"tool_response":   output,
	}
	data, _ := json.Marshal(p)
	return string(data)
}

// hookContext runs a hook and returns the additionalContext it handed back.
func hookContext(t *testing.T, dir, payload string, args ...string) string {
	t.Helper()
	out := mustCLI(t, dir, payload, append([]string{"hook"}, args...)...)
	if strings.TrimSpace(out) == "" {
		return ""
	}
	var r struct {
		HookSpecificOutput struct {
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("hook wrote something that is not a response: %v\n%s", err, out)
	}
	return r.HookSpecificOutput.AdditionalContext
}

// TestHookNeverFails is K7. Every one of these is a real way the world
// breaks, and not one of them may reach the user's agent as an error.
func TestHookNeverFails(t *testing.T) {
	good := newProject(t)
	bare := t.TempDir()

	broken := newProject(t)
	if err := os.WriteFile(filepath.Join(broken, projectDir, "config.yaml"), []byte("a:\n b: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	noRecords := newProject(t)
	if err := os.RemoveAll(filepath.Join(noRecords, projectDir, "records")); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		why     string
		dir     string
		payload string
		args    []string
	}{
		{why: "nothing on stdin", dir: good, payload: ""},
		{why: "not JSON", dir: good, payload: "wat"},
		{why: "no event name", dir: good, payload: `{"session_id":"s"}`},
		{why: "an event nobody mapped", dir: good, payload: `{"hook_event_name":"Notification"}`},
		{why: "no project at all", dir: bare, payload: hookPayload("PostToolUseFailure", "s", "go build", "boom")},
		{why: "a broken config", dir: broken, payload: hookPayload("PostToolUseFailure", "s", "go build", "./a.go:1:1: undefined: x")},
		{why: "the records directory is gone", dir: noRecords, payload: `{"hook_event_name":"SessionStart","session_id":"s"}`},
		{why: "an unknown adapter", dir: good, payload: "{}", args: []string{"--adapter", "nope"}},
		{why: "a missing mapping file", dir: good, payload: "{}", args: []string{"--mapping", "/no/such.json"}},
		{why: "a flag that does not exist", dir: good, payload: "{}", args: []string{"--wat"}},
	}

	for _, c := range cases {
		_, _, err := cli(t, c.dir, c.payload, append([]string{"hook"}, c.args...)...)
		if err != nil {
			t.Errorf("%s: hook returned an error: %v", c.why, err)
		}
	}
}

// TestHookWritesNothingWhenItKnowsNothing: silence is the cheap case and has
// to stay free.
func TestHookIsSilentOnAnUnknownError(t *testing.T) {
	dir := newProject(t)
	payload := hookPayload("PostToolUseFailure", "s1", "go build ./...", "./main.go:5:14: undefined: greet")
	if got := mustCLI(t, dir, payload, "hook"); got != "" {
		t.Errorf("hook spoke with nothing to say: %q", got)
	}
}

func TestHookInjectsAKnownFix(t *testing.T) {
	dir := newProject(t)
	const title = "Import the package that defines greet"
	mustCLI(t, dir, "./main.go:5:14: undefined: greet\n",
		"record", "--type", "fix", "--title", title, "--command", "go build ./...", "--error-file", "-")

	payload := hookPayload("PostToolUseFailure", "s1", "go test ./...", "./other.go:9:2: undefined: greet")
	got := hookContext(t, dir, payload)
	if !strings.Contains(got, title) {
		t.Errorf("the fix was not injected:\n%s", got)
	}
}

// TestAnInferredMissLeavesNoTrace is the answer to something measured
// rather than feared.
//
// A PostToolUse is a command that worked. The mapping calls it a failure
// when the output looks like one, because a build piped through head exits
// zero and that is the only way to see it (ADR-0022). But output that
// prints an error — a log being catted, a test fixture being written —
// looks exactly the same, and over one real session every inferred failure
// was of that kind: thirteen lookups, none of them a failure.
//
// So an inferred failure that matches nothing is not recorded at all. It
// would otherwise inflate "nothing known" with commands that never failed,
// and the report would mislead about the one thing it exists to measure.
func TestAnInferredMissLeavesNoTrace(t *testing.T) {
	dir := newProject(t)

	// A command that succeeded while printing something error-shaped.
	payload := hookPayload("PostToolUse", "s1", "cat build.log",
		"./main.go:5:14: undefined: greet")
	if got := mustCLI(t, dir, payload, "hook"); got != "" {
		t.Errorf("hook spoke: %q", got)
	}
	if l := readLedger(t, dir); len(l.Entries) != 0 {
		t.Errorf("an inferred miss was written to the ledger: %+v", l.Entries)
	}

	// A failure the agent itself reported still records its miss: nothing
	// was known, and that is a real measurement.
	reported := hookPayload("PostToolUseFailure", "s1", "go build ./...",
		"./main.go:5:14: undefined: greet")
	mustCLI(t, dir, reported, "hook")
	if l := readLedger(t, dir); len(l.Entries) != 1 {
		t.Fatalf("a reported miss was not recorded: %+v", l.Entries)
	}
}

// TestAnInferredHitIsRecorded: the saving is only on the guesses that come
// to nothing. When something is known about the error, the injection and
// the ledger line are both worth having.
func TestAnInferredHitIsRecorded(t *testing.T) {
	dir := newProject(t)
	const title = "Import the package that defines greet"
	mustCLI(t, dir, "./main.go:5:14: undefined: greet\n",
		"record", "--type", "fix", "--title", title, "--command", "go build ./...", "--error-file", "-")

	payload := hookPayload("PostToolUse", "s1", "go build ./... 2>&1 | head -40",
		"./main.go:5:14: undefined: greet")
	if got := hookContext(t, dir, payload); !strings.Contains(got, title) {
		t.Errorf("an inferred failure that matched was not injected:\n%s", got)
	}
	l := readLedger(t, dir)
	if len(l.Entries) != 1 || l.Entries[0].Event != "inject" {
		t.Errorf("got %+v", l.Entries)
	}
}

func TestHookRecordsLatency(t *testing.T) {
	dir := newProject(t)
	payload := hookPayload("PostToolUseFailure", "s1", "go build ./...", "./main.go:5:14: undefined: greet")
	mustCLI(t, dir, payload, "hook")

	l := readLedger(t, dir)
	if len(l.Entries) != 1 {
		t.Fatalf("got %d entries", len(l.Entries))
	}
	if l.Entries[0].HookMS < 0 {
		t.Errorf("hook_ms = %d", l.Entries[0].HookMS)
	}
	if out := mustCLI(t, dir, "", "report"); !strings.Contains(out, "hook latency") {
		t.Errorf("report does not show latency:\n%s", out)
	}
}

// TestControlArmGetsNoContextThroughTheHook: the arm has to be invisible on
// this path too, or the trial measures nothing.
func TestControlArmGetsNoContextThroughTheHook(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, "./main.go:5:14: undefined: greet\n",
		"record", "--type", "fix", "--title", "A known fix", "--command", "go build ./...", "--error-file", "-")
	teamConfig(t, dir, "measure.ab.control_percent: 100\n")

	payload := hookPayload("PostToolUseFailure", "s1", "go build ./...", "./main.go:5:14: undefined: greet")
	if got := hookContext(t, dir, payload); got != "" {
		t.Errorf("the control arm was given context:\n%s", got)
	}
	l := readLedger(t, dir)
	if len(l.Entries) != 1 || l.Entries[0].Event != "control" {
		t.Errorf("entries = %+v", l.Entries)
	}
}

func TestSessionStartIndexHoldsOnlyCommandsAndDecisions(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, "", "record", "--type", "command", "--title", "Run the tests with ./scripts/test.sh")
	mustCLI(t, dir, "", "record", "--type", "decision", "--title", "SQLite is pure Go so the binary stays static")
	mustCLI(t, dir, "./main.go:5:14: undefined: greet\n",
		"record", "--type", "fix", "--title", "A fix that must not be in the index",
		"--command", "go build", "--error-file", "-")
	mustCLI(t, dir, "", "record", "--type", "note", "--title", "A note that must not be in the index")

	got := hookContext(t, dir, `{"hook_event_name":"SessionStart","session_id":"s1"}`)
	for _, want := range []string{"Run the tests", "SQLite is pure Go"} {
		if !strings.Contains(got, want) {
			t.Errorf("the index is missing %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"must not be in the index"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("the index carries something it should not:\n%s", got)
		}
	}
}

// TestSessionStartIndexRespectsItsBudget: the index is a fixed cost paid by
// every session, so it is the one thing that must never grow with the store.
func TestSessionStartIndexRespectsItsBudget(t *testing.T) {
	dir := newProject(t)
	for i := 0; i < 60; i++ {
		mustCLI(t, dir, "", "record", "--type", "command",
			"--title", fmt.Sprintf("Command number %d with a reasonably long title attached", i))
	}
	teamConfig(t, dir, "inject.budget_tokens: 50\n")

	got := hookContext(t, dir, `{"hook_event_name":"SessionStart","session_id":"s1"}`)
	if got == "" {
		t.Fatal("the index was empty")
	}
	if estimated := (len(got) + 3) / 4; estimated > 80 {
		t.Errorf("the index is about %d tokens against a budget of 50:\n%s", estimated, got)
	}
}

func TestSessionStartIsSilentOnAnEmptyStore(t *testing.T) {
	dir := newProject(t)
	if got := hookContext(t, dir, `{"hook_event_name":"SessionStart","session_id":"s1"}`); got != "" {
		t.Errorf("an empty store produced an index: %q", got)
	}
}

// TestACommandThatStartsWorkingBecomesACandidate is the capture path: the
// error stopped, which is a reason to ask, not a reason to record.
func TestACommandThatStartsWorkingBecomesACandidate(t *testing.T) {
	dir := newProject(t)
	const command = "go build ./..."

	mustCLI(t, dir, hookPayload("PostToolUseFailure", "s1", command, "./main.go:5:14: undefined: greet"), "hook")
	mustCLI(t, dir, hookPayload("PostToolUse", "s1", command, "ok"), "hook")

	out := mustCLI(t, dir, "", "review")
	if !strings.Contains(out, command) {
		t.Errorf("no candidate was proposed:\n%s", out)
	}

	// Nothing has been recorded yet.
	if stats := mustCLI(t, dir, "", "stats"); !strings.Contains(stats, "0 record(s)") {
		t.Errorf("a candidate became a record on its own:\n%s", stats)
	}
}

func TestCandidateNeedsATitleToBecomeARecord(t *testing.T) {
	dir := newProject(t)
	const command = "go build ./..."
	mustCLI(t, dir, hookPayload("PostToolUseFailure", "s1", command, "./main.go:5:14: undefined: greet"), "hook")
	mustCLI(t, dir, hookPayload("PostToolUse", "s1", command, "ok"), "hook")

	var pending struct {
		Candidates []struct {
			Fingerprint string `json:"fingerprint"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(mustCLI(t, dir, "", "review", "--json")), &pending); err != nil {
		t.Fatal(err)
	}
	if len(pending.Candidates) != 1 {
		t.Fatalf("got %d candidates", len(pending.Candidates))
	}
	sum := pending.Candidates[0].Fingerprint

	if _, _, err := cli(t, dir, "", "review", "--accept", sum); err == nil {
		t.Error("a candidate was recorded without a title")
	}

	mustCLI(t, dir, "", "review", "--accept", sum, "--title", "greet lives in internal/greeter")

	if out := mustCLI(t, dir, "", "review"); !strings.Contains(out, "nothing waiting") {
		t.Errorf("the candidate survived being accepted:\n%s", out)
	}
	// And the memory now works.
	got := hookContext(t, dir, hookPayload("PostToolUseFailure", "s2", "go vet ./...", "./x.go:1:1: undefined: greet"))
	if !strings.Contains(got, "greet lives in internal/greeter") {
		t.Errorf("the accepted candidate does not inject:\n%s", got)
	}
}

func TestCandidateCanBeDropped(t *testing.T) {
	dir := newProject(t)
	const command = "go build ./..."
	mustCLI(t, dir, hookPayload("PostToolUseFailure", "s1", command, "./main.go:5:14: undefined: greet"), "hook")
	mustCLI(t, dir, hookPayload("PostToolUse", "s1", command, "ok"), "hook")

	out := mustCLI(t, dir, "", "review")
	sum := strings.Fields(out)[0]
	mustCLI(t, dir, "", "review", "--drop", sum)
	if out := mustCLI(t, dir, "", "review"); !strings.Contains(out, "nothing waiting") {
		t.Errorf("the candidate was not dropped:\n%s", out)
	}
	if _, _, err := cli(t, dir, "", "review", "--drop", sum); err == nil {
		t.Error("dropping a candidate twice succeeded")
	}
}

// TestADifferentCommandDoesNotCloseTheLoop: only the command that failed can
// propose the fix for its own error.
func TestADifferentCommandDoesNotProposeAFix(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, hookPayload("PostToolUseFailure", "s1", "go build ./...", "./main.go:5:14: undefined: greet"), "hook")
	mustCLI(t, dir, hookPayload("PostToolUse", "s1", "ls", "ok"), "hook")

	if out := mustCLI(t, dir, "", "review"); !strings.Contains(out, "nothing waiting") {
		t.Errorf("an unrelated command proposed a fix:\n%s", out)
	}
}

func TestSessionEndMentionsWaitingCandidates(t *testing.T) {
	dir := newProject(t)
	const command = "go build ./..."
	mustCLI(t, dir, hookPayload("PostToolUseFailure", "s1", command, "./main.go:5:14: undefined: greet"), "hook")
	mustCLI(t, dir, hookPayload("PostToolUse", "s1", command, "ok"), "hook")

	got := hookContext(t, dir, `{"hook_event_name":"SessionEnd","session_id":"s1","reason":"clear"}`)
	if !strings.Contains(got, "forgelore review") {
		t.Errorf("session end did not mention the candidates:\n%s", got)
	}
}

func TestSessionEndIsSilentWithNothingWaiting(t *testing.T) {
	dir := newProject(t)
	if got := hookContext(t, dir, `{"hook_event_name":"SessionEnd","session_id":"s1"}`); got != "" {
		t.Errorf("session end spoke with nothing to say: %q", got)
	}
}

// TestSessionsDoNotSeeEachOther: two runs failing on the same error must not
// close each other's loops.
func TestSessionsDoNotSeeEachOther(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, hookPayload("PostToolUseFailure", "s1", "go build ./...", "./main.go:5:14: undefined: greet"), "hook")
	mustCLI(t, dir, hookPayload("PostToolUse", "s2", "go build ./...", "ok"), "hook")

	if out := mustCLI(t, dir, "", "review"); !strings.Contains(out, "nothing waiting") {
		t.Errorf("one session closed another's loop:\n%s", out)
	}
}

func TestHookAcceptsAMappingFromDisk(t *testing.T) {
	dir := newProject(t)
	mapping := filepath.Join(t.TempDir(), "other.json")
	if err := os.WriteFile(mapping, []byte(`{
	  "mapping_version": 1, "agent": "other",
	  "response":{"context_path":"additionalContext"},
	  "common": {"event_name": "kind", "session": "sid"},
	  "events": [{"agent_event": "boom", "kind": "command_failed",
	              "command": "cmd", "output": "out", "failure": {"always": true}}]
	}`), 0o644); err != nil {
		t.Fatal(err)
	}
	mustCLI(t, dir, "./main.go:5:14: undefined: greet\n",
		"record", "--type", "fix", "--title", "A known fix", "--command", "go build ./...", "--error-file", "-")

	// This mapping asks for context at the top level rather than inside
	// Claude Code's wrapper, which is the point: the reply shape is data.
	payload := `{"kind":"boom","sid":"s1","cmd":"go build ./...","out":"./main.go:5:14: undefined: greet"}`
	out := mustCLI(t, dir, payload, "hook", "--mapping", mapping)

	var got struct {
		AdditionalContext string `json:"additionalContext"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(got.AdditionalContext, "A known fix") {
		t.Errorf("the mapping from disk did not work:\n%s", out)
	}
	if strings.Contains(out, "hookSpecificOutput") {
		t.Errorf("the reply used Claude Code's shape for another agent:\n%s", out)
	}
}
