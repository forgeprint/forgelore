package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// project writes config files into a fresh .forgelore directory. team and
// local may be empty, which means the file is not written at all.
func project(t *testing.T, team, local string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), ".forgelore")
	if err := os.MkdirAll(filepath.Join(root, "local"), 0o755); err != nil {
		t.Fatal(err)
	}
	if team != "" {
		write(t, filepath.Join(root, teamFile), team)
	}
	if local != "" {
		write(t, filepath.Join(root, localFile), local)
	}
	return root
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func noEnv(string) (string, bool) { return "", false }

func TestDefaultsWhenNothingIsConfigured(t *testing.T) {
	c, problems := Load(project(t, "", ""), noEnv, nil)
	if len(problems) != 0 {
		t.Errorf("an unconfigured project reported problems: %v", problems)
	}
	if got := c.Int("report.period_days"); got != 30 {
		t.Errorf("report.period_days = %d, want 30", got)
	}
	if got := c.Int("measure.ab.control_percent"); got != 0 {
		t.Errorf("A/B is on by default: control_percent = %d, want 0", got)
	}
	if !c.Bool("measure.ledger") {
		t.Error("the ledger is off by default")
	}
	if source, _ := c.Origin("report.period_days"); source != SourceDefault {
		t.Errorf("source = %q, want %q", source, SourceDefault)
	}
}

// TestPrecedence walks the whole chain of ADR-0018 from the bottom up. Each
// step adds a higher source and must win over everything below it.
func TestPrecedence(t *testing.T) {
	const key = "report.period_days"
	root := project(t, "report.period_days: 7\n", "report.period_days: 14\n")

	env := func(name string) (string, bool) {
		if name == "FORGELORE_REPORT_PERIOD_DAYS" {
			return "21", true
		}
		return "", false
	}

	cases := []struct {
		why       string
		useEnv    bool
		overrides map[string]string
		want      int64
		source    Source
	}{
		{why: "local beats team", want: 14, source: SourceLocal},
		{why: "the environment beats local", useEnv: true, want: 21, source: SourceEnv},
		{
			why:       "a flag beats the environment",
			useEnv:    true,
			overrides: map[string]string{key: "28"},
			want:      28,
			source:    SourceFlag,
		},
	}
	for _, tc := range cases {
		lookup := noEnv
		if tc.useEnv {
			lookup = env
		}
		c, problems := Load(root, lookup, tc.overrides)
		if len(problems) != 0 {
			t.Errorf("%s: problems: %v", tc.why, problems)
		}
		if got := c.Int(key); got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.why, got, tc.want)
		}
		if source, _ := c.Origin(key); source != tc.source {
			t.Errorf("%s: source = %q, want %q", tc.why, source, tc.source)
		}
	}
}

func TestTeamConfigAloneIsUsed(t *testing.T) {
	c, _ := Load(project(t, "measure.ab.control_percent: 20\n", ""), noEnv, nil)
	if got := c.Int("measure.ab.control_percent"); got != 20 {
		t.Errorf("got %d, want 20", got)
	}
	if source, _ := c.Origin("measure.ab.control_percent"); source != SourceTeam {
		t.Errorf("source = %q, want %q", source, SourceTeam)
	}
}

// TestUnknownKeyIsAWarning: a config written for a newer Forgelore has to
// keep working on an older one (ADR-0011).
func TestUnknownKeyIsAWarning(t *testing.T) {
	root := project(t, "report.period_days: 7\nfrom.a.later.version: 500\n", "")
	c, problems := Load(root, noEnv, nil)

	if got := c.Int("report.period_days"); got != 7 {
		t.Errorf("the known key was lost: got %d, want 7", got)
	}
	if len(problems) != 1 {
		t.Fatalf("got %d problems, want 1: %v", len(problems), problems)
	}
	if !strings.Contains(problems[0].Msg, "from.a.later.version") {
		t.Errorf("the warning does not name the key: %s", problems[0])
	}
}

// TestWrongTypeIsIgnoredNotFatal: the same reasoning as an unknown key. The
// default stands and doctor says why.
func TestWrongTypeIsIgnored(t *testing.T) {
	c, problems := Load(project(t, `report.period_days: "soon"`+"\n", ""), noEnv, nil)
	if got := c.Int("report.period_days"); got != 30 {
		t.Errorf("got %d, want the default 30", got)
	}
	if len(problems) != 1 || !strings.Contains(problems[0].Msg, "wants an integer") {
		t.Errorf("problems = %v", problems)
	}
}

// TestUnparseableFileIsSkipped: a half-written config must not stop a recall
// (ADR-0007).
func TestUnparseableFileIsSkipped(t *testing.T) {
	c, problems := Load(project(t, "report:\n  nested: 1\n", ""), noEnv, nil)
	if got := c.Int("report.period_days"); got != 30 {
		t.Errorf("got %d, want the default 30", got)
	}
	if len(problems) != 1 {
		t.Fatalf("got %d problems, want 1: %v", len(problems), problems)
	}
	if !strings.Contains(problems[0].String(), "config.yaml") {
		t.Errorf("the problem does not name the file: %s", problems[0])
	}
}

func TestBadEnvironmentValueIsAProblem(t *testing.T) {
	env := func(name string) (string, bool) {
		if name == "FORGELORE_REPORT_PERIOD_DAYS" {
			return "a fortnight", true
		}
		return "", false
	}
	c, problems := Load(project(t, "", ""), env, nil)
	if got := c.Int("report.period_days"); got != 30 {
		t.Errorf("got %d, want the default 30", got)
	}
	if len(problems) != 1 {
		t.Errorf("problems = %v", problems)
	}
}

func TestUnknownFlagOverrideIsAProblem(t *testing.T) {
	_, problems := Load(project(t, "", ""), noEnv, map[string]string{"nope.not.a.key": "1"})
	if len(problems) != 1 || !strings.Contains(problems[0].Msg, "unknown setting") {
		t.Errorf("problems = %v", problems)
	}
}

func TestEnvName(t *testing.T) {
	cases := map[string]string{
		"measure.ab.control_percent": "FORGELORE_MEASURE_AB_CONTROL_PERCENT",
		"report.period_days":         "FORGELORE_REPORT_PERIOD_DAYS",
		"measure.ledger":             "FORGELORE_MEASURE_LEDGER",
	}
	for key, want := range cases {
		if got := EnvName(key); got != want {
			t.Errorf("EnvName(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestBooleanFromEveryWritableSource(t *testing.T) {
	root := project(t, "measure.ledger: false\n", "")
	c, _ := Load(root, noEnv, nil)
	if c.Bool("measure.ledger") {
		t.Error("the team config did not turn the ledger off")
	}

	c, _ = Load(root, noEnv, map[string]string{"measure.ledger": "true"})
	if !c.Bool("measure.ledger") {
		t.Error("the flag did not turn the ledger back on")
	}

	_, problems := Load(root, noEnv, map[string]string{"measure.ledger": "yes"})
	if len(problems) != 1 {
		t.Errorf(`"yes" was accepted as a boolean: %v`, problems)
	}
}

// TestEffectiveListsEverySetting is what doctor prints, so every setting this
// version understands has to appear whether it was configured or not.
func TestEffectiveListsEverySetting(t *testing.T) {
	c, _ := Load(project(t, "report.period_days: 7\n", ""), noEnv, nil)
	effective := c.Effective()
	if len(effective) != len(known) {
		t.Fatalf("got %d settings, want %d", len(effective), len(known))
	}
	for _, s := range effective {
		if s.Key == "" || s.Source == "" || s.From == "" {
			t.Errorf("incomplete row: %+v", s)
		}
	}
	if effective[len(effective)-1].Key != "report.period_days" {
		t.Errorf("the declared order was not kept: %+v", effective)
	}
}

func TestKeysDocumentsEverySetting(t *testing.T) {
	for _, k := range Keys() {
		if k.Doc == "" {
			t.Errorf("%s has no description", k.Key)
		}
	}
}

func TestAskingForAnUnknownKeyPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("reading an undeclared setting did not panic")
		}
	}()
	c, _ := Load(project(t, "", ""), noEnv, nil)
	c.Int("nope.not.a.key")
}
