package candidate

import (
	"os"
	"path/filepath"
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
