package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// localFix records a fix in the local scope and returns its id.
func localFix(t *testing.T, dir, title string, tainted ...bool) string {
	t.Helper()
	args := []string{
		"record", "--type", "fix", "--title", title, "--scope", "local",
		"--command", "go build ./...", "--error-file", "-",
	}
	if len(tainted) > 0 && tainted[0] {
		args = append(args, "--tainted")
	}
	out := mustCLI(t, dir, "./main.go:5:14: undefined: greet\n", args...)
	return strings.Fields(out)[0]
}

func TestPromoteMovesARecordIntoTheTeamScope(t *testing.T) {
	dir := newProject(t)
	id := localFix(t, dir, "Import the package that defines greet")

	mustCLI(t, dir, "", "promote", id)

	if _, err := os.Stat(filepath.Join(dir, projectDir, "records", id+".md")); err != nil {
		t.Errorf("the record is not in the team scope: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, projectDir, "local", id+".md")); !os.IsNotExist(err) {
		t.Error("the record is still in the local scope")
	}
	if out := mustCLI(t, dir, "", "show", id); !strings.Contains(out, "scope: team") {
		t.Errorf("the record still says local:\n%s", out)
	}
}

// TestPromotingTwiceIsRefused: the second one would be a no-op dressed up as
// an action.
func TestPromotingATeamRecordIsRefused(t *testing.T) {
	dir := newProject(t)
	id := localFix(t, dir, "A fix")
	mustCLI(t, dir, "", "promote", id)

	if _, _, err := cli(t, dir, "", "promote", id); err == nil {
		t.Error("promoting a team record succeeded")
	}
}

// TestTaintedNeedsSayingSoTwice is ADR-0013 at the moment it matters:
// promoting shares something nobody on the team wrote.
func TestTaintedPromotionNeedsTheFlag(t *testing.T) {
	dir := newProject(t)
	id := localFix(t, dir, "Something a web page said", true)

	_, _, err := cli(t, dir, "", "promote", id)
	if err == nil {
		t.Fatal("a tainted record was promoted without the flag")
	}
	if !strings.Contains(err.Error(), "--force-tainted") {
		t.Errorf("the refusal does not say how to proceed: %v", err)
	}

	mustCLI(t, dir, "", "promote", id, "--force-tainted")
	if out := mustCLI(t, dir, "", "show", id); !strings.Contains(out, "tainted: true") {
		t.Errorf("the taint was lost on the way:\n%s", out)
	}
}

func TestCheckPassesOnACleanStore(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, "", "record", "--type", "note", "--title", "Nothing secret here")
	if out := mustCLI(t, dir, "", "check"); !strings.Contains(out, "nothing to hide") {
		t.Errorf("check complained about a clean store:\n%s", out)
	}
}

// TestCheckCatchesAHandEditedSecret is the gap check exists for. Redaction
// runs when Forgelore writes a record; a record can also arrive by hand or
// by merge, and neither went through the writer.
func TestCheckCatchesAHandEditedSecret(t *testing.T) {
	dir := newProject(t)
	out := mustCLI(t, dir, "", "record", "--type", "note", "--title", "How to reach staging")
	id := strings.Fields(out)[0]

	path := filepath.Join(dir, projectDir, "records", id+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Written straight to disk, the way a merge or an editor would.
	edited := string(data) + "\nDEPLOY_TOKEN=placeholder-not-a-real-value\n"
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := cli(t, dir, "", "check")
	if err == nil {
		t.Fatal("check passed a record carrying a secret")
	}
	if !strings.Contains(stdout, id) {
		t.Errorf("check did not name the file:\n%s", stdout)
	}
}

// TestCheckIgnoresTheLocalScope: local records never leave the machine, and
// reporting a person's own notes back at them is not this command's job.
func TestCheckIgnoresTheLocalScope(t *testing.T) {
	dir := newProject(t)
	out := mustCLI(t, dir, "", "record", "--type", "note", "--title", "My own note", "--scope", "local")
	id := strings.Fields(out)[0]

	path := filepath.Join(dir, projectDir, "local", id+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(string(data)+"\nAPI_TOKEN=placeholder-not-a-real-value\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := cli(t, dir, "", "check"); err != nil {
		t.Errorf("check failed on a local record: %v", err)
	}
}

func TestCheckJSON(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, "", "record", "--type", "note", "--title", "Clean")
	out := mustCLI(t, dir, "", "check", "--json")
	var got struct {
		Scanned     int   `json:"scanned"`
		GitleaksRan bool  `json:"gitleaks_ran"`
		Leaks       []any `json:"leaks"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got.Scanned != 1 || len(got.Leaks) != 0 {
		t.Errorf("%+v", got)
	}
}

func TestDedupeFindsRecordsAnsweringOneError(t *testing.T) {
	dir := newProject(t)
	const output = "./main.go:5:14: undefined: greet\n"
	first := strings.Fields(mustCLI(t, dir, output, "record", "--type", "fix",
		"--title", "Define greet", "--command", "go build ./...", "--error-file", "-"))[0]
	second := strings.Fields(mustCLI(t, dir, output, "record", "--type", "fix",
		"--title", "Import the package that defines greet", "--command", "go build ./...", "--error-file", "-"))[0]

	out := mustCLI(t, dir, "", "dedupe")
	for _, id := range []string{first, second} {
		if !strings.Contains(out, id) {
			t.Errorf("dedupe missed %s:\n%s", id, out)
		}
	}

	mustCLI(t, dir, "", "dedupe", "--supersede", first, "--by", second)

	// The retired record stays, marked, and stops being injected.
	shown := mustCLI(t, dir, "", "show", first)
	if !strings.Contains(shown, "superseded_by: "+second) {
		t.Errorf("superseded_by was not written:\n%s", shown)
	}
	recalled := mustCLI(t, dir, output, "recall", "--command", "go build ./...")
	if strings.Contains(recalled, "Define greet") {
		t.Errorf("a superseded record was still injected:\n%s", recalled)
	}
	if !strings.Contains(recalled, "Import the package") {
		t.Errorf("the replacement is not injected:\n%s", recalled)
	}
	if out := mustCLI(t, dir, "", "dedupe"); !strings.Contains(out, "no fingerprint has more than one") {
		t.Errorf("the group survived:\n%s", out)
	}
}

func TestDedupeRefusals(t *testing.T) {
	dir := newProject(t)
	id := strings.Fields(mustCLI(t, dir, "./main.go:5:14: undefined: greet\n", "record", "--type", "fix",
		"--title", "A fix", "--command", "go build ./...", "--error-file", "-"))[0]

	for _, args := range [][]string{
		{"dedupe", "--supersede", id},
		{"dedupe", "--by", id},
		{"dedupe", "--supersede", id, "--by", id},
		{"dedupe", "--supersede", id, "--by", "01K68P9AB2C3D4E5F6G7H8J9K0"},
	} {
		if _, _, err := cli(t, dir, "", args...); err == nil {
			t.Errorf("forgelore %s: got nil error", strings.Join(args, " "))
		}
	}
}

// TestTeamRecordsCarryNothingPersonal: the measurement ledger is local and
// must stay that way, and a committed record must not name a session.
func TestTeamRecordsCarryNothingPersonal(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, "./main.go:5:14: undefined: greet\n", "record", "--type", "fix",
		"--title", "A fix", "--command", "go build ./...", "--error-file", "-")
	mustCLI(t, dir, "./main.go:5:14: undefined: greet\n", "recall", "--command", "go build ./...", "--session", "a-session-id")

	entries, err := os.ReadDir(filepath.Join(dir, projectDir, "records"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(dir, projectDir, "records", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"a-session-id", "session", "cost", "est_tokens"} {
			if strings.Contains(string(data), forbidden) {
				t.Errorf("%s mentions %q:\n%s", entry.Name(), forbidden, data)
			}
		}
	}

	ignore, err := os.ReadFile(filepath.Join(dir, projectDir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"ledger/", "local/", "cache/"} {
		if !strings.Contains(string(ignore), want) {
			t.Errorf("%s is not excluded from the repository", want)
		}
	}
}

func TestInitCanInstallAPreCommitHook(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	out := mustCLI(t, dir, "", "init", "--with-git-hook")
	hook := filepath.Join(dir, ".git", "hooks", "pre-commit")
	data, err := os.ReadFile(hook)
	if err != nil {
		t.Fatalf("no hook: %v", err)
	}
	if !strings.Contains(string(data), "forgelore check") {
		t.Errorf("the hook does not run the check:\n%s", data)
	}
	if !strings.Contains(out, "pre-commit") {
		t.Errorf("init did not say it wrote a hook:\n%s", out)
	}
	info, err := os.Stat(hook)
	if err != nil || info.Mode()&0o111 == 0 {
		t.Errorf("the hook is not executable: %v", info.Mode())
	}
}

// TestInitRefusesToClobberAHook: a tool that silently replaces a project's
// own hook loses somebody a morning.
func TestInitRefusesToClobberAnExistingHook(t *testing.T) {
	dir := t.TempDir()
	hooks := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	mine := filepath.Join(hooks, "pre-commit")
	if err := os.WriteFile(mine, []byte("#!/bin/sh\necho mine\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, _, err := cli(t, dir, "", "init", "--with-git-hook")
	if err == nil {
		t.Fatal("init overwrote an existing hook")
	}
	if !strings.Contains(err.Error(), hookLine) {
		t.Errorf("the refusal does not say what to add: %v", err)
	}
	data, _ := os.ReadFile(mine)
	if !strings.Contains(string(data), "echo mine") {
		t.Error("the existing hook was changed")
	}
}

func TestInitWithoutAGitRepositorySaysSo(t *testing.T) {
	if _, _, err := cli(t, t.TempDir(), "", "init", "--with-git-hook"); err == nil {
		t.Error("init claimed to install a hook outside a repository")
	}
}

// TestInitPrintsTheCheckLine: even without the hook, somebody has to be told
// the check exists.
func TestInitPrintsTheCheckLine(t *testing.T) {
	if out := mustCLI(t, t.TempDir(), "", "init"); !strings.Contains(out, "forgelore check") {
		t.Errorf("init does not mention the check:\n%s", out)
	}
}

// TestTwoDevelopersDoNotCollide is the first half of the phase's acceptance
// criterion, run rather than argued: two clones, a record added in each, and
// a merge.
func TestTwoDevelopersDoNotCollide(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	origin := t.TempDir()
	git := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return string(out)
	}

	git(origin, "init", "--bare", "--initial-branch=main", ".")

	// One developer sets the project up and pushes.
	alice := t.TempDir()
	git(alice, "clone", origin, ".")
	mustCLI(t, alice, "", "init")
	git(alice, "add", "-A")
	git(alice, "commit", "-m", "set up forgelore")
	git(alice, "push", "origin", "main")

	bob := t.TempDir()
	git(bob, "clone", origin, ".")

	// Both record something, without seeing each other's work.
	mustCLI(t, alice, "./main.go:5:14: undefined: greet\n", "record", "--type", "fix",
		"--title", "Alice's fix", "--command", "go build ./...", "--error-file", "-")
	mustCLI(t, bob, "./other.go:9:2: undefined: parse\n", "record", "--type", "fix",
		"--title", "Bob's fix", "--command", "go build ./...", "--error-file", "-")

	git(alice, "add", "-A")
	git(alice, "commit", "-m", "alice")
	git(alice, "push", "origin", "main")

	git(bob, "add", "-A")
	git(bob, "commit", "-m", "bob")
	git(bob, "fetch", "origin")

	// The merge is the claim. A record is one file named by a ULID, so
	// there is nothing for two people to disagree about.
	cmd := exec.Command("git", "merge", "origin/main", "-m", "merge")
	cmd.Dir = bob
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the merge conflicted:\n%s", out)
	}

	// And both memories survived it.
	out := mustCLI(t, bob, "", "search", "fix")
	for _, want := range []string{"Alice's fix", "Bob's fix"} {
		if !strings.Contains(out, want) {
			t.Errorf("%q did not survive the merge:\n%s", want, out)
		}
	}

	// Nothing personal was committed.
	tracked := git(bob, "ls-files")
	for _, forbidden := range []string{"ledger/", "cache/", ".forgelore/local/"} {
		if strings.Contains(tracked, forbidden) {
			t.Errorf("%s was committed:\n%s", forbidden, tracked)
		}
	}
}
