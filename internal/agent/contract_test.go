package agent

import (
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
// Capturing a tool event needs a logged-in CLI and a model turn, and the
// session events below came from a run that got neither. The gap is listed
// rather than ignored because the two events missing from it are the ones
// the whole injection path depends on, and the mapping's treatment of them
// is still an inference from documentation (ADR-0020).
var withoutSamples = map[string]string{
	"PostToolUseFailure": "needs a logged-in CLI and a failing command in a real turn",
	"PostToolUse":        "needs a logged-in CLI and a successful command in a real turn",
}

// TestContractAgainstCapturedPayloads translates every captured payload with
// the mapping that ships, and fails if the mapping has drifted from what the
// agent actually sends.
func TestContractAgainstCapturedPayloads(t *testing.T) {
	versions, err := os.ReadDir(corpusRoot)
	if err != nil {
		t.Fatalf("no captured payloads at all: %v", err)
	}

	mapping := mustBuilt(t)
	seen := map[string]bool{}
	total := 0

	for _, version := range versions {
		if !version.IsDir() {
			continue
		}
		dir := filepath.Join(corpusRoot, version.Name())
		files, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			if filepath.Ext(f.Name()) != ".json" {
				continue
			}
			total++
			payload, err := os.ReadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				t.Fatal(err)
			}

			// The file is named after the event it holds.
			event := strings.SplitN(strings.TrimSuffix(f.Name(), ".json"), "-", 2)[0]
			seen[event] = true

			e, ok, err := mapping.Translate(payload, now)
			if err != nil {
				t.Errorf("%s/%s: %v", version.Name(), f.Name(), err)
				continue
			}
			if !ok {
				t.Errorf("%s/%s: the mapping no longer covers %s", version.Name(), f.Name(), event)
				continue
			}
			if e.Session == "" {
				t.Errorf("%s/%s: the session id did not resolve", version.Name(), f.Name())
			}
			if e.Cwd == "" {
				t.Errorf("%s/%s: the working directory did not resolve", version.Name(), f.Name())
			}

			want := map[string]Kind{
				"SessionStart":       SessionStarted,
				"SessionEnd":         SessionEnded,
				"PostToolUseFailure": CommandFailed,
				"PostToolUse":        CommandSucceeded,
			}[event]
			if want != "" && e.Kind != want {
				t.Errorf("%s/%s: kind = %q, want %q", version.Name(), f.Name(), e.Kind, want)
			}
		}
	}

	if total == 0 {
		t.Fatal("the corpus directory exists but holds no payloads")
	}
	t.Logf("%d captured payload(s) check out", total)

	// Every event the mapping claims to handle either has a sample or is
	// listed as not having one. A new mapped event with neither fails here.
	for _, m := range mapping.Events {
		if seen[m.AgentEvent] {
			continue
		}
		why, listed := withoutSamples[m.AgentEvent]
		if !listed {
			t.Errorf("%s is mapped but has no captured payload and is not listed in withoutSamples", m.AgentEvent)
			continue
		}
		t.Logf("%-20s no sample yet: %s", m.AgentEvent, why)
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

	err = filepath.WalkDir(corpusRoot, func(path string, d os.DirEntry, err error) error {
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

// TestScratchpadIsOftenAbsent records something the documentation only hints
// at and the captured payloads settle: scratchpad_dir is not there. The
// fallback to the project's own cache is the ordinary path, not the edge
// case, and deleting it would break session state for everyone.
func TestScratchpadIsAbsentFromCapturedPayloads(t *testing.T) {
	payload, err := os.ReadFile(filepath.Join(corpusRoot, "2.1.289", "SessionStart-1.json"))
	if err != nil {
		t.Skip("the 2.1.289 session start sample is gone")
	}
	e, ok, err := mustBuilt(t).Translate(payload, now)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if e.Scratchpad != "" {
		t.Skipf("this capture has a scratchpad (%q); the fallback matters less than it did", e.Scratchpad)
	}
	if dir := filepath.Dir(StatePath(e, "/cache")); dir != filepath.Join("/cache", "sessions") {
		t.Errorf("without a scratchpad the state went to %q", dir)
	}
}
