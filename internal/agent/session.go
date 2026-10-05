package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

// State is what one session has to remember between hook invocations.
//
// Every hook runs as its own process and exits (K8), so "this command already
// failed earlier in this run" cannot live in memory. It lives in a small file
// in the agent's own scratch directory for the session, and in the project's
// cache when the agent offers none.
type State struct {
	// Failed maps an error fingerprint to the command that produced it.
	Failed map[string]string `json:"failed"`
}

// StatePath is where a session's state file goes.
func StatePath(e Event, cacheDir string) string {
	dir := e.Scratchpad
	if dir == "" {
		dir = filepath.Join(cacheDir, "sessions")
	}
	// The session id reaches this from the agent and ends up in a path, so
	// it is hashed rather than trusted: a traversal in it would otherwise
	// choose the file Forgelore writes.
	sum := sha256.Sum256([]byte(e.Session))
	return filepath.Join(dir, "forgelore-"+hex.EncodeToString(sum[:8])+".json")
}

// LoadState reads a session's state. A missing or unreadable file is an empty
// state: losing it costs a candidate, and failing here would cost the hook.
func LoadState(path string) *State {
	s := &State{Failed: map[string]string{}}
	data, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	if json.Unmarshal(data, s) != nil || s.Failed == nil {
		return &State{Failed: map[string]string{}}
	}
	return s
}

// SaveState writes a session's state, creating its directory.
func SaveState(path string, s *State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
