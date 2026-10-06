package worklog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// corpus is the transcript captured from a real session, which is the only
// statement anyone has of what the format actually looks like. Its content was
// removed outside the Messages API shaped entries, which is exactly the part
// this package refuses to read, so it tests the real thing.
const corpus = "../../testdata/agents/claude-code/2.1.291/transcript.jsonl"

func TestTheCapturedTranscript(t *testing.T) {
	s, err := ReadFile(corpus, 0)
	if err != nil {
		t.Fatal(err)
	}
	if s.Agent != "2.1.291" {
		t.Errorf("agent version = %q, want 2.1.291", s.Agent)
	}
	if s.ID == "" {
		t.Error("the session has no id")
	}
	if s.Truncated {
		t.Error("a 20 KB transcript was truncated")
	}
	if len(s.Commands) != 2 {
		t.Fatalf("commands = %d, want 2: %+v", len(s.Commands), s.Commands)
	}
	// The capture runs one command that cannot work and one that cannot
	// fail, which is the whole point of the fixture it runs in.
	build, version := s.Commands[0], s.Commands[1]
	if !strings.Contains(build.Text, "go build") || !build.Failed {
		t.Errorf("first command = %+v, want a failing go build", build)
	}
	if !strings.Contains(version.Text, "go version") || version.Failed {
		t.Errorf("second command = %+v, want a passing go version", version)
	}
	if len(s.Fixed()) != 0 {
		t.Errorf("fixed = %v, want none: nothing was fixed in that session", s.Fixed())
	}
	if s.Start.IsZero() || s.End.Before(s.Start) {
		t.Errorf("times = %v..%v", s.Start, s.End)
	}
}

// line builds one transcript entry. Written out rather than marshalled from a
// struct, because what is being tested is the reading of a format this
// project does not control.
func line(parts ...string) string { return "{" + strings.Join(parts, ",") + "}\n" }

func TestFactsOutOfEntries(t *testing.T) {
	in := line(`"type":"assistant"`, `"sessionId":"s1"`, `"version":"9.9.9"`,
		`"cwd":"/w"`, `"gitBranch":"main"`, `"timestamp":"2026-10-06T10:00:00Z"`,
		`"message":{"content":[{"type":"tool_use","id":"a","name":"Bash","input":{"command":"go test ./..."}}]}`) +
		line(`"type":"user"`, `"timestamp":"2026-10-06T10:00:01Z"`,
			`"message":{"content":[{"type":"tool_result","tool_use_id":"a","is_error":true}]}`) +
		// An edit, twice, which is one file.
		line(`"type":"assistant"`, `"timestamp":"2026-10-06T10:01:00Z"`,
			`"message":{"content":[{"type":"tool_use","id":"b","name":"Edit","input":{"file_path":"/w/x.go"}}]}`) +
		line(`"type":"assistant"`, `"timestamp":"2026-10-06T10:01:30Z"`,
			`"message":{"content":[{"type":"tool_use","id":"c","name":"Write","input":{"file_path":"/w/x.go"}}]}`) +
		// The same command again, passing this time.
		line(`"type":"assistant"`, `"timestamp":"2026-10-06T10:02:00Z"`,
			`"message":{"content":[{"type":"tool_use","id":"d","name":"Bash","input":{"command":"go test ./..."}}]}`) +
		line(`"type":"user"`, `"timestamp":"2026-10-06T10:02:05Z"`,
			`"message":{"content":[{"type":"tool_result","tool_use_id":"d","is_error":false}]}`) +
		// A plain prompt: content is a string, not blocks.
		line(`"type":"user"`, `"timestamp":"2026-10-06T10:03:00Z"`, `"message":{"content":"thanks"}`) +
		// The agent's own bookkeeping, which is not read at all.
		line(`"type":"attachment"`, `"attachment":{"type":"skill_listing","names":["private-skill"]}`) +
		line(`"type":"file-history-snapshot"`, `"snapshot":{"files":{"/w/secret.go":"content"}}`)

	s, err := Parse(strings.NewReader(in), false)
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != "s1" || s.Agent != "9.9.9" || s.Cwd != "/w" || s.Branch != "main" {
		t.Errorf("identity = %+v", s)
	}
	if want := []string{"/w/x.go"}; len(s.Edits) != 1 || s.Edits[0] != want[0] {
		t.Errorf("edits = %v, want %v", s.Edits, want)
	}
	if len(s.Commands) != 2 {
		t.Fatalf("commands = %+v, want 2", s.Commands)
	}
	if !s.Commands[0].Failed || s.Commands[1].Failed {
		t.Errorf("results = %v, %v; want failed then passed",
			s.Commands[0].Failed, s.Commands[1].Failed)
	}
	if fixed := s.Fixed(); len(fixed) != 1 || fixed[0] != "go test ./..." {
		t.Errorf("fixed = %v, want the command that started working", fixed)
	}
	if s.End.Sub(s.Start) != 3*time.Minute {
		t.Errorf("span = %v, want 3m", s.End.Sub(s.Start))
	}
}

func TestNothingIsReadOutsideTheApiShapedLayer(t *testing.T) {
	// Everything here is content, and none of it is a fact about the work.
	in := line(`"type":"attachment"`, `"attachment":{"type":"prompt_snapshot","content":"the system prompt"}`) +
		line(`"type":"cost-state"`, `"snapshot":{"totalCostUsd":1.23}`) +
		line(`"type":"atis-latch"`, `"atis":{"names":["a","b"]}`)

	s, err := Parse(strings.NewReader(in), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Edits) != 0 || len(s.Commands) != 0 || s.Unreadable != 0 {
		t.Errorf("read something it should not have: %+v", s)
	}
}

func TestBrokenLinesAreCountedAndSurvived(t *testing.T) {
	in := "not json at all\n" +
		line(`"type":"assistant"`, `"sessionId":"s"`,
			`"message":{"content":[{"type":"tool_use","id":"a","name":"Bash","input":{"command":"ls"}}]}`) +
		"{\"type\":\"user\",\n" + // a line that stops mid-object
		line(`"type":"assistant"`, `"message":{"content":[{"type":"tool_use","id":"b","name":"Bash","input":{"command":"pwd"}}]}`)

	s, err := Parse(strings.NewReader(in), false)
	if err != nil {
		t.Fatal(err)
	}
	if s.Unreadable != 2 {
		t.Errorf("unreadable = %d, want 2", s.Unreadable)
	}
	if len(s.Commands) != 2 {
		t.Errorf("commands = %+v, want both readable ones", s.Commands)
	}
}

func TestAnOverlongLineIsSkippedNotLoaded(t *testing.T) {
	huge := line(`"type":"assistant"`, `"junk":"`+strings.Repeat("x", maxLine+1)+`"`)
	in := huge + line(`"type":"assistant"`, `"message":{"content":[{"type":"tool_use","id":"a","name":"Bash","input":{"command":"ls"}}]}`)

	s, err := Parse(strings.NewReader(in), false)
	if err != nil {
		t.Fatal(err)
	}
	if s.Unreadable != 1 {
		t.Errorf("unreadable = %d, want 1", s.Unreadable)
	}
	if len(s.Commands) != 1 {
		t.Errorf("the line after the long one was lost: %+v", s.Commands)
	}
}

func TestTheTailOfATooLargeTranscript(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.jsonl")

	early := line(`"type":"assistant"`, `"sessionId":"s"`,
		`"message":{"content":[{"type":"tool_use","id":"early","name":"Edit","input":{"file_path":"/w/early.go"}}]}`)
	late := line(`"type":"assistant"`, `"sessionId":"s"`,
		`"message":{"content":[{"type":"tool_use","id":"late","name":"Edit","input":{"file_path":"/w/late.go"}}]}`)
	if err := os.WriteFile(path, []byte(early+strings.Repeat(late, 50)), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := ReadFile(path, int64(len(late)*10))
	if err != nil {
		t.Fatal(err)
	}
	if !s.Truncated {
		t.Error("a file read from its tail did not say so")
	}
	for _, f := range s.Edits {
		if f == "/w/early.go" {
			t.Error("the tail reached the beginning of the file")
		}
	}
	if len(s.Edits) != 1 || s.Edits[0] != "/w/late.go" {
		t.Errorf("edits = %v, want the late one only", s.Edits)
	}
}

func TestAnEmptyTranscript(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.jsonl")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := ReadFile(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Commands) != 0 || len(s.Edits) != 0 || s.Unreadable != 0 {
		t.Errorf("an empty file produced %+v", s)
	}
}

func TestPreviousPicksFinishedSiblings(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	write := func(name string, age time.Duration) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		at := now.Add(-age)
		if err := os.Chtimes(path, at, at); err != nil {
			t.Fatal(err)
		}
		return path
	}

	current := write("current.jsonl", 0)
	old := write("old.jsonl", time.Hour)
	recent := write("recent.jsonl", 10*time.Minute)
	write("still-running.jsonl", 5*time.Second)
	write("notes.txt", time.Hour)

	found, err := Previous(current, now, Settled)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{recent, old}
	if len(found) != len(want) {
		t.Fatalf("found %v, want %v", found, want)
	}
	for i := range want {
		if found[i] != want[i] {
			t.Errorf("found[%d] = %s, want %s (newest first)", i, found[i], want[i])
		}
	}
}

func TestMarksRoundTripAndPrune(t *testing.T) {
	dir := t.TempDir()
	path := MarksPath(dir)
	now := time.Now()

	m := LoadMarks(path)
	if m.Seen("/t/a.jsonl") {
		t.Error("an empty cache claimed to have read something")
	}
	m.Add("/t/a.jsonl", now)
	m.Add("/t/stale.jsonl", now.Add(-keep-time.Hour))
	if err := SaveMarks(path, m, now); err != nil {
		t.Fatal(err)
	}

	again := LoadMarks(path)
	if !again.Seen("/t/a.jsonl") {
		t.Error("a mark did not survive a round trip")
	}
	if again.Seen("/t/stale.jsonl") {
		t.Error("a mark older than the keep window was not pruned")
	}
}

func TestUnreadableMarksAreAnEmptySet(t *testing.T) {
	dir := t.TempDir()
	path := MarksPath(dir)
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if m := LoadMarks(path); len(m.Read) != 0 {
		t.Errorf("a corrupt cache produced %+v", m.Read)
	}
}
