// Package store reads and writes record files in the two scopes.
//
// The directory decides a record's scope, and a record is one file, so two
// developers adding memory at the same time never touch the same bytes
// (ADR-0009, ADR-0010).
package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/forgeprint/forgelore/internal/record"
)

// Directory names inside a .forgelore directory.
const (
	teamDir  = "records"
	localDir = "local"
)

const ext = ".md"

// Store is rooted at a project's .forgelore directory.
type Store struct {
	root string
}

// New returns a store for the given .forgelore directory.
func New(root string) *Store { return &Store{root: root} }

// Root returns the .forgelore directory this store reads.
func (s *Store) Root() string { return s.root }

// Problem is a file that could not be used as it stands. Skipped says whether
// the record was dropped entirely: a malformed record is skipped rather than
// repaired, because it may be malformed only because someone is editing it
// (ADR-0007). doctor reports these.
type Problem struct {
	Path    string
	Err     error
	Skipped bool
}

func (p Problem) String() string {
	verb := "corrected"
	if p.Skipped {
		verb = "skipped"
	}
	return fmt.Sprintf("%s: %s (%s)", p.Path, p.Err, verb)
}

// Dir returns the directory holding a scope's records.
func (s *Store) Dir(scope record.Scope) (string, error) {
	switch scope {
	case record.ScopeTeam:
		return filepath.Join(s.root, teamDir), nil
	case record.ScopeLocal:
		return filepath.Join(s.root, localDir), nil
	}
	return "", fmt.Errorf("store: unknown scope %q", scope)
}

// Path returns where a record with this id lives in this scope.
func (s *Store) Path(scope record.Scope, id string) (string, error) {
	dir, err := s.Dir(scope)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, id+ext), nil
}

// Init creates the directories for both scopes.
func (s *Store) Init() error {
	for _, scope := range []record.Scope{record.ScopeTeam, record.ScopeLocal} {
		dir, err := s.Dir(scope)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("store: creating %s: %w", dir, err)
		}
	}
	return nil
}

// Put writes a record into its scope. The write is atomic: a reader sees
// either the old file or the new one, never half of either.
func (s *Store) Put(r *record.Record) error {
	data, err := r.Encode()
	if err != nil {
		return fmt.Errorf("store: encoding %s: %w", r.ID, err)
	}
	path, err := s.Path(r.Scope, r.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("store: creating %s: %w", filepath.Dir(path), err)
	}
	return writeAtomic(path, data)
}

// Get returns the record with this id, from whichever scope holds it. Lookups
// are case-insensitive (ADR: docs/record-format.md).
func (s *Store) Get(id string) (*record.Record, error) {
	norm, ok := record.NormalizeULID(id)
	if !ok {
		return nil, fmt.Errorf("store: %q is not a ULID", id)
	}

	var found *record.Record
	var foundPath string
	for _, scope := range []record.Scope{record.ScopeTeam, record.ScopeLocal} {
		path, err := s.Path(scope, norm)
		if err != nil {
			return nil, err
		}
		rec, problem := s.load(path, scope)
		if problem != nil && problem.Skipped {
			if os.IsNotExist(problem.Err) {
				continue
			}
			return nil, fmt.Errorf("store: %s: %w", path, problem.Err)
		}
		if found != nil {
			// The same identifier in both scopes is an anomaly a human has to
			// resolve; picking one silently would hide it.
			return nil, fmt.Errorf("store: %s exists in both scopes (%s and %s)", norm, foundPath, path)
		}
		found, foundPath = rec, path
	}
	if found == nil {
		return nil, fmt.Errorf("store: no record %s", norm)
	}
	return found, nil
}

// List returns one scope's records sorted by identifier, which is also
// chronological, together with whatever could not be read.
func (s *Store) List(scope record.Scope) ([]*record.Record, []Problem, error) {
	dir, err := s.Dir(scope)
	if err != nil {
		return nil, nil, err
	}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("store: reading %s: %w", dir, err)
	}

	var (
		records  []*record.Record
		problems []Problem
	)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ext) {
			continue
		}
		rec, problem := s.load(filepath.Join(dir, e.Name()), scope)
		if problem != nil {
			problems = append(problems, *problem)
		}
		if rec != nil {
			records = append(records, rec)
		}
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	return records, problems, nil
}

// All returns the records in both scopes.
func (s *Store) All() ([]*record.Record, []Problem, error) {
	var (
		records  []*record.Record
		problems []Problem
	)
	for _, scope := range []record.Scope{record.ScopeTeam, record.ScopeLocal} {
		recs, probs, err := s.List(scope)
		if err != nil {
			return nil, nil, err
		}
		records = append(records, recs...)
		problems = append(problems, probs...)
	}
	return records, problems, nil
}

// Promote moves a local record into team scope. It is a file move, not an
// edit, because the directory is what decides scope.
func (s *Store) Promote(id string) error {
	norm, ok := record.NormalizeULID(id)
	if !ok {
		return fmt.Errorf("store: %q is not a ULID", id)
	}
	from, err := s.Path(record.ScopeLocal, norm)
	if err != nil {
		return err
	}
	rec, problem := s.load(from, record.ScopeLocal)
	if problem != nil && problem.Skipped {
		return fmt.Errorf("store: %s: %w", from, problem.Err)
	}
	rec.Scope = record.ScopeTeam
	if err := s.Put(rec); err != nil {
		return err
	}
	if err := os.Remove(from); err != nil {
		return fmt.Errorf("store: removing %s: %w", from, err)
	}
	return nil
}

// load reads one file. A record whose scope field disagrees with its directory
// is corrected rather than skipped: the directory is authoritative, and the
// disagreement is reported.
func (s *Store) load(path string, scope record.Scope) (*record.Record, *Problem) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &Problem{Path: path, Err: err, Skipped: true}
	}
	rec, err := record.Decode(data)
	if err != nil {
		return nil, &Problem{Path: path, Err: err, Skipped: true}
	}
	if want := rec.ID + ext; want != filepath.Base(path) {
		return nil, &Problem{
			Path:    path,
			Err:     fmt.Errorf("id %s does not match the file name", rec.ID),
			Skipped: true,
		}
	}
	if rec.Scope != scope {
		was := rec.Scope
		rec.Scope = scope
		return rec, &Problem{
			Path: path,
			Err:  fmt.Errorf("scope says %s but the directory says %s", was, scope),
		}
	}
	return rec, nil
}

// writeAtomic writes data to a temporary file in the same directory and
// renames it into place, so a reader never sees a partial record.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".forgelore-*.tmp")
	if err != nil {
		return fmt.Errorf("store: creating a temporary file in %s: %w", dir, err)
	}
	tmp := f.Name()
	defer os.Remove(tmp) // harmless once the rename has succeeded

	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("store: writing %s: %w", tmp, err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("store: flushing %s: %w", tmp, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("store: closing %s: %w", tmp, err)
	}
	if err := os.Chmod(tmp, 0o644); err != nil {
		return fmt.Errorf("store: setting permissions on %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("store: renaming into %s: %w", path, err)
	}
	syncDir(dir)
	return nil
}

// syncDir flushes the directory entry so the rename survives a crash. Not
// every platform supports it; where it does not, this does nothing.
func syncDir(dir string) {
	d, err := os.Open(dir)
	if err != nil {
		return
	}
	defer d.Close()
	_ = d.Sync()
}
