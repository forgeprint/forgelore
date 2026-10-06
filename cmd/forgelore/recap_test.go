package main

import (
	"encoding/json"
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
