package worklog

import (
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Settled is how long a transcript has to have gone unwritten before it
// counts as a finished session's.
//
// The documented behaviour is that an agent writes its transcript
// asynchronously, which is why the file belonging to a session that is ending
// is read at the next session's start rather than at its own end (ADR-0026).
// The same asynchrony means a file that was touched a moment ago may belong to
// a session still running — on a machine with two agents open, often does.
// Below a minute that race is real; far above it, two sessions opened back to
// back would skip the first one's work.
const Settled = time.Minute

// Previous returns the transcripts beside current that look finished, newest
// first. The directory comes from the path the agent handed over rather than
// from any knowledge of where an agent keeps its sessions: the one thing every
// agent with a transcript does is name its own.
func Previous(current string, now time.Time, settled time.Duration) ([]string, error) {
	if settled <= 0 {
		settled = Settled
	}
	dir := filepath.Dir(current)
	names, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	type found struct {
		path string
		mod  time.Time
	}
	var out []found
	for _, name := range names {
		if name.IsDir() || filepath.Ext(name.Name()) != ".jsonl" {
			continue
		}
		path := filepath.Join(dir, name.Name())
		if path == current {
			continue
		}
		info, err := name.Info()
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) < settled {
			// Still being written, or close enough that reading it would be
			// reading half a session.
			continue
		}
		out = append(out, found{path, info.ModTime()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].mod.After(out[j].mod) })

	paths := make([]string, len(out))
	for i, f := range out {
		paths[i] = f.path
	}
	return paths, nil
}
