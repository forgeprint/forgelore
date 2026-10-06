package worklog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// marksFile is where the record of transcripts already read goes. It lives in
// the cache, which is derived and may be deleted at any time: losing it costs
// one transcript being read twice, which costs a duplicate candidate and
// nothing else.
const marksFile = "worklog.json"

// keep is how long a mark is worth holding. Agents delete their own old
// transcripts, so a mark outlives what it refers to; without pruning this
// file would grow for as long as the project exists.
const keep = 30 * 24 * time.Hour

// Marks is the set of transcripts already read.
//
// Paths are stored as they are rather than hashed. They are never used to
// build a filename here — unlike a session id, which arrives from the agent
// and does become part of a path — and a cache a person can read is easier to
// explain than one they cannot.
type Marks struct {
	Read map[string]time.Time `json:"read"`
}

// MarksPath is the marks file inside a cache directory.
func MarksPath(cacheDir string) string { return filepath.Join(cacheDir, marksFile) }

// LoadMarks reads the marks. A missing or unreadable file is an empty set:
// this is a cache, and failing on it would cost more than it saves.
func LoadMarks(path string) *Marks {
	m := &Marks{Read: map[string]time.Time{}}
	data, err := os.ReadFile(path)
	if err != nil {
		return m
	}
	if json.Unmarshal(data, m) != nil || m.Read == nil {
		return &Marks{Read: map[string]time.Time{}}
	}
	return m
}

// Seen says whether a transcript has already been read.
func (m *Marks) Seen(path string) bool {
	_, ok := m.Read[path]
	return ok
}

// Add marks a transcript as read.
func (m *Marks) Add(path string, at time.Time) {
	if m.Read == nil {
		m.Read = map[string]time.Time{}
	}
	m.Read[path] = at
}

// SaveMarks writes the marks, dropping the ones older than keep.
func SaveMarks(path string, m *Marks, now time.Time) error {
	for p, at := range m.Read {
		if now.Sub(at) > keep {
			delete(m.Read, p)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
