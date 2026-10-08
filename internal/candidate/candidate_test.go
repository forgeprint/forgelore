package candidate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var now = time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)

func TestRoundTrip(t *testing.T) {
	root := t.TempDir()
	want := Candidate{
		Time: now, Session: "s1", Fingerprint: "aaaabbbbccccdddd",
		Command: "go build ./...", Title: "Import the package", Tainted: true,
	}
	if err := Append(root, want); err != nil {
		t.Fatal(err)
	}
	got, err := Read(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != want {
		t.Errorf("got %+v, want [%+v]", got, want)
	}
}

func TestReadingNothingIsNotAnError(t *testing.T) {
	got, err := Read(t.TempDir())
	if err != nil || len(got) != 0 {
		t.Errorf("got %+v, %v", got, err)
	}
}

// TestABadLineDoesNotHideTheRest: a half-written line is what a killed
// process leaves, and it must not cost the candidates around it.
func TestABadLineDoesNotHideTheRest(t *testing.T) {
	root := t.TempDir()
	for _, sum := range []string{"1111111111111111", "2222222222222222"} {
		if err := Append(root, Candidate{Time: now, Fingerprint: sum}); err != nil {
			t.Fatal(err)
		}
	}
	f, err := os.OpenFile(Path(root), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"fingerprint":"333`); err != nil {
		t.Fatal(err)
	}
	f.Close()

	got, err := Read(root)
	if err != nil {
		t.Fatalf("a truncated line stopped the read: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d candidates, want 2", len(got))
	}
}

func TestTake(t *testing.T) {
	cs := []Candidate{
		{Fingerprint: "1111111111111111"},
		{Fingerprint: "2222222222222222"},
	}
	taken, rest, err := Take(cs, "1111111111111111")
	if err != nil {
		t.Fatal(err)
	}
	if taken.Fingerprint != "1111111111111111" || len(rest) != 1 {
		t.Errorf("taken=%+v rest=%+v", taken, rest)
	}
	if _, _, err := Take(cs, "9999999999999999"); err == nil {
		t.Error("taking a candidate that is not there succeeded")
	}
}

// TestWriteRemovesAnEmptyFile: an empty file would make `review` report
// nothing waiting while leaving something behind to puzzle over.
func TestWriteRemovesAnEmptyFile(t *testing.T) {
	root := t.TempDir()
	if err := Append(root, Candidate{Time: now, Fingerprint: "aaaabbbbccccdddd"}); err != nil {
		t.Fatal(err)
	}
	if err := Write(root, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(Path(root)); !os.IsNotExist(err) {
		t.Errorf("the file survived: %v", err)
	}
	if err := Write(root, nil); err != nil {
		t.Errorf("writing nothing twice errored: %v", err)
	}
}

// TestCandidatesLiveInTheLocalScope: a proposal is not a memory and must not
// be committed by accident (ADR-0010).
func TestCandidatesLiveInTheLocalScope(t *testing.T) {
	if dir := filepath.Base(filepath.Dir(Path("/project/.forgelore"))); dir != "local" {
		t.Errorf("candidates are kept in %q", dir)
	}
}

func TestWorkIsKeyedBySessionAndFoldsTogether(t *testing.T) {
	root := t.TempDir()

	if err := Upsert(root, Candidate{Session: "s1", Kind: KindWork, Body: "what the agent said"}); err != nil {
		t.Fatal(err)
	}
	// The second half of one session's memory: the facts, arriving later.
	if err := Upsert(root, Candidate{Session: "s1", Kind: KindWork, Body: "files: a.go"}); err != nil {
		t.Fatal(err)
	}
	// And the same text again, which must not stack up.
	if err := Upsert(root, Candidate{Session: "s1", Kind: KindWork, Body: "files: a.go"}); err != nil {
		t.Fatal(err)
	}
	// A fix from the same session is a different thing, waiting separately.
	if err := Upsert(root, Candidate{Session: "s1", Fingerprint: "abc123", Command: "go build"}); err != nil {
		t.Fatal(err)
	}

	waiting, err := Read(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(waiting) != 2 {
		t.Fatalf("waiting = %d, want one work and one fix: %+v", len(waiting), waiting)
	}

	work, fix := waiting[0], waiting[1]
	if work.Key() != "s1" {
		t.Errorf("work key = %q, want the session id", work.Key())
	}
	if fix.Key() != "abc123" {
		t.Errorf("fix key = %q, want the fingerprint", fix.Key())
	}
	if !strings.Contains(work.Body, "what the agent said") || !strings.Contains(work.Body, "files: a.go") {
		t.Errorf("the two halves did not fold together: %q", work.Body)
	}
	if strings.Count(work.Body, "files: a.go") != 1 {
		t.Errorf("the same text was appended twice: %q", work.Body)
	}

	// Taking one by key leaves the other alone.
	taken, rest, err := Take(waiting, "s1")
	if err != nil {
		t.Fatal(err)
	}
	if taken.Kind != KindWork || len(rest) != 1 || rest[0].Key() != "abc123" {
		t.Errorf("took %+v, left %+v", taken, rest)
	}
}
