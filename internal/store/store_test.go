package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forgeprint/forgelore/internal/record"
)

const (
	idA = "01K68P7YQZ3M4N5R6S7T8V9W0X"
	idB = "01K68P9AB2C3D4E5F6G7H8J9K0"
	idC = "01K68PB0C1D2E3F4G5H6J7K8M9"
)

func TestPutAndGet(t *testing.T) {
	s := New(t.TempDir())
	want := newRecord(idA, record.ScopeTeam)

	if err := s.Put(want); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := s.Get(idA)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Title != want.Title || got.Scope != want.Scope || got.Type != want.Type {
		t.Errorf("read back %+v, want title %q scope %q", got, want.Title, want.Scope)
	}

	path, _ := s.Path(record.ScopeTeam, idA)
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected the record at %s: %v", path, err)
	}
}

func TestGetIsCaseInsensitive(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Put(newRecord(idA, record.ScopeLocal)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if _, err := s.Get(strings.ToLower(idA)); err != nil {
		t.Errorf("Get with a lowercase id: %v", err)
	}
}

func TestPutIsAtomicAndLeavesNoDebris(t *testing.T) {
	s := New(t.TempDir())
	r := newRecord(idA, record.ScopeTeam)
	if err := s.Put(r); err != nil {
		t.Fatalf("Put: %v", err)
	}
	r.Title = "A second write over the first"
	if err := s.Put(r); err != nil {
		t.Fatalf("Put over an existing record: %v", err)
	}

	dir, _ := s.Dir(record.ScopeTeam)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory holds %v, want only the record", names)
	}

	got, err := s.Get(idA)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Title != "A second write over the first" {
		t.Errorf("Title = %q, want the second write to have replaced the first", got.Title)
	}
}

func TestListIsSortedAndSkipsWhatItCannotRead(t *testing.T) {
	s := New(t.TempDir())
	for _, id := range []string{idC, idA, idB} {
		if err := s.Put(newRecord(id, record.ScopeTeam)); err != nil {
			t.Fatalf("Put %s: %v", id, err)
		}
	}
	dir, _ := s.Dir(record.ScopeTeam)
	// A file mid-edit, and a file that is not a record at all.
	writeFile(t, filepath.Join(dir, "01K68PC000000000000000000X.md"), "---\nnot: frontmatter\n")
	writeFile(t, filepath.Join(dir, "notes.txt"), "ignored")

	records, problems, err := s.List(record.ScopeTeam)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3", len(records))
	}
	if records[0].ID != idA || records[1].ID != idB || records[2].ID != idC {
		t.Errorf("order = %s %s %s, want %s %s %s",
			records[0].ID, records[1].ID, records[2].ID, idA, idB, idC)
	}
	if len(problems) != 1 {
		t.Fatalf("got %d problems, want 1: %v", len(problems), problems)
	}
	if !problems[0].Skipped {
		t.Error("a malformed record must be skipped, not corrected")
	}
}

func TestScopeMismatchIsCorrectedAndReported(t *testing.T) {
	// The directory is authoritative; the field is validated against it.
	s := New(t.TempDir())
	r := newRecord(idA, record.ScopeLocal)
	data, err := r.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	dir, _ := s.Dir(record.ScopeTeam)
	writeFile(t, filepath.Join(dir, idA+".md"), string(data))

	records, problems, err := s.List(record.ScopeTeam)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want the record to survive", len(records))
	}
	if records[0].Scope != record.ScopeTeam {
		t.Errorf("Scope = %q, want the directory to win", records[0].Scope)
	}
	if len(problems) != 1 || problems[0].Skipped {
		t.Fatalf("problems = %v, want one that does not skip the record", problems)
	}
	if !strings.Contains(problems[0].Err.Error(), "directory says") {
		t.Errorf("problem = %q, want it to explain the mismatch", problems[0].Err)
	}
}

func TestFilenameMismatchIsSkipped(t *testing.T) {
	s := New(t.TempDir())
	data, err := newRecord(idA, record.ScopeTeam).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	dir, _ := s.Dir(record.ScopeTeam)
	writeFile(t, filepath.Join(dir, idB+".md"), string(data))

	records, problems, err := s.List(record.ScopeTeam)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("got %d records, want none", len(records))
	}
	if len(problems) != 1 || !problems[0].Skipped {
		t.Fatalf("problems = %v, want one skipped file", problems)
	}
}

func TestPromoteMovesTheFile(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Put(newRecord(idA, record.ScopeLocal)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := s.Promote(idA); err != nil {
		t.Fatalf("Promote: %v", err)
	}

	localPath, _ := s.Path(record.ScopeLocal, idA)
	if _, err := os.Stat(localPath); !os.IsNotExist(err) {
		t.Errorf("the local copy is still there: %v", err)
	}
	got, err := s.Get(idA)
	if err != nil {
		t.Fatalf("Get after Promote: %v", err)
	}
	if got.Scope != record.ScopeTeam {
		t.Errorf("Scope = %q, want team", got.Scope)
	}
}

func TestGetRefusesAnIdThatExistsInBothScopes(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Put(newRecord(idA, record.ScopeTeam)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := s.Put(newRecord(idA, record.ScopeLocal)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	_, err := s.Get(idA)
	if err == nil {
		t.Fatal("Get: got nil error, want a complaint about both scopes")
	}
	if !strings.Contains(err.Error(), "both scopes") {
		t.Errorf("error = %q, want it to name the ambiguity", err)
	}
}

func TestAllReadsBothScopes(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Put(newRecord(idA, record.ScopeTeam)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := s.Put(newRecord(idB, record.ScopeLocal)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	records, problems, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(records) != 2 || len(problems) != 0 {
		t.Errorf("got %d records and %d problems, want 2 and 0", len(records), len(problems))
	}
}

func TestListOfAnAbsentScopeIsEmpty(t *testing.T) {
	// A project that has not run init yet is not an error.
	records, problems, err := New(t.TempDir()).List(record.ScopeTeam)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 0 || len(problems) != 0 {
		t.Errorf("got %d records and %d problems, want none", len(records), len(problems))
	}
}

func TestInitCreatesBothScopes(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	for _, scope := range []record.Scope{record.ScopeTeam, record.ScopeLocal} {
		dir, _ := s.Dir(scope)
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("%s: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func newRecord(id string, scope record.Scope) *record.Record {
	return &record.Record{
		Schema:      record.CurrentSchema,
		ID:          id,
		Type:        record.TypeFix,
		Scope:       scope,
		Title:       "Set CGO_ENABLED=0 before cross-building",
		Created:     time.Date(2026, 9, 25, 17, 42, 3, 0, time.UTC),
		Source:      record.SourceHook,
		Fingerprint: "9f2c4a1e7b3d0856",
		Body:        "\nbody\n",
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
