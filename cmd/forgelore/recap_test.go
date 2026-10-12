package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeTranscript puts one transcript in a directory of its own, aged so that
// Previous counts it as a finished session's.
func writeTranscript(t *testing.T, dir, name, body string, age time.Duration) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	at := time.Now().Add(-age)
	if err := os.Chtimes(path, at, at); err != nil {
		t.Fatal(err)
	}
	return path
}

const oneEditOneFailure = `{"type":"assistant","sessionId":"abc","version":"2.1.291","gitBranch":"main","timestamp":"2026-10-06T10:00:00Z","message":{"content":[{"type":"tool_use","id":"a","name":"Edit","input":{"file_path":"/w/main.go"}}]}}
{"type":"assistant","timestamp":"2026-10-06T10:00:10Z","message":{"content":[{"type":"tool_use","id":"b","name":"Bash","input":{"command":"go build ./..."}}]}}
{"type":"user","timestamp":"2026-10-06T10:00:12Z","message":{"content":[{"type":"tool_result","tool_use_id":"b","is_error":true}]}}
`

func TestRecapReportsTheFacts(t *testing.T) {
	dir := newProject(t)
	path := writeTranscript(t, t.TempDir(), "s.jsonl", oneEditOneFailure, 0)

	out := mustCLI(t, dir, "", "recap", "--transcript", path)
	for _, want := range []string{"abc", "2.1.291", "main.go", "go build", "failed"} {
		if !strings.Contains(out, want) {
			t.Errorf("recap did not mention %q:\n%s", want, out)
		}
	}
}

func TestRecapJSON(t *testing.T) {
	dir := newProject(t)
	path := writeTranscript(t, t.TempDir(), "s.jsonl", oneEditOneFailure, 0)

	var got struct {
		Transcript string `json:"transcript"`
		Session    struct {
			ID       string `json:"session"`
			Edits    []string
			Commands []struct {
				Text   string `json:"text"`
				Failed bool   `json:"failed"`
			} `json:"commands"`
		} `json:"session"`
	}
	out := mustCLI(t, dir, "", "recap", "--transcript", path, "--json")
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got.Transcript != path || got.Session.ID != "abc" {
		t.Errorf("json = %+v", got)
	}
	if len(got.Session.Commands) != 1 || !got.Session.Commands[0].Failed {
		t.Errorf("commands = %+v", got.Session.Commands)
	}
}

func TestRecapPreviousSkipsTheLiveSessionAndTheReadOnes(t *testing.T) {
	project := newProject(t)
	transcripts := t.TempDir()

	// What a hook would be handed: the session that is running now.
	current := writeTranscript(t, transcripts, "current.jsonl", oneEditOneFailure, 0)
	// One that is still being written, which is not ours to read yet.
	writeTranscript(t, transcripts, "live.jsonl", oneEditOneFailure, time.Second)
	// And the one that finished.
	writeTranscript(t, transcripts, "done.jsonl", oneEditOneFailure, time.Hour)

	out := mustCLI(t, project, "", "recap", "--transcript", current, "--previous")
	if !strings.Contains(out, "go build") {
		t.Fatalf("the finished transcript was not read:\n%s", out)
	}

	// Marking is the only thing this command writes, and only when asked.
	mustCLI(t, project, "", "recap", "--transcript", current, "--previous", "--mark")
	out = mustCLI(t, project, "", "recap", "--transcript", current, "--previous")
	if !strings.Contains(out, "no finished transcript") {
		t.Errorf("a marked transcript was read again:\n%s", out)
	}
	out = mustCLI(t, project, "", "recap", "--transcript", current, "--previous", "--all")
	if !strings.Contains(out, "go build") {
		t.Errorf("--all did not override the mark:\n%s", out)
	}
}

func TestRecapWritesNothingWithoutMark(t *testing.T) {
	project := newProject(t)
	path := writeTranscript(t, t.TempDir(), "s.jsonl", oneEditOneFailure, time.Hour)

	before := treeOf(t, filepath.Join(project, ".forgelore"))
	mustCLI(t, project, "", "recap", "--transcript", path)
	if after := treeOf(t, filepath.Join(project, ".forgelore")); after != before {
		t.Errorf("recap changed the store:\nbefore: %s\nafter:  %s", before, after)
	}
}

// treeOf lists a directory's files, so that a test can say nothing was
// written without knowing what a store holds.
func treeOf(t *testing.T, root string) string {
	t.Helper()
	var paths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(paths, "\n")
}

func TestRecapNeedsATranscript(t *testing.T) {
	dir := newProject(t)
	if _, _, err := cli(t, dir, "", "recap"); err == nil {
		t.Error("recap without --transcript was accepted")
	}
}

// closingAndWork is a session that said something at the end, edited a file
// and watched a command start working: everything work memory keeps.
const closingAndWork = `{"type":"assistant","sessionId":"sess-9","version":"2.1.291","timestamp":"2026-10-06T10:00:00Z","message":{"content":[{"type":"tool_use","id":"a","name":"Bash","input":{"command":"go test ./..."}}]}}
{"type":"user","timestamp":"2026-10-06T10:00:05Z","message":{"content":[{"type":"tool_result","tool_use_id":"a","is_error":true}]}}
{"type":"assistant","timestamp":"2026-10-06T10:01:00Z","message":{"content":[{"type":"tool_use","id":"b","name":"Edit","input":{"file_path":"/w/store.go"}}]}}
{"type":"assistant","timestamp":"2026-10-06T10:02:00Z","message":{"content":[{"type":"tool_use","id":"c","name":"Bash","input":{"command":"go test ./..."}}]}}
{"type":"user","timestamp":"2026-10-06T10:02:05Z","message":{"content":[{"type":"tool_result","tool_use_id":"c","is_error":false}]}}
{"type":"assistant","timestamp":"2026-10-06T10:03:00Z","message":{"content":[{"type":"text","text":"The index was rebuilt on every open, which is why the tests were slow."}]}}
`

func TestProposeMakesOneCandidateAPersonCanAccept(t *testing.T) {
	project := newProject(t)
	path := writeTranscript(t, t.TempDir(), "s.jsonl", closingAndWork, 0)

	mustCLI(t, project, "", "recap", "--transcript", path, "--propose")

	out := mustCLI(t, project, "", "review")
	if !strings.Contains(out, "sess-9") {
		t.Fatalf("the candidate is not keyed by its session:\n%s", out)
	}
	if !strings.Contains(out, "index was rebuilt") {
		t.Errorf("review did not show what the session said:\n%s", out)
	}

	// Proposing the same transcript again is the same candidate, not a
	// second decision about one session.
	mustCLI(t, project, "", "recap", "--transcript", path, "--propose")
	out = mustCLI(t, project, "", "review")
	if n := strings.Count(out, "sess-9"); n != 1 {
		t.Errorf("proposing twice left %d candidates:\n%s", n, out)
	}

	// A title is still the person's to write, and the record it makes is one
	// that is never injected on its own.
	out = mustCLI(t, project, "", "review", "--accept", "sess-9",
		"--title", "The index was rebuilt on every open")
	id := strings.Fields(out)[0]
	shown := mustCLI(t, project, "", "show", id)
	for _, want := range []string{"type: work", "The index was rebuilt", "/w/store.go", "go test ./..."} {
		if !strings.Contains(shown, want) {
			t.Errorf("the record does not carry %q:\n%s", want, shown)
		}
	}
	if strings.Contains(shown, "fingerprint:") {
		t.Errorf("work memory was recorded against a fingerprint:\n%s", shown)
	}
}

func TestProposeRedactsBeforeItWrites(t *testing.T) {
	project := newProject(t)
	leaky := `{"type":"assistant","sessionId":"sess-leak","timestamp":"2026-10-06T10:00:00Z","message":{"content":[{"type":"text","text":"Fixed it. The key was AKIAIOSFODNN7EXAMPLE and it has to go in the env."}]}}` + "\n"
	path := writeTranscript(t, t.TempDir(), "s.jsonl", leaky, 0)

	mustCLI(t, project, "", "recap", "--transcript", path, "--propose")
	out := mustCLI(t, project, "", "review")
	if strings.Contains(out, "AKIAIOSFODNN7EXAMPLE") {
		t.Errorf("a secret reached the review queue:\n%s", out)
	}
}

func TestRecapWithoutProposeQueuesNothing(t *testing.T) {
	project := newProject(t)
	path := writeTranscript(t, t.TempDir(), "s.jsonl", closingAndWork, 0)

	mustCLI(t, project, "", "recap", "--transcript", path)
	if out := mustCLI(t, project, "", "review"); !strings.Contains(out, "nothing waiting") {
		t.Errorf("recap queued something without being asked:\n%s", out)
	}
}

// TestWorkHasItsOwnSessionBudget is why the two budgets are separate: work
// arrives once per session, and a busy fortnight must not push the decisions
// out of the index.
func TestWorkHasItsOwnSessionBudget(t *testing.T) {
	project := newProject(t)

	mustCLI(t, project, "", "record", "--type", "decision", "--scope", "team",
		"--title", "SQLite is the index and markdown is the truth", "--body", "x")
	// Enough work records that one shared budget would crowd the decision
	// out: they sort newest first and every one of these is newer.
	for i := 0; i < 40; i++ {
		mustCLI(t, project, "", "record", "--type", "work", "--scope", "local",
			"--title", fmt.Sprintf("Session %02d went into the store and the index", i),
			"--body", "x")
	}

	got := hookContext(t, project, `{"hook_event_name":"SessionStart","session_id":"s1"}`)
	if !strings.Contains(got, "SQLite is the index") {
		t.Errorf("forty sessions of work pushed the decision out:\n%s", got)
	}
	if !strings.Contains(got, "What recent sessions learned") {
		t.Errorf("the work half is missing:\n%s", got)
	}
	// And the work half is capped: 200 tokens is nowhere near forty titles.
	if n := strings.Count(got, "went into the store"); n == 0 || n > 12 {
		t.Errorf("work titles injected = %d, want a few:\n%s", n, got)
	}
}

// TestSessionStartHarvestsThePreviousSession is the automatic half: nobody ran
// a command, and the session before this one is in the review queue.
func TestSessionStartHarvestsThePreviousSession(t *testing.T) {
	project := newProject(t)
	transcripts := t.TempDir()
	current := writeTranscript(t, transcripts, "current.jsonl", oneEditOneFailure, 0)
	writeTranscript(t, transcripts, "done.jsonl", closingAndWork, time.Hour)

	payload := fmt.Sprintf(`{"hook_event_name":"SessionStart","session_id":"now","transcript_path":%q}`, current)
	mustCLI(t, project, payload, "hook")

	out := mustCLI(t, project, "", "review")
	if !strings.Contains(out, "sess-9") {
		t.Fatalf("session start did not harvest the finished session:\n%s", out)
	}

	// And it is harvested once. A second session start finds it marked.
	mustCLI(t, project, payload, "hook")
	if out := mustCLI(t, project, "", "review"); strings.Count(out, "sess-9") != 1 {
		t.Errorf("the same transcript was harvested twice:\n%s", out)
	}
}

func TestHarvestLeavesALiveSessionAlone(t *testing.T) {
	project := newProject(t)
	transcripts := t.TempDir()
	current := writeTranscript(t, transcripts, "current.jsonl", oneEditOneFailure, 0)
	// Written seconds ago: this is somebody's open session, not ours.
	writeTranscript(t, transcripts, "live.jsonl", closingAndWork, 2*time.Second)

	payload := fmt.Sprintf(`{"hook_event_name":"SessionStart","session_id":"now","transcript_path":%q}`, current)
	mustCLI(t, project, payload, "hook")

	if out := mustCLI(t, project, "", "review"); !strings.Contains(out, "nothing waiting") {
		t.Errorf("a running session was harvested:\n%s", out)
	}
}

func TestHarvestCanBeTurnedOff(t *testing.T) {
	project := newProject(t)
	if err := os.WriteFile(filepath.Join(project, ".forgelore", "config.yaml"),
		[]byte("work.harvest: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	transcripts := t.TempDir()
	current := writeTranscript(t, transcripts, "current.jsonl", oneEditOneFailure, 0)
	writeTranscript(t, transcripts, "done.jsonl", closingAndWork, time.Hour)

	payload := fmt.Sprintf(`{"hook_event_name":"SessionStart","session_id":"now","transcript_path":%q}`, current)
	mustCLI(t, project, payload, "hook")

	if out := mustCLI(t, project, "", "review"); !strings.Contains(out, "nothing waiting") {
		t.Errorf("harvesting happened with the setting off:\n%s", out)
	}
}

// TestHarvestSurvivesEveryWayATranscriptCanBeWrong is K7 for this path. None
// of these may stop the hook, and none may stop the index being handed over.
func TestHarvestSurvivesEveryWayATranscriptCanBeWrong(t *testing.T) {
	project := newProject(t)
	mustCLI(t, project, "", "record", "--type", "decision", "--scope", "team",
		"--title", "The index is derived and the files are the truth", "--body", "x")

	transcripts := t.TempDir()
	writeTranscript(t, transcripts, "broken.jsonl", "{not json\n", time.Hour)
	current := writeTranscript(t, transcripts, "current.jsonl", oneEditOneFailure, 0)

	for _, tc := range []struct{ name, path string }{
		{"a directory that is not there", filepath.Join(t.TempDir(), "gone", "x.jsonl")},
		{"a transcript that is not json", current},
		{"no transcript at all", ""},
	} {
		payload := fmt.Sprintf(`{"hook_event_name":"SessionStart","session_id":"s","transcript_path":%q}`, tc.path)
		got := hookContext(t, project, payload)
		if !strings.Contains(got, "The index is derived") {
			t.Errorf("%s: the session index was lost:\n%s", tc.name, got)
		}
	}
}

// TestTheHarvestRedactsEverythingItCarries is ADR-0013 on the hardest input
// this project has: a transcript holds whole prompts, whole commands and
// whatever the agent wrote back.
func TestTheHarvestRedactsEverythingItCarries(t *testing.T) {
	project := newProject(t)
	transcripts := t.TempDir()

	// A secret in a command that failed and then passed, so it reaches the
	// body through the fix trace, and another in what the agent said.
	leaky := `{"type":"assistant","sessionId":"sess-x","timestamp":"2026-10-12T10:00:00Z","message":{"content":[{"type":"tool_use","id":"a","name":"Bash","input":{"command":"curl -H \"Authorization: Basic dXNlcjpwYXNzd29yZA==\" https://api.example.com"}}]}}
{"type":"user","timestamp":"2026-10-12T10:00:01Z","message":{"content":[{"type":"tool_result","tool_use_id":"a","is_error":true}]}}
{"type":"assistant","timestamp":"2026-10-12T10:01:00Z","message":{"content":[{"type":"tool_use","id":"b","name":"Bash","input":{"command":"curl -H \"Authorization: Basic dXNlcjpwYXNzd29yZA==\" https://api.example.com"}}]}}
{"type":"user","timestamp":"2026-10-12T10:01:01Z","message":{"content":[{"type":"tool_result","tool_use_id":"b","is_error":false}]}}
{"type":"assistant","timestamp":"2026-10-12T10:02:00Z","message":{"content":[{"type":"text","text":"It works now. The deploy key is ghp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa and <private>my laptop password is hunter2</private> so keep it safe."}]}}
`
	current := writeTranscript(t, transcripts, "current.jsonl", oneEditOneFailure, 0)
	writeTranscript(t, transcripts, "leaky.jsonl", leaky, time.Hour)

	payload := fmt.Sprintf(`{"hook_event_name":"SessionStart","session_id":"now","transcript_path":%q}`, current)
	mustCLI(t, project, payload, "hook")

	// Read the queue off disk rather than through review, so that nothing
	// about how it is displayed can hide what was stored.
	queued, err := os.ReadFile(filepath.Join(project, ".forgelore", "local", "candidates.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(queued)
	for _, secret := range []string{
		"dXNlcjpwYXNzd29yZA==",
		"ghp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"hunter2",
		"my laptop password",
	} {
		if strings.Contains(body, secret) {
			t.Errorf("%q reached the candidate queue:\n%s", secret, body)
		}
	}
	if !strings.Contains(body, "It works now") {
		t.Errorf("redaction took the whole account with it:\n%s", body)
	}
}
