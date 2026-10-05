package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/forgeprint/forgelore/internal/measure"
)

// cliEnv is cli with an environment, for the one layer of ADR-0018 that the
// other tests cannot reach.
func cliEnv(t *testing.T, wd, stdin string, vars map[string]string, args ...string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	err := run(env{
		args: args, wd: wd,
		stdin: strings.NewReader(stdin), stdout: &out, stderr: &errBuf,
		lookupEnv: func(name string) (string, bool) { v, ok := vars[name]; return v, ok },
	})
	return out.String(), err
}

// teamConfig writes .forgelore/config.yaml in an initialised project.
func teamConfig(t *testing.T, dir, content string) {
	t.Helper()
	path := filepath.Join(dir, projectDir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// readLedger returns everything the project has recorded.
func readLedger(t *testing.T, dir string) *measure.Ledger {
	t.Helper()
	l, err := measure.Read(filepath.Join(dir, projectDir),
		time.Now().UTC().AddDate(0, 0, -1), time.Now().UTC().AddDate(0, 0, 1))
	if err != nil {
		t.Fatal(err)
	}
	return l
}

const goError = "# alpha\n./main.go:5:14: undefined: greet\n"

func TestInitIgnoresTheLedger(t *testing.T) {
	dir := newProject(t)
	data, err := os.ReadFile(filepath.Join(dir, projectDir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "ledger/") {
		t.Errorf("the ledger is not excluded from the repository:\n%s", data)
	}
}

func TestRecallRecordsAMiss(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, goError, "recall", "--command", "go build ./...", "--session", "s1")

	l := readLedger(t, dir)
	if len(l.Entries) != 1 {
		t.Fatalf("got %d ledger entries, want 1: %+v", len(l.Entries), l.Entries)
	}
	e := l.Entries[0]
	if e.Event != measure.EventMiss {
		t.Errorf("event = %q, want %q", e.Event, measure.EventMiss)
	}
	if e.Session != "s1" || e.Group != measure.GroupTreatment {
		t.Errorf("session=%q group=%q", e.Session, e.Group)
	}
	if e.Bytes != 0 {
		t.Errorf("a miss was charged %d bytes", e.Bytes)
	}
}

func TestRecallRecordsAnInjection(t *testing.T) {
	dir := newProject(t)
	const title = "Import the package that defines greet"
	mustCLI(t, dir, goError, "record", "--type", "fix", "--title", title,
		"--command", "go build ./...", "--error-file", "-")

	mustCLI(t, dir, goError, "recall", "--command", "go build ./...", "--session", "s1")

	l := readLedger(t, dir)
	if len(l.Entries) != 1 {
		t.Fatalf("got %d entries, want 1: %+v", len(l.Entries), l.Entries)
	}
	e := l.Entries[0]
	if e.Event != measure.EventInject {
		t.Fatalf("event = %q, want %q", e.Event, measure.EventInject)
	}
	if e.Bytes != len(title) {
		t.Errorf("bytes = %d, want the length of the injected title %d", e.Bytes, len(title))
	}
	if e.EstTokens != (len(title)+3)/4 {
		t.Errorf("est_tokens = %d", e.EstTokens)
	}
	if e.RecordID == "" {
		t.Error("the entry does not say which record was injected")
	}
}

// TestControlArmIsIndistinguishable is what makes the trial worth running: a
// control session has to see exactly what a session with no memory sees, or
// the agent can tell which arm it is in.
func TestControlArmIsIndistinguishable(t *testing.T) {
	dir := newProject(t)
	const title = "Import the package that defines greet"
	mustCLI(t, dir, goError, "record", "--type", "fix", "--title", title,
		"--command", "go build ./...", "--error-file", "-")

	teamConfig(t, dir, "measure.ab.control_percent: 100\n")
	out := mustCLI(t, dir, goError, "recall", "--command", "go build ./...", "--session", "s1")

	if strings.Contains(out, title) {
		t.Errorf("the control arm was shown the hint:\n%s", out)
	}
	if !strings.Contains(out, "nothing recorded") {
		t.Errorf("the control arm does not look like a miss:\n%s", out)
	}

	jsonOut := mustCLI(t, dir, goError, "recall", "--command", "go build ./...", "--session", "s1", "--json")
	if strings.Contains(jsonOut, title) {
		t.Errorf("the control arm leaked the hint through --json:\n%s", jsonOut)
	}

	// The ledger still has to know a match existed, or the arms cannot be
	// compared.
	l := readLedger(t, dir)
	if len(l.Entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(l.Entries))
	}
	for _, e := range l.Entries {
		if e.Event != measure.EventControl {
			t.Errorf("event = %q, want %q", e.Event, measure.EventControl)
		}
		if e.Group != measure.GroupControl {
			t.Errorf("group = %q", e.Group)
		}
	}
}

func TestLedgerCanBeTurnedOff(t *testing.T) {
	dir := newProject(t)
	teamConfig(t, dir, "measure.ledger: false\n")
	mustCLI(t, dir, goError, "recall", "--command", "go build ./...", "--session", "s1")

	if l := readLedger(t, dir); len(l.Entries) != 0 {
		t.Errorf("the ledger was written while switched off: %+v", l.Entries)
	}
}

// statusLineJSON is the shape Claude Code hands its status line command.
func statusLineJSON(session string, cost float64) string {
	return fmt.Sprintf(`{
	  "session_id": %q,
	  "version": "2.1.90",
	  "cost": {"total_cost_usd": %v, "total_duration_ms": 45000},
	  "context_window": {"total_input_tokens": 15500, "total_output_tokens": 1200}
	}`, session, cost)
}

func TestUsageReadsTheStatusLinePayload(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, statusLineJSON("s1", 0.1234), "usage")

	l := readLedger(t, dir)
	if len(l.Usage) != 1 {
		t.Fatalf("got %d usage rows, want 1", len(l.Usage))
	}
	u := l.Usage[0]
	if u.Session != "s1" || u.CostUSD != 0.1234 {
		t.Errorf("session=%q cost=%v", u.Session, u.CostUSD)
	}
	if u.AgentVersion != "2.1.90" || u.Agent != "claude-code" {
		t.Errorf("agent=%q version=%q", u.Agent, u.AgentVersion)
	}
	if u.DurationMS != 45000 || u.ContextIn != 15500 {
		t.Errorf("duration=%d contextIn=%d", u.DurationMS, u.ContextIn)
	}
}

func TestUsagePrintsNothingSoItComposesInAPipe(t *testing.T) {
	dir := newProject(t)
	if out := mustCLI(t, dir, statusLineJSON("s1", 0.5), "usage"); out != "" {
		t.Errorf("usage wrote %q to stdout", out)
	}
}

func TestUsageRejectsWhatItCannotAttribute(t *testing.T) {
	dir := newProject(t)
	cases := map[string]struct {
		stdin string
		args  []string
	}{
		"no session id in the payload": {stdin: `{"cost":{"total_cost_usd":1}}`, args: []string{"usage"}},
		"not JSON at all":              {stdin: "the status line prose", args: []string{"usage"}},
		"empty stdin":                  {stdin: "", args: []string{"usage"}},
		"--format none without a cost": {stdin: "", args: []string{"usage", "--format", "none", "--session", "s1"}},
		"an unknown format":            {stdin: "{}", args: []string{"usage", "--format", "nope"}},
	}
	for why, c := range cases {
		if _, _, err := cli(t, dir, c.stdin, c.args...); err == nil {
			t.Errorf("%s: got nil error", why)
		}
	}
}

func TestUsageWithoutJSONForOtherAgents(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, "", "usage", "--format", "none", "--agent", "some-agent", "--session", "s9", "--cost-usd", "2.5")

	l := readLedger(t, dir)
	if len(l.Usage) != 1 || l.Usage[0].CostUSD != 2.5 || l.Usage[0].Agent != "some-agent" {
		t.Errorf("usage = %+v", l.Usage)
	}
}

func TestUsageWiringHelp(t *testing.T) {
	out := mustCLI(t, newProject(t), "", "usage", "--help-wiring")
	if !strings.Contains(out, "statusLine") {
		t.Errorf("the wiring help does not mention the status line:\n%s", out)
	}
}

// TestReportOnSyntheticData is the acceptance criterion's first half: given a
// ledger whose answer is known, the report has to produce it.
func TestReportOnSyntheticData(t *testing.T) {
	dir := newProject(t)
	root := filepath.Join(dir, projectDir)
	now := time.Now().UTC().Add(-time.Hour)

	// Ten sessions a side. Treatment costs one dollar, control three.
	for i := 0; i < 10; i++ {
		for _, arm := range []struct {
			group string
			cost  float64
		}{{measure.GroupTreatment, 1}, {measure.GroupControl, 3}} {
			session := fmt.Sprintf("%s-%d", arm.group, i)
			if err := measure.AppendUsage(root, measure.Usage{
				Time: now, Session: session, Group: arm.group, Agent: "claude-code", CostUSD: arm.cost,
			}); err != nil {
				t.Fatal(err)
			}
			if err := measure.AppendEntry(root, measure.Entry{
				Time: now, Session: session, Group: arm.group,
				Event: measure.EventInject, Fingerprint: "aaaa", Bytes: 40, EstTokens: 10,
			}); err != nil {
				t.Fatal(err)
			}
		}
	}

	out := mustCLI(t, dir, "", "report")
	for _, want := range []string{"hints injected      20", "bytes injected      800", "Cost per session", "Treatment is lower by $2.0000"} {
		if !strings.Contains(out, want) {
			t.Errorf("report is missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "estimated at four bytes per token") {
		t.Errorf("the token figure is not labelled an estimate:\n%s", out)
	}
}

// TestReportRefusesToClaimASaving is the honesty boundary of step 5.
func TestReportWithoutUsageClaimsNothing(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, goError, "recall", "--command", "go build ./...", "--session", "s1")

	out := mustCLI(t, dir, "", "report")
	if !strings.Contains(out, "No usage was reported") {
		t.Errorf("report did not say the saving is unmeasured:\n%s", out)
	}
	if strings.Contains(out, "Treatment is lower") {
		t.Errorf("report claimed a saving with no usage data:\n%s", out)
	}
	if !strings.Contains(out, "errors looked up    1") {
		t.Errorf("the spending side is missing:\n%s", out)
	}
}

// TestReportMarksAnUnmeaningfulDifference: two arms drawn alike must be
// reported as not meaningful rather than as a small win.
func TestReportMarksAnUnmeaningfulDifference(t *testing.T) {
	dir := newProject(t)
	root := filepath.Join(dir, projectDir)
	now := time.Now().UTC().Add(-time.Hour)

	costs := []float64{1.0, 1.1, 0.9, 1.2, 0.8, 1.05, 0.95, 1.15}
	for i, cost := range costs {
		for _, group := range []string{measure.GroupTreatment, measure.GroupControl} {
			if err := measure.AppendUsage(root, measure.Usage{
				Time: now, Session: fmt.Sprintf("%s-%d", group, i), Group: group, CostUSD: cost,
			}); err != nil {
				t.Fatal(err)
			}
		}
	}

	out := mustCLI(t, dir, "", "report")
	if !strings.Contains(out, "Not meaningful") {
		t.Errorf("an identical pair of arms was not marked as such:\n%s", out)
	}
}

func TestReportJSON(t *testing.T) {
	dir := newProject(t)
	mustCLI(t, dir, goError, "recall", "--command", "go build ./...", "--session", "s1")

	out := mustCLI(t, dir, "", "report", "--json")
	var got struct {
		Injections int `json:"Injections"`
		Misses     int `json:"Misses"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got.Misses != 1 {
		t.Errorf("misses = %d, want 1:\n%s", got.Misses, out)
	}
}

func TestReportDaysFlagOverridesConfig(t *testing.T) {
	dir := newProject(t)
	teamConfig(t, dir, "report.period_days: 90\n")
	if out := mustCLI(t, dir, "", "report"); !strings.Contains(out, "Last 90 days") {
		t.Errorf("the configured period was not used:\n%s", out)
	}
	if out := mustCLI(t, dir, "", "report", "--days", "7"); !strings.Contains(out, "Last 7 days") {
		t.Errorf("--days did not win:\n%s", out)
	}
}

// TestDoctorShowsWhereEachSettingCameFrom: ADR-0018 requires it, because five
// sources mean no single file shows the effective value.
func TestDoctorShowsSettingProvenance(t *testing.T) {
	dir := newProject(t)
	teamConfig(t, dir, "report.period_days: 7\n")

	out := mustCLI(t, dir, "", "doctor")
	if !strings.Contains(out, "report.period_days") || !strings.Contains(out, "config.yaml") {
		t.Errorf("doctor does not say where the value came from:\n%s", out)
	}
	if !strings.Contains(out, "built in") {
		t.Errorf("doctor does not show the defaults:\n%s", out)
	}
}

func TestDoctorReportsAnUnknownSetting(t *testing.T) {
	dir := newProject(t)
	teamConfig(t, dir, "measure.ledger: true\nnot.a.setting: 1\n")

	out := mustCLI(t, dir, "", "doctor")
	if !strings.Contains(out, "not.a.setting") {
		t.Errorf("doctor did not report the unknown key:\n%s", out)
	}
	if strings.Contains(out, "nothing to report") {
		t.Errorf("doctor called a store with a bad config clean:\n%s", out)
	}
}

// TestEnvironmentOverridesTheFile exercises the layer of ADR-0018 that no
// file can reach.
func TestEnvironmentOverridesTheFile(t *testing.T) {
	dir := newProject(t)
	teamConfig(t, dir, "report.period_days: 7\n")

	out, err := cliEnv(t, dir, "", map[string]string{"FORGELORE_REPORT_PERIOD_DAYS": "45"}, "report")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Last 45 days") {
		t.Errorf("the environment did not override the file:\n%s", out)
	}
}

// TestBrokenConfigDoesNotStopARecall: fail open (ADR-0007).
func TestBrokenConfigDoesNotStopARecall(t *testing.T) {
	dir := newProject(t)
	teamConfig(t, dir, "measure:\n  nested: true\n")

	out, _, err := cli(t, dir, goError, "recall", "--command", "go build ./...", "--session", "s1")
	if err != nil {
		t.Fatalf("a broken config stopped a recall: %v", err)
	}
	if !strings.Contains(out, "nothing recorded") {
		t.Errorf("recall did not run:\n%s", out)
	}
	if out := mustCLI(t, dir, "", "doctor"); !strings.Contains(out, "config.yaml") {
		t.Errorf("doctor did not report the broken config:\n%s", out)
	}
}
