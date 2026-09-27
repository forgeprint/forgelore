package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/forgeprint/forgelore/internal/record"

	_ "modernc.org/sqlite" // pure Go, no cgo (ADR-0002)
)

// The derived index. Nothing authoritative lives here: it can be deleted at
// any moment and rebuilt from the files, which is why a corrupt index is a
// silent rebuild rather than an error the user has to handle (ADR-0009).

const (
	cacheDir  = "cache"
	indexFile = "index.db"

	// indexSchemaVersion is bumped whenever the tables change. A mismatch
	// throws the file away; there is nothing to migrate.
	indexSchemaVersion = "1"
)

const schemaSQL = `
create table if not exists meta (
	key   text primary key,
	value text not null
);
create table if not exists records (
	rowid         integer primary key,
	id            text not null unique,
	scope         text not null,
	type          text not null,
	title         text not null,
	fingerprint   text not null default '',
	created       text not null,
	source        text not null,
	tainted       integer not null default 0,
	superseded_by text not null default '',
	tags          text not null default '',
	path          text not null,
	mtime         integer not null,
	size          integer not null,
	hash          text not null
);
create index if not exists records_by_fingerprint on records(fingerprint);
create virtual table if not exists records_fts using fts5(title, body, tags);
`

// Index is a read-and-write handle on the derived SQLite index.
type Index struct {
	db    *sql.DB
	store *Store
	path  string
}

// Hit is a search result. It carries only what the caller needs to decide
// whether to ask for more: staged access is enforced by what this struct does
// not contain (ADR-0006).
type Hit struct {
	ID         string
	Type       string
	Title      string
	Scope      record.Scope
	Superseded bool
}

// SyncStats reports what a sync did.
type SyncStats struct {
	Scanned   int
	Added     int
	Updated   int
	Unchanged int
	Removed   int
	Problems  []Problem
	Duration  time.Duration
}

// IndexPath returns where the index file lives.
func (s *Store) IndexPath() string {
	return filepath.Join(s.root, cacheDir, indexFile)
}

// OpenIndex opens the index, creating it if it is missing and replacing it if
// it cannot be read. A caller never has to think about a damaged index.
func (s *Store) OpenIndex() (*Index, error) {
	path := s.IndexPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("store: creating %s: %w", filepath.Dir(path), err)
	}

	idx, err := openIndexAt(s, path)
	if err == nil {
		return idx, nil
	}

	// Derived data: throw it away and start again.
	for _, suffix := range []string{"", "-wal", "-shm"} {
		os.Remove(path + suffix)
	}
	idx, err = openIndexAt(s, path)
	if err != nil {
		return nil, fmt.Errorf("store: opening the index: %w", err)
	}
	return idx, nil
}

func openIndexAt(s *Store, path string) (*Index, error) {
	dsn := "file:" + filepath.ToSlash(path) + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, err
	}

	var version string
	switch err := db.QueryRow(`select value from meta where key = 'schema_version'`).Scan(&version); {
	case err == sql.ErrNoRows:
		if _, err := db.Exec(`insert into meta(key, value) values('schema_version', ?)`, indexSchemaVersion); err != nil {
			db.Close()
			return nil, err
		}
	case err != nil:
		db.Close()
		return nil, err
	case version != indexSchemaVersion:
		db.Close()
		return nil, fmt.Errorf("index schema is version %s, this build writes %s", version, indexSchemaVersion)
	}

	return &Index{db: db, store: s, path: path}, nil
}

// Close releases the index.
func (i *Index) Close() error { return i.db.Close() }

// Rebuild discards the index contents and indexes every record again.
func (i *Index) Rebuild() (SyncStats, error) {
	if _, err := i.db.Exec(`delete from records; delete from records_fts;`); err != nil {
		return SyncStats{}, fmt.Errorf("store: clearing the index: %w", err)
	}
	return i.Sync()
}

// indexed is what the index already knows about one file, used to decide
// whether it has to be read again.
type indexed struct {
	rowid int64
	mtime int64
	size  int64
	hash  string
}

// Sync brings the index up to date with the files. A file whose modification
// time and size are unchanged is not read at all; one that was touched without
// changing is recognised by its hash and not re-parsed.
func (i *Index) Sync() (SyncStats, error) {
	start := time.Now()
	stats := SyncStats{}

	known, err := i.known()
	if err != nil {
		return stats, err
	}

	tx, err := i.db.Begin()
	if err != nil {
		return stats, err
	}
	defer tx.Rollback()

	seen := make(map[string]bool, len(known))
	for _, scope := range []record.Scope{record.ScopeTeam, record.ScopeLocal} {
		dir, err := i.store.Dir(scope)
		if err != nil {
			return stats, err
		}
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return stats, fmt.Errorf("store: reading %s: %w", dir, err)
		}

		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ext) {
				continue
			}
			path := filepath.Join(dir, e.Name())
			id := strings.TrimSuffix(e.Name(), ext)
			seen[id] = true
			stats.Scanned++

			info, err := e.Info()
			if err != nil {
				stats.Problems = append(stats.Problems, Problem{Path: path, Err: err, Skipped: true})
				continue
			}
			prev, isKnown := known[id]
			if isKnown && prev.mtime == info.ModTime().UnixNano() && prev.size == info.Size() {
				stats.Unchanged++
				continue
			}

			data, err := os.ReadFile(path)
			if err != nil {
				stats.Problems = append(stats.Problems, Problem{Path: path, Err: err, Skipped: true})
				continue
			}
			sum := sha256.Sum256(data)
			hash := hex.EncodeToString(sum[:])

			if isKnown && prev.hash == hash {
				// Touched but not changed: correct the stat fields and move on.
				if _, err := tx.Exec(`update records set mtime = ?, size = ? where rowid = ?`,
					info.ModTime().UnixNano(), info.Size(), prev.rowid); err != nil {
					return stats, err
				}
				stats.Unchanged++
				continue
			}

			rec, problem := i.store.loadBytes(data, path, scope)
			if problem != nil {
				stats.Problems = append(stats.Problems, *problem)
				if problem.Skipped {
					// A record that cannot be read does not exist for this run,
					// so it must not keep serving an older version of itself.
					if isKnown {
						if err := deleteRow(tx, prev.rowid); err != nil {
							return stats, err
						}
						stats.Removed++
						delete(known, id)
					}
					continue
				}
			}

			if isKnown {
				if err := updateRow(tx, prev.rowid, rec, path, info.ModTime().UnixNano(), info.Size(), hash); err != nil {
					return stats, err
				}
				stats.Updated++
				continue
			}
			if err := insertRow(tx, rec, path, info.ModTime().UnixNano(), info.Size(), hash); err != nil {
				return stats, err
			}
			stats.Added++
		}
	}

	for id, prev := range known {
		if seen[id] {
			continue
		}
		if err := deleteRow(tx, prev.rowid); err != nil {
			return stats, err
		}
		stats.Removed++
	}

	if err := tx.Commit(); err != nil {
		return stats, fmt.Errorf("store: committing the index: %w", err)
	}
	stats.Duration = time.Since(start)
	return stats, nil
}

func (i *Index) known() (map[string]indexed, error) {
	rows, err := i.db.Query(`select rowid, id, mtime, size, hash from records`)
	if err != nil {
		return nil, fmt.Errorf("store: reading the index: %w", err)
	}
	defer rows.Close()

	out := map[string]indexed{}
	for rows.Next() {
		var id string
		var v indexed
		if err := rows.Scan(&v.rowid, &id, &v.mtime, &v.size, &v.hash); err != nil {
			return nil, err
		}
		out[id] = v
	}
	return out, rows.Err()
}

func insertRow(tx *sql.Tx, r *record.Record, path string, mtime, size int64, hash string) error {
	res, err := tx.Exec(`
		insert into records (id, scope, type, title, fingerprint, created, source, tainted, superseded_by, tags, path, mtime, size, hash)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, string(r.Scope), r.Type, r.Title, r.Fingerprint, r.Created.UTC().Format(time.RFC3339),
		r.Source, boolToInt(r.Tainted), r.SupersededBy, strings.Join(r.Tags, " "), path, mtime, size, hash)
	if err != nil {
		return err
	}
	rowid, err := res.LastInsertId()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`insert into records_fts(rowid, title, body, tags) values (?, ?, ?, ?)`,
		rowid, r.Title, r.Body, strings.Join(r.Tags, " "))
	return err
}

func updateRow(tx *sql.Tx, rowid int64, r *record.Record, path string, mtime, size int64, hash string) error {
	_, err := tx.Exec(`
		update records set scope = ?, type = ?, title = ?, fingerprint = ?, created = ?, source = ?,
			tainted = ?, superseded_by = ?, tags = ?, path = ?, mtime = ?, size = ?, hash = ?
		where rowid = ?`,
		string(r.Scope), r.Type, r.Title, r.Fingerprint, r.Created.UTC().Format(time.RFC3339), r.Source,
		boolToInt(r.Tainted), r.SupersededBy, strings.Join(r.Tags, " "), path, mtime, size, hash, rowid)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`update records_fts set title = ?, body = ?, tags = ? where rowid = ?`,
		r.Title, r.Body, strings.Join(r.Tags, " "), rowid)
	return err
}

func deleteRow(tx *sql.Tx, rowid int64) error {
	if _, err := tx.Exec(`delete from records_fts where rowid = ?`, rowid); err != nil {
		return err
	}
	_, err := tx.Exec(`delete from records where rowid = ?`, rowid)
	return err
}

// Search returns the records matching a full-text query, best first.
//
// The query is treated as a set of terms that all have to appear, not as FTS5
// syntax. A search box that can raise a syntax error is a search box people
// stop trusting.
func (i *Index) Search(query string, limit int) ([]Hit, error) {
	match := ftsQuery(query)
	if match == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	rows, err := i.db.Query(`
		select r.id, r.type, r.title, r.scope, r.superseded_by
		from records_fts f
		join records r on r.rowid = f.rowid
		where records_fts match ?
		order by bm25(records_fts), r.id desc
		limit ?`, match, limit)
	if err != nil {
		return nil, fmt.Errorf("store: searching: %w", err)
	}
	defer rows.Close()
	return scanHits(rows)
}

// ByFingerprint returns the records recorded against an error fingerprint,
// newest first. Superseded records are left out: they are never injected, and
// this is the injection path.
func (i *Index) ByFingerprint(fingerprint string) ([]Hit, error) {
	if fingerprint == "" {
		return nil, nil
	}
	rows, err := i.db.Query(`
		select id, type, title, scope, superseded_by
		from records
		where fingerprint = ? and superseded_by = ''
		order by id desc`, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("store: looking up a fingerprint: %w", err)
	}
	defer rows.Close()
	return scanHits(rows)
}

// Count returns how many records the index holds.
func (i *Index) Count() (int, error) {
	var n int
	err := i.db.QueryRow(`select count(*) from records`).Scan(&n)
	return n, err
}

func scanHits(rows *sql.Rows) ([]Hit, error) {
	var out []Hit
	for rows.Next() {
		var h Hit
		var scope, superseded string
		if err := rows.Scan(&h.ID, &h.Type, &h.Title, &scope, &superseded); err != nil {
			return nil, err
		}
		h.Scope = record.Scope(scope)
		h.Superseded = superseded != ""
		out = append(out, h)
	}
	return out, rows.Err()
}

// ftsQuery turns free text into an FTS5 expression that cannot be a syntax
// error: every term becomes a quoted phrase, and the phrases are ANDed.
func ftsQuery(query string) string {
	var terms []string
	for _, field := range strings.Fields(query) {
		cleaned := strings.Trim(field, `"`)
		cleaned = strings.ReplaceAll(cleaned, `"`, `""`)
		if cleaned == "" {
			continue
		}
		terms = append(terms, `"`+cleaned+`"`)
	}
	return strings.Join(terms, " AND ")
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
