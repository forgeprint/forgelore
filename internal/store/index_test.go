package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forgeprint/forgelore/internal/record"
)

func openIndex(t *testing.T, s *Store) *Index {
	t.Helper()
	idx, err := s.OpenIndex()
	if err != nil {
		t.Fatalf("OpenIndex: %v", err)
	}
	t.Cleanup(func() { idx.Close() })
	return idx
}

func TestSyncAddsUpdatesAndRemoves(t *testing.T) {
	s := New(t.TempDir())
	r := newRecord(idA, record.ScopeTeam)
	if _, err := s.Put(r); err != nil {
		t.Fatalf("Put: %v", err)
	}
	idx := openIndex(t, s)

	stats, err := idx.Sync()
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if stats.Added != 1 || stats.Scanned != 1 {
		t.Errorf("first sync = %+v, want one added", stats)
	}

	// Nothing changed: the file is not even read.
	stats, err = idx.Sync()
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if stats.Unchanged != 1 || stats.Added != 0 || stats.Updated != 0 {
		t.Errorf("second sync = %+v, want one unchanged", stats)
	}

	r.Title = "A different title entirely"
	touch(t)
	if _, err := s.Put(r); err != nil {
		t.Fatalf("Put: %v", err)
	}
	stats, err = idx.Sync()
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if stats.Updated != 1 {
		t.Errorf("third sync = %+v, want one updated", stats)
	}

	path, _ := s.Path(record.ScopeTeam, idA)
	if err := os.Remove(path); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	stats, err = idx.Sync()
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if stats.Removed != 1 {
		t.Errorf("fourth sync = %+v, want one removed", stats)
	}
	if n, _ := idx.Count(); n != 0 {
		t.Errorf("index holds %d records, want none", n)
	}
}

func TestSearchMatchesTitleAndBody(t *testing.T) {
	s := New(t.TempDir())
	put(t, s, idA, record.ScopeTeam, "Set CGO_ENABLED=0 before cross-building", "\nthe arm64 linker fails otherwise\n")
	put(t, s, idB, record.ScopeLocal, "Restart the language server after a schema change", "\ngopls caches the old types\n")
	idx := openIndex(t, s)
	if _, err := idx.Sync(); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	hits, err := idx.Search("arm64", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 || hits[0].ID != idA {
		t.Fatalf("search for a body word returned %+v", hits)
	}
	if hits[0].Title == "" || hits[0].Scope != record.ScopeTeam {
		t.Errorf("hit = %+v, want the title and scope filled in", hits[0])
	}

	hits, _ = idx.Search("language server", 10)
	if len(hits) != 1 || hits[0].ID != idB {
		t.Errorf("search for two title words returned %+v", hits)
	}

	hits, _ = idx.Search("nothing matches this", 10)
	if len(hits) != 0 {
		t.Errorf("search with no matches returned %+v", hits)
	}
}

func TestSearchNeverRaisesASyntaxError(t *testing.T) {
	// Whatever a person types goes through as terms, not as FTS5 syntax.
	s := New(t.TempDir())
	put(t, s, idA, record.ScopeTeam, "Set CGO_ENABLED=0 before cross-building", "\nthe arm64 linker fails\n")
	idx := openIndex(t, s)
	if _, err := idx.Sync(); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	for _, query := range []string{
		`"unbalanced`, `AND`, `OR NOT`, `arm64 AND`, `(`, `*`, `NEAR(`, `^`, `-`, `""`, `   `,
	} {
		if _, err := idx.Search(query, 10); err != nil {
			t.Errorf("Search(%q): %v", query, err)
		}
	}
}

func TestByFingerprintSkipsSupersededRecords(t *testing.T) {
	s := New(t.TempDir())

	old := newRecord(idA, record.ScopeTeam)
	old.SupersededBy = idB
	if _, err := s.Put(old); err != nil {
		t.Fatalf("Put: %v", err)
	}
	current := newRecord(idB, record.ScopeTeam)
	if _, err := s.Put(current); err != nil {
		t.Fatalf("Put: %v", err)
	}

	idx := openIndex(t, s)
	if _, err := idx.Sync(); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	hits, err := idx.ByFingerprint("9f2c4a1e7b3d0856")
	if err != nil {
		t.Fatalf("ByFingerprint: %v", err)
	}
	if len(hits) != 1 || hits[0].ID != idB {
		t.Errorf("ByFingerprint returned %+v, want only the current record", hits)
	}

	// It is still findable by search, marked as superseded.
	found, err := idx.Search("cross-building", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	var sawSuperseded bool
	for _, h := range found {
		if h.ID == idA && h.Superseded {
			sawSuperseded = true
		}
	}
	if !sawSuperseded {
		t.Errorf("search returned %+v, want the superseded record marked and present", found)
	}
}

func TestARecordThatBecomesUnreadableLeavesTheIndex(t *testing.T) {
	s := New(t.TempDir())
	if _, err := s.Put(newRecord(idA, record.ScopeTeam)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	idx := openIndex(t, s)
	if _, err := idx.Sync(); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	path, _ := s.Path(record.ScopeTeam, idA)
	touch(t)
	if err := os.WriteFile(path, []byte("---\nbroken\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	stats, err := idx.Sync()
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if stats.Removed != 1 || len(stats.Problems) != 1 {
		t.Errorf("sync = %+v, want the record removed and one problem", stats)
	}
	if n, _ := idx.Count(); n != 0 {
		t.Errorf("index still holds %d records; a record that cannot be read must not keep serving an old version", n)
	}
}

func TestACorruptIndexIsReplacedWithoutComplaint(t *testing.T) {
	s := New(t.TempDir())
	if _, err := s.Put(newRecord(idA, record.ScopeTeam)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.IndexPath()), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(s.IndexPath(), []byte("this is not a database"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	idx, err := s.OpenIndex()
	if err != nil {
		t.Fatalf("OpenIndex on a corrupt file: %v", err)
	}
	defer idx.Close()

	stats, err := idx.Sync()
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if stats.Added != 1 {
		t.Errorf("sync after a rebuild = %+v, want the record indexed again", stats)
	}
}

func TestRebuildProducesTheSameContent(t *testing.T) {
	s := New(t.TempDir())
	put(t, s, idA, record.ScopeTeam, "Set CGO_ENABLED=0 before cross-building", "\nlinker\n")
	put(t, s, idB, record.ScopeLocal, "Restart the language server", "\ngopls\n")
	idx := openIndex(t, s)
	if _, err := idx.Sync(); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	before, _ := idx.Count()

	stats, err := idx.Rebuild()
	if err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	if stats.Added != before {
		t.Errorf("rebuild added %d, want %d", stats.Added, before)
	}
	after, _ := idx.Count()
	if after != before {
		t.Errorf("count after rebuild = %d, want %d", after, before)
	}
}

// TestIndexScale is the phase 1 acceptance measurement: a full index build and
// a search over ten thousand records. It logs the numbers rather than asserting
// a threshold, because a threshold tuned on one machine fails on another; the
// numbers go in the report.
func TestIndexScale(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping the ten thousand record measurement in short mode")
	}
	const n = 10000

	root := t.TempDir()
	s := New(root)
	dir, _ := s.Dir(record.ScopeTeam)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	words := []string{"linker", "cgo", "arm64", "gopls", "migration", "timeout", "docker", "proxy", "certificate", "encoding"}
	writeStart := time.Now()
	for i := 0; i < n; i++ {
		id, err := record.NewULID()
		if err != nil {
			t.Fatalf("NewULID: %v", err)
		}
		r := &record.Record{
			Schema:      record.CurrentSchema,
			ID:          id,
			Type:        record.TypeFix,
			Scope:       record.ScopeTeam,
			Title:       fmt.Sprintf("Fix %d for the %s problem in the build", i, words[i%len(words)]),
			Created:     time.Date(2026, 9, 25, 17, 42, 3, 0, time.UTC),
			Source:      record.SourceHook,
			Fingerprint: fmt.Sprintf("%016x", i%997),
			Tags:        []string{"go", words[i%len(words)]},
			Body: "\n## Symptom\n\nThe " + words[i%len(words)] + " step fails with an error.\n\n" +
				"## Fix\n\nChange the setting and run it again. " + strings.Repeat("Detail. ", 20) + "\n",
		}
		data, err := r.Encode()
		if err != nil {
			t.Fatalf("Encode: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, id+".md"), data, 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	writeElapsed := time.Since(writeStart)

	idx := openIndex(t, s)

	stats, err := idx.Sync()
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if stats.Added != n {
		// At this scale a failure is a pattern, not one bad file, so say
		// enough to see the pattern without printing thousands of lines.
		entries, _ := os.ReadDir(dir)
		t.Logf("files on disk: %d, scanned: %d, problems: %d", len(entries), stats.Scanned, len(stats.Problems))
		for k, p := range stats.Problems {
			if k >= 3 {
				break
			}
			t.Logf("  %s", p)
		}
		t.Fatalf("indexed %d records, want %d", stats.Added, n)
	}

	resync, err := idx.Sync()
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if resync.Unchanged != n {
		t.Errorf("second sync = %+v, want everything unchanged", resync)
	}

	searchStart := time.Now()
	const searches = 100
	var hits int
	for i := 0; i < searches; i++ {
		got, err := idx.Search(words[i%len(words)]+" build", 20)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		hits += len(got)
	}
	searchElapsed := time.Since(searchStart)

	fpStart := time.Now()
	for i := 0; i < searches; i++ {
		if _, err := idx.ByFingerprint(fmt.Sprintf("%016x", i%997)); err != nil {
			t.Fatalf("ByFingerprint: %v", err)
		}
	}
	fpElapsed := time.Since(fpStart)

	info, err := os.Stat(s.IndexPath())
	var size int64
	if err == nil {
		size = info.Size()
	}

	t.Logf("records:          %d", n)
	t.Logf("writing files:    %v", writeElapsed.Round(time.Millisecond))
	t.Logf("full index build: %v", stats.Duration.Round(time.Millisecond))
	t.Logf("no-change resync: %v", resync.Duration.Round(time.Millisecond))
	t.Logf("search:           %v for %d queries (%v each, %d hits)",
		searchElapsed.Round(time.Millisecond), searches,
		(searchElapsed / searches).Round(time.Microsecond), hits)
	t.Logf("fingerprint:      %v for %d lookups (%v each)",
		fpElapsed.Round(time.Millisecond), searches,
		(fpElapsed / searches).Round(time.Microsecond))
	t.Logf("index file:       %.1f MB", float64(size)/(1<<20))
}

func put(t *testing.T, s *Store, id string, scope record.Scope, title, body string) {
	t.Helper()
	r := newRecord(id, scope)
	r.Title = title
	r.Body = body
	if _, err := s.Put(r); err != nil {
		t.Fatalf("Put %s: %v", id, err)
	}
}

// touch waits long enough that a rewritten file gets a different modification
// time on filesystems with a coarse clock.
func touch(t *testing.T) {
	t.Helper()
	time.Sleep(10 * time.Millisecond)
}
