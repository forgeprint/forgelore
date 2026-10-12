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

// stubAgent writes a fake agent CLI and a mapping that points at it, so the
// whole path can be exercised without a model, an account or a network.
func stubAgent(t *testing.T, reply string) (mappingPath string) {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agent")
	// It reads the material on stdin the way the real one does, so a test
	// that forgot to send it would see an empty body here.
	body := "#!/bin/sh\nmaterial=$(cat)\nif [ -z \"$material\" ]; then echo 'no material on stdin' >&2; exit 3; fi\n" +
		"printf '%s\\n' " + shellQuote(reply) + "\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}

	mapping := map[string]any{
		"mapping_version": 2,
		"agent":           "claude-code",
		"common":          map[string]any{"event_name": "hook_event_name", "session": "session_id"},
		"events":          []any{map[string]any{"agent_event": "SessionStart", "kind": "session_started"}},
		"distill": map[string]any{
			"command":  script,
			"args":     []string{"{{instruction}}"},
			"material": "stdin",
		},
		"response": map[string]any{"context_path": "hookSpecificOutput.additionalContext"},
	}
	data, err := json.Marshal(mapping)
	if err != nil {
		t.Fatal(err)
	}
	mappingPath = filepath.Join(dir, "mapping.json")
	if err := os.WriteFile(mappingPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return mappingPath
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// queueWork puts one untitled work candidate in the review queue.
func queueWork(t *testing.T, project, body string) {
	t.Helper()
	path := writeTranscript(t, t.TempDir(), "s.jsonl",
		fmt.Sprintf(`{"type":"assistant","sessionId":"sess-d","timestamp":"2026-10-12T10:00:00Z","message":{"content":[{"type":"text","text":%q}]}}`+"\n", body), 0)
	mustCLI(t, project, "", "recap", "--transcript", path, "--propose")
}

func TestDistillWritesTheOneLineACandidateIsMissing(t *testing.T) {
	project := newProject(t)
	if err := os.WriteFile(filepath.Join(project, ".forgelore", "config.yaml"),
		[]byte("distill.enabled: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	queueWork(t, project, "The index was rebuilt on every open, which is why the tests were slow.")

	mapping := stubAgent(t, `"The index is rebuilt on every open."`)
	out := mustCLI(t, project, "", "distill", "--key", "sess-d", "--mapping", mapping)
	if !strings.Contains(out, "The index is rebuilt on every open") {
		t.Fatalf("no title came back:\n%s", out)
	}
	// Quotes and the trailing period are the model's habits, not the title.
	if strings.Contains(out, `"`) || strings.Contains(out, "open.") {
		t.Errorf("the title kept the model's packaging:\n%s", out)
	}

	// The queue carries it, and a second call has nothing to do.
	if out := mustCLI(t, project, "", "review"); !strings.Contains(out, "The index is rebuilt") {
		t.Errorf("review does not show the title:\n%s", out)
	}
	if out := mustCLI(t, project, "", "distill", "--key", "sess-d", "--mapping", mapping); !strings.Contains(out, "already has a title") {
		t.Errorf("distilling twice did more than nothing:\n%s", out)
	}

	// Accepting it records who wrote the sentence.
	out = mustCLI(t, project, "", "review", "--accept", "sess-d")
	id := strings.Fields(out)[0]
	if shown := mustCLI(t, project, "", "show", id); !strings.Contains(shown, "source: distill") {
		t.Errorf("the record does not say a model wrote the line:\n%s", shown)
	}
}

func TestDistillIsOffUntilAskedFor(t *testing.T) {
	project := newProject(t)
	queueWork(t, project, "Something happened and it is worth a line.")
	mapping := stubAgent(t, "A line")

	if _, _, err := cli(t, project, "", "distill", "--key", "sess-d", "--mapping", mapping); err == nil {
		t.Error("distillation ran with the setting off")
	}
	// And --dry-run works without it, because it runs nothing.
	out := mustCLI(t, project, "", "distill", "--key", "sess-d", "--mapping", mapping, "--dry-run")
	if !strings.Contains(out, "on stdin") || !strings.Contains(out, "Something happened") {
		t.Errorf("--dry-run did not show what would be sent:\n%s", out)
	}
}

func TestDistillRefusesWhatIsNotATitle(t *testing.T) {
	for _, tc := range []struct{ name, reply, want string }{
		{"a refusal", "SKIP", ""},
		{"an essay", strings.Repeat("long ", 40), ""},
		{"three lines", "The first line\nand more\nand more", "The first line"},
		{"a secret", "The key ghp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa works", "redacted"},
	} {
		got := oneTitle(tc.reply)
		if tc.want == "" && got != "" {
			t.Errorf("%s: title = %q, want nothing", tc.name, got)
		}
		if tc.want != "" && !strings.Contains(got, tc.want) {
			t.Errorf("%s: title = %q, want %q in it", tc.name, got, tc.want)
		}
	}
}

func TestDistillRecordsWhatItSpent(t *testing.T) {
	project := newProject(t)
	if err := os.WriteFile(filepath.Join(project, ".forgelore", "config.yaml"),
		[]byte("distill.enabled: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	queueWork(t, project, "A session that learned one thing worth keeping.")
	mustCLI(t, project, "", "distill", "--key", "sess-d", "--mapping", stubAgent(t, "One thing"))

	ledger := filepath.Join(project, ".forgelore", "ledger",
		fmt.Sprintf("injections-%s.jsonl", time.Now().UTC().Format("2006-01")))
	data, err := os.ReadFile(ledger)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"event":"distill"`) {
		t.Errorf("the spend was not recorded:\n%s", data)
	}
}
