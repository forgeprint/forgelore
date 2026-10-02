package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cli runs the whole command line the way main does, without a process. It
// returns stdout, stderr and the error run gave back.
func cli(t *testing.T, wd, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	err = run(env{
		args:   args,
		wd:     wd,
		stdin:  strings.NewReader(stdin),
		stdout: &out,
		stderr: &errBuf,
	})
	return out.String(), errBuf.String(), err
}

// mustCLI fails the test if the command does not succeed.
func mustCLI(t *testing.T, wd, stdin string, args ...string) string {
	t.Helper()
	out, errOut, err := cli(t, wd, stdin, args...)
	if err != nil {
		t.Fatalf("forgelore %s: %v\nstderr: %s", strings.Join(args, " "), err, errOut)
	}
	return out
}

// newProject returns an initialised project in a directory of its own.
func newProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustCLI(t, dir, "", "init")
	return dir
}

func TestRunVersion(t *testing.T) {
	out := mustCLI(t, t.TempDir(), "", "version")
	if !strings.HasPrefix(out, "forgelore ") {
		t.Errorf("version output = %q, want it to start with %q", out, "forgelore ")
	}
}

func TestRunNoArgsPrintsUsage(t *testing.T) {
	out := mustCLI(t, t.TempDir(), "")
	if !strings.Contains(out, "Usage:") {
		t.Errorf("no-args output = %q, want it to contain %q", out, "Usage:")
	}
}

func TestRunUnknownCommand(t *testing.T) {
	out, _, err := cli(t, t.TempDir(), "", "nope")
	if err == nil {
		t.Fatal("run with unknown command: got nil error, want an error")
	}
	if out != "" {
		t.Errorf("unknown command wrote %q to stdout, want nothing", out)
	}
}

func TestInitCreatesTheStore(t *testing.T) {
	dir := t.TempDir()
	out := mustCLI(t, dir, "", "init")

	for _, want := range []string{"records", "local", ".gitignore"} {
		path := filepath.Join(dir, projectDir, want)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("init did not create %s: %v", want, err)
		}
	}
	if !strings.Contains(out, "AGENTS.md") {
		t.Errorf("init output does not mention AGENTS.md:\n%s", out)
	}
	if strings.Contains(out, "kept") {
		t.Errorf("a first init reported something as kept:\n%s", out)
	}

	ignore, err := os.ReadFile(filepath.Join(dir, projectDir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"cache/", "local/"} {
		if !strings.Contains(string(ignore), want) {
			t.Errorf(".forgelore/.gitignore does not exclude %s", want)
		}
	}
}

// TestInitDoesNotWriteAgentsFile: the pointer is printed, never written.
// AGENTS.md belongs to the project.
func TestInitDoesNotWriteAgentsFile(t *testing.T) {
	dir := t.TempDir()
	mustCLI(t, dir, "", "init")
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Errorf("init wrote AGENTS.md")
	}
}

func TestInitIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	mustCLI(t, dir, "", "init")

	// A hand-edited .gitignore must survive a second init.
	ignore := filepath.Join(dir, projectDir, ".gitignore")
	if err := os.WriteFile(ignore, []byte("cache/\nlocal/\n# mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := mustCLI(t, dir, "", "init")
	if !strings.Contains(out, "kept") {
		t.Errorf("a second init reported nothing as kept:\n%s", out)
	}
	data, err := os.ReadFile(ignore)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "# mine") {
		t.Error("a second init overwrote a hand-edited .gitignore")
	}
}

func TestCommandsNeedAProject(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{
		{"search", "anything"},
		{"stats"},
		{"doctor"},
		{"index", "rebuild"},
		{"record", "--type", "note", "--title", "x"},
	} {
		_, _, err := cli(t, dir, "", args...)
		if err == nil {
			t.Errorf("forgelore %s in a bare directory: got nil error", strings.Join(args, " "))
			continue
		}
		if !strings.Contains(err.Error(), "forgelore init") {
			t.Errorf("forgelore %s: error %q does not say how to fix it", strings.Join(args, " "), err)
		}
	}
}

func TestProjectIsFoundFromASubdirectory(t *testing.T) {
	dir := newProject(t)
	sub := filepath.Join(dir, "services", "api")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	// An agent runs its commands where the failing build is, not at the root.
	mustCLI(t, sub, "", "stats")
}

// TestRecordThenRecall is the acceptance path of phase 2: an error is
// recorded against its fingerprint, and the same error met again finds it.
func TestRecordThenRecall(t *testing.T) {
	dir := newProject(t)
	const output = "# alpha\n./main.go:5:14: undefined: greet\n"

	mustCLI(t, dir, output,
		"record",
		"--type", "fix",
		"--title", "Import the package that defines greet",
		"--command", "go build ./...",
		"--error-file", "-",
	)

	// The same error, in another workspace, on another line.
	const later = "# beta\n./main.go:12:14: undefined: greet\n"
	out := mustCLI(t, dir, later, "recall", "--command", "go test ./...", "--error-file", "-")

	if !strings.Contains(out, "Import the package that defines greet") {
		t.Errorf("recall did not find the recorded fix:\n%s", out)
	}
	if !strings.Contains(out, "1 error(s), 1 with something recorded") {
		t.Errorf("recall did not report one known error:\n%s", out)
	}
}

func TestRecallReportsAnUnknownError(t *testing.T) {
	dir := newProject(t)
	out := mustCLI(t, dir, "./main.go:5:14: undefined: greet\n", "recall", "--command", "go build ./...")
	if !strings.Contains(out, "nothing recorded") {
		t.Errorf("recall did not say the error is unknown:\n%s", out)
	}
}

// TestRecallWithoutAProjectStillFingerprints: recall is the injection path
// and must not fail on a project that was never initialised (ADR-0007).
func TestRecallWithoutAProject(t *testing.T) {
	dir := t.TempDir()
	out, _, err := cli(t, dir, "./main.go:5:14: undefined: greet\n", "recall", "--command", "go build ./...")
	if err != nil {
		t.Fatalf("recall without a project returned an error: %v", err)
	}
	if !strings.Contains(out, "nothing recorded") {
		t.Errorf("recall without a project did not fingerprint the error:\n%s", out)
	}
}

func TestRecallOnUnrecognisedOutput(t *testing.T) {
	dir := newProject(t)
	out := mustCLI(t, dir, "something went wrong, no idea what\n", "recall", "--command", "make")
	if !strings.Contains(out, "no error recognised") {
		t.Errorf("recall invented an error:\n%s", out)
	}
}

// TestRecordRefusesToGuessBetweenErrors: output with two errors in it does
// not say which one the fix is for, and this command does not get to decide.
func TestRecordRefusesToGuessBetweenErrors(t *testing.T) {
	dir := newProject(t)
	const output = "app.ts(1,13): error TS2304: Cannot find name 'a'.\n" +
		"app.ts(2,7): error TS2322: Type 'string' is not assignable to type 'number'.\n"

	_, _, err := cli(t, dir, output,
		"record", "--type", "fix", "--title", "x",
		"--command", "tsc", "--error-file", "-")
	if err == nil {
		t.Fatal("got nil error, want a refusal")
	}
	if !strings.Contains(err.Error(), "--fingerprint") {
		t.Errorf("the refusal does not say how to proceed: %v", err)
	}
	if strings.Count(err.Error(), "\n  ") != 2 {
		t.Errorf("the refusal does not list both fingerprints: %v", err)
	}
}

func TestRecordRequiresATitle(t *testing.T) {
	dir := newProject(t)
	_, _, err := cli(t, dir, "", "record", "--type", "note")
	if err == nil || !strings.Contains(err.Error(), "--title") {
		t.Errorf("got %v, want a complaint about --title", err)
	}
}

func TestRecordRejectsTwoWaysOfSayingTheSameThing(t *testing.T) {
	dir := newProject(t)
	for _, args := range [][]string{
		{"record", "--type", "note", "--title", "x", "--body", "a", "--body-file", "-"},
		{"record", "--type", "fix", "--title", "x", "--fingerprint", "abcd", "--error-file", "-"},
	} {
		if _, _, err := cli(t, dir, "", args...); err == nil {
			t.Errorf("forgelore %s: got nil error", strings.Join(args, " "))
		}
	}
}

// TestRecordRedacts: a secret pasted into a record never reaches the disk,
// and the user is told (ADR-0013).
//
// The value below is deliberately dull. An entropic-looking fixture is the
// kind of thing scripts/gitleaks.sh is there to stop, and a test that has to
// be allowlisted would blind the scanner to a real leak in this same file.
func TestRecordRedacts(t *testing.T) {
	const fixture = "placeholder-not-a-real-value"

	dir := newProject(t)
	out := mustCLI(t, dir, "The fix needs AWS_SECRET_ACCESS_KEY="+fixture+"\n",
		"record", "--type", "note", "--title", "Credentials live in the environment", "--body-file", "-")
	if !strings.Contains(out, "redacted") {
		t.Fatalf("record did not report the redaction:\n%s", out)
	}

	id := strings.Fields(out)[0]
	shown := mustCLI(t, dir, "", "show", id)
	if strings.Contains(shown, fixture) {
		t.Error("the secret reached the record file")
	}
}

// TestSearchReturnsNoBody is staged access: a search cannot be made to return
// the contents of the store (ADR-0006).
func TestSearchReturnsNoBody(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, "the body holds a distinctive phrase: xyzzyplugh",
		"record", "--type", "note", "--title", "A searchable title", "--body-file", "-")

	out := mustCLI(t, dir, "", "search", "searchable")
	if !strings.Contains(out, "A searchable title") {
		t.Fatalf("search did not find the record:\n%s", out)
	}
	if strings.Contains(out, "xyzzyplugh") {
		t.Error("search returned the body")
	}
}

func TestSearchNeedsAQuery(t *testing.T) {
	dir := newProject(t)
	if _, _, err := cli(t, dir, "", "search"); err == nil {
		t.Error("got nil error, want a complaint about the missing query")
	}
}

func TestShowReturnsTheWholeRecord(t *testing.T) {
	dir := newProject(t)
	out := mustCLI(t, dir, "the body", "record", "--type", "note", "--title", "A title", "--body-file", "-")
	id := strings.Fields(out)[0]

	shown := mustCLI(t, dir, "", "show", id)
	for _, want := range []string{"schema: 1", "id: " + id, "A title", "the body"} {
		if !strings.Contains(shown, want) {
			t.Errorf("show output is missing %q:\n%s", want, shown)
		}
	}
}

func TestShowRejectsSomethingThatIsNotAnID(t *testing.T) {
	dir := newProject(t)
	if _, _, err := cli(t, dir, "", "show", "not-an-id"); err == nil {
		t.Error("got nil error, want a complaint")
	}
}

func TestDoctorReportsASkippedFile(t *testing.T) {
	dir := newProject(t)
	bad := filepath.Join(dir, projectDir, "records", "01K68P9AB2C3D4E5F6G7H8J9K0.md")
	if err := os.WriteFile(bad, []byte("not a record at all\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := mustCLI(t, dir, "", "doctor")
	if !strings.Contains(out, "skipped") {
		t.Errorf("doctor did not report the unreadable file:\n%s", out)
	}
}

func TestDoctorOnACleanStore(t *testing.T) {
	dir := newProject(t)
	out := mustCLI(t, dir, "", "doctor")
	if !strings.Contains(out, "nothing to report") {
		t.Errorf("doctor complained about a clean store:\n%s", out)
	}
}

func TestStatsCounts(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, "", "record", "--type", "note", "--title", "One")
	mustCLI(t, dir, "", "record", "--type", "decision", "--title", "Two", "--scope", "local")

	out := mustCLI(t, dir, "", "stats")
	for _, want := range []string{"2 record(s)", "note       1", "decision   1", "team       1", "local      1"} {
		if !strings.Contains(out, want) {
			t.Errorf("stats output is missing %q:\n%s", want, out)
		}
	}
}

func TestIndexRebuild(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, "", "record", "--type", "note", "--title", "One")

	out := mustCLI(t, dir, "", "index", "rebuild")
	if !strings.Contains(out, "1 indexed") {
		t.Errorf("rebuild did not index the record:\n%s", out)
	}
}

func TestIndexNeedsASubcommand(t *testing.T) {
	dir := newProject(t)
	for _, args := range [][]string{{"index"}, {"index", "nope"}} {
		if _, _, err := cli(t, dir, "", args...); err == nil {
			t.Errorf("forgelore %s: got nil error", strings.Join(args, " "))
		}
	}
}

// TestJSONOutput checks that every command that takes --json produces one
// parseable object. An agent that has to pick an answer out of prose is an
// agent that will pick wrong.
func TestJSONOutput(t *testing.T) {
	dir := newProject(t)
	recorded := mustCLI(t, dir, "", "record", "--type", "note", "--title", "A title", "--json")

	var rec struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(recorded), &rec); err != nil {
		t.Fatalf("record --json: %v\n%s", err, recorded)
	}
	if rec.ID == "" || rec.Path == "" {
		t.Errorf("record --json is missing fields: %s", recorded)
	}

	cases := [][]string{
		{"init", "--json"},
		{"search", "title", "--json"},
		{"show", rec.ID, "--json"},
		{"stats", "--json"},
		{"doctor", "--json"},
		{"index", "rebuild", "--json"},
		{"recall", "--command", "go build ./...", "--json"},
	}
	for _, args := range cases {
		out := mustCLI(t, dir, "./main.go:5:14: undefined: greet\n", args...)
		var any map[string]any
		if err := json.Unmarshal([]byte(out), &any); err != nil {
			t.Errorf("forgelore %s --json is not an object: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
}

// TestRecallJSONCarriesTheFingerprint: the hook and MCP layers are built on
// this output, so the fingerprint has to be in it.
func TestRecallJSONCarriesTheFingerprint(t *testing.T) {
	dir := newProject(t)
	out := mustCLI(t, dir, "./main.go:5:14: undefined: greet\n",
		"recall", "--command", "go build ./...", "--json")

	var got struct {
		Errors []struct {
			Fingerprint string `json:"fingerprint"`
			Message     string `json:"message"`
			Matches     []any  `json:"matches"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if len(got.Errors) != 1 {
		t.Fatalf("got %d errors, want 1:\n%s", len(got.Errors), out)
	}
	if len(got.Errors[0].Fingerprint) != 16 {
		t.Errorf("fingerprint %q is not sixteen characters", got.Errors[0].Fingerprint)
	}
	if got.Errors[0].Message != "undefined: greet" {
		t.Errorf("message = %q", got.Errors[0].Message)
	}
	if got.Errors[0].Matches == nil {
		t.Error("matches is null; it should be an empty list so a caller can range over it")
	}
}

// TestFlagsAfterPositionalArguments: `forgelore search sqlite --json` is what
// anyone would type, and before parseInterspersed it searched for the string
// "sqlite --json".
func TestFlagsAfterPositionalArguments(t *testing.T) {
	dir := newProject(t)
	recorded := mustCLI(t, dir, "", "record", "--type", "note", "--title", "Vendoring costs 135 MB", "--json")
	var rec struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(recorded), &rec); err != nil {
		t.Fatal(err)
	}

	out := mustCLI(t, dir, "", "search", "vendoring", "--json")
	var got struct {
		Query   string `json:"query"`
		Results []struct {
			ID string `json:"id"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got.Query != "vendoring" {
		t.Errorf("query = %q, want %q", got.Query, "vendoring")
	}
	if len(got.Results) != 1 {
		t.Fatalf("got %d results, want 1:\n%s", len(got.Results), out)
	}

	// The same for show, whose positional argument comes first.
	shown := mustCLI(t, dir, "", "show", rec.ID, "--json")
	if !strings.HasPrefix(strings.TrimSpace(shown), "{") {
		t.Errorf("show did not honour --json after the id:\n%s", shown)
	}
}
