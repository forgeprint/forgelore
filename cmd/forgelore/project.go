package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/forgeprint/forgelore/internal/store"
)

// projectDir is the directory a project keeps its memory in. It sits at the
// root of the project, beside .git.
const projectDir = ".forgelore"

// errNoProject is returned when no store can be found. It names the fix,
// because the first time anyone meets this message they have not run init.
var errNoProject = errors.New("no " + projectDir + " directory here or in any parent (run: forgelore init)")

// findProject walks up from dir until it finds a project, the way git finds
// its repository. Working in a subdirectory is the normal case: an agent runs
// its commands wherever the failing build is, not at the project root.
func findProject(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, projectDir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errNoProject
		}
		dir = parent
	}
}

// openStore finds the project and returns a store for it. start overrides
// where the search begins.
func openStore(e env, start string) (*store.Store, error) {
	if start == "" {
		start = e.wd
	}
	root, err := findProject(start)
	if err != nil {
		return nil, err
	}
	return store.New(root), nil
}

// openIndex opens the store's index and brings it up to date.
//
// Every command that reads does this first. The resync costs nothing when
// nothing has changed — a measured 28 to 56 ms over ten thousand records —
// and the alternative is a tool that answers from a stale index and is
// trusted less for it.
func openIndex(s *store.Store) (*store.Index, error) {
	idx, err := s.OpenIndex()
	if err != nil {
		return nil, err
	}
	if _, err := idx.Sync(); err != nil {
		idx.Close()
		return nil, fmt.Errorf("bringing the index up to date: %w", err)
	}
	return idx, nil
}
