package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/forgeprint/forgelore/internal/store"
)

// projectGitignore keeps the two directories that must not be committed out
// of the user's repository. Writing it is the whole reason init touches
// anything beyond the directories: a user who commits their index has
// committed a derived file that conflicts on every pull, and a user who
// commits their local scope has shared records that were meant to be theirs.
const projectGitignore = `# Derived. Rebuilt from the records whenever it is missing (K9).
cache/

# Personal. Records in this scope are yours and are not shared (K10).
local/

# Measurements. Local, and never leaves the machine (K12).
ledger/
`

// agentsLine is what a project adds to AGENTS.md to put the tool in front of
// its agent. init prints it and does not write it: AGENTS.md belongs to the
// project, and a tool that edits it unasked is a tool people stop installing.
const agentsLine = `When a shell command fails, run
` + "`forgelore recall --command \"<the command>\" --error-file -`" + ` with the
command's output on stdin before trying a fix.`

func cmdInit(e env, args []string) error {
	fs := newFlags(e, "init")
	dir := fs.String("dir", "", "create the store here instead of the current directory")
	withHook := fs.Bool("with-git-hook", false, "also install a pre-commit hook that runs forgelore check")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	base := e.wd
	if *dir != "" {
		base = *dir
	}
	root, err := filepath.Abs(filepath.Join(base, projectDir))
	if err != nil {
		return err
	}

	// What already exists is reported as existing rather than created, so
	// that running init twice tells the truth both times.
	paths := []string{
		filepath.Join(root, "records"),
		filepath.Join(root, "local"),
		filepath.Join(root, ".gitignore"),
	}
	existed := make(map[string]bool, len(paths))
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			existed[p] = true
		}
	}

	s := store.New(root)
	if err := s.Init(); err != nil {
		return err
	}

	ignore := filepath.Join(root, ".gitignore")
	if !existed[ignore] {
		if err := os.WriteFile(ignore, []byte(projectGitignore), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", ignore, err)
		}
	}

	hook := ""
	if *withHook {
		installed, err := installPreCommitHook(base)
		if err != nil {
			return err
		}
		hook = installed
	}

	var created, kept []string
	for _, p := range paths {
		rel := relativeTo(base, p)
		if existed[p] {
			kept = append(kept, rel)
		} else {
			created = append(created, rel)
		}
	}

	if *asJSON {
		return writeJSON(e.stdout, struct {
			Root       string   `json:"root"`
			Created    []string `json:"created"`
			AlreadyThe []string `json:"already_there"`
			AgentsLine string   `json:"agents_md_line"`
			GitHook    string   `json:"git_hook,omitempty"`
		}{root, created, kept, agentsLine, hook})
	}

	for _, p := range created {
		fmt.Fprintf(e.stdout, "created  %s\n", p)
	}
	for _, p := range kept {
		fmt.Fprintf(e.stdout, "kept     %s\n", p)
	}
	fmt.Fprintf(e.stdout, `
Records in %s are meant to be committed. The index and your local scope are
not; the .gitignore just written says so.

To put this in front of your agent, add to AGENTS.md:

%s

forgelore does not write AGENTS.md for you.
`, relativeTo(base, filepath.Join(root, "records")), indent(agentsLine, "    "))

	if hook != "" {
		fmt.Fprintf(e.stdout, "\ncreated  %s\nIt runs `forgelore check` before every commit.\n", relativeTo(base, hook))
	} else {
		fmt.Fprintf(e.stdout, `
A team record must never carry a secret. Check before committing:

    %s

Add it to your pre-commit hook, or let forgelore write one with
--with-git-hook.
`, hookLine)
	}
	return nil
}

// relativeTo prints a path the way the user typed their way to it, falling
// back to the absolute path when the two share no root.
func relativeTo(base, path string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return rel
}
