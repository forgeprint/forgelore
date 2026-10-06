package fingerprint

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// corpusDir holds real output from real failing commands. See its README for
// how it was captured and what the two variants of each family are for.
const corpusDir = "../../testdata/errors"

// variantsIdentical lists the families whose two captured variants are byte
// for byte the same, so that comparing them proves nothing.
//
// go/unknown-import is the one. Its message names neither the package nor a
// line that moved — "main.go:3:8" is the same in both captures — so the
// capture that was meant to differ does not. The family stays in the corpus
// because it is a real error worth extracting; it is excluded from the
// variant comparison so that a test cannot claim evidence it does not have.
// Recapturing it needs the Windows machine the corpus was made on.
var variantsIdentical = map[string]bool{
	"go/unknown-import": true,

	// git says the same sentence wherever it is not a repository: the
	// directory it was run in appears nowhere in the message. There is no
	// way to vary the capture that does not also change the error, so the
	// two files are the same and the comparison is skipped rather than
	// counted as evidence.
	"git/not-a-repository": true,
}

// sameError groups families that are one error reached through different
// commands. They are expected to share a fingerprint, and the test below
// insists on it rather than tolerating it.
//
// This is the property that dropping the verb from a fingerprint bought: a
// fix recorded while running `go build` has to be found when `go vet`
// reports the same thing, and vet says it in a different shape.
var sameError = [][]string{
	{"go/undefined-identifier", "go/vet-undefined"},
}

// equivalent reports whether two families are the same error by declaration.
func equivalent(a, b string) bool {
	for _, group := range sameError {
		var foundA, foundB bool
		for _, family := range group {
			foundA = foundA || family == a
			foundB = foundB || family == b
		}
		if foundA && foundB {
			return true
		}
	}
	return false
}

type sample struct {
	family  string
	variant string
	command string
	output  string
}

// loadCorpus reads every captured sample. The frontmatter is the restricted
// YAML of ADR-0017, but only two fields matter here, so this reads them
// directly rather than pulling in the record decoder and its required fields.
func loadCorpus(t *testing.T) []sample {
	t.Helper()

	var samples []sample
	err := filepath.WalkDir(corpusDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".txt" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		command, body, err := splitSample(string(raw))
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		variant := strings.TrimSuffix(filepath.Base(path), ".txt")
		family := filepath.ToSlash(filepath.Dir(path))
		family = strings.TrimPrefix(family, filepath.ToSlash(corpusDir)+"/")
		samples = append(samples, sample{family: family, variant: variant, command: command, output: body})
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", corpusDir, err)
	}
	if len(samples) == 0 {
		t.Fatalf("no samples under %s", corpusDir)
	}
	sort.Slice(samples, func(i, j int) bool {
		if samples[i].family != samples[j].family {
			return samples[i].family < samples[j].family
		}
		return samples[i].variant < samples[j].variant
	})
	return samples
}

// splitSample separates the frontmatter from the captured output and returns
// the command that produced it.
func splitSample(raw string) (command, body string, err error) {
	lines := splitLines(raw)
	if len(lines) == 0 || lines[0] != "---" {
		return "", "", errNoFrontmatter
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return command, strings.Join(lines[i+1:], "\n"), nil
		}
		key, value, found := strings.Cut(lines[i], ":")
		if !found || strings.TrimSpace(key) != "command" {
			continue
		}
		command = strings.Trim(strings.TrimSpace(value), `"`)
	}
	return "", "", errNoFrontmatter
}

type corpusError string

func (e corpusError) Error() string { return string(e) }

const errNoFrontmatter = corpusError("sample has no closing frontmatter delimiter")

// TestCorpusExtraction prints what the extractor makes of every sample. It
// asserts only that each sample yields at least one diagnostic — a sample that
// yields none is an error the tool reported and this package did not see — and
// leaves the rest on the record for a human to read.
func TestCorpusExtraction(t *testing.T) {
	for _, s := range loadCorpus(t) {
		events := Scan(s.command, s.output)
		if len(events) == 0 {
			t.Errorf("%s/%s: no diagnostic extracted from output that a tool failed with", s.family, s.variant)
			continue
		}
		if s.variant != "a" {
			continue
		}
		for _, e := range events {
			code := e.Code
			if code == "" {
				code = "-"
			}
			t.Logf("%-32s %s  [%s] %s  %q", s.family, e.Sum, code, e.Command, e.Message)
		}
	}
}

// TestVariantsAgree is the first of the two claims the corpus exists to check:
// the same error, captured in a differently named workspace with the failing
// line in a different place, must fingerprint the same. A difference here is a
// miss — memory stays silent when it should have spoken.
func TestVariantsAgree(t *testing.T) {
	byFamily := make(map[string]map[string][]string)
	for _, s := range loadCorpus(t) {
		if byFamily[s.family] == nil {
			byFamily[s.family] = make(map[string][]string)
		}
		byFamily[s.family][s.variant] = sums(Scan(s.command, s.output))
	}

	var compared, skipped, missed int
	for _, family := range sortedKeys(byFamily) {
		if variantsIdentical[family] {
			skipped++
			t.Logf("%-32s skipped: the two captures are identical, so agreement proves nothing", family)
			continue
		}
		a, b := byFamily[family]["a"], byFamily[family]["b"]
		compared++
		if !equal(a, b) {
			missed++
			t.Errorf("%s: variants disagree\n  a: %v\n  b: %v", family, a, b)
		}
	}
	t.Logf("variants compared: %d, skipped: %d, misses: %d", compared, skipped, missed)
}

// TestFamiliesSeparate is the second claim: two different errors must not
// share a fingerprint. A collision here is worse than a miss, because it
// injects a hint for an error the user does not have.
func TestFamiliesSeparate(t *testing.T) {
	owner := make(map[string]string)
	var collisions int
	for _, s := range loadCorpus(t) {
		for _, e := range Scan(s.command, s.output) {
			switch prior, seen := owner[e.Sum]; {
			case !seen:
				owner[e.Sum] = s.family
			case prior != s.family && !equivalent(prior, s.family):
				collisions++
				t.Errorf("fingerprint %s is shared by %s and %s: %q", e.Sum, prior, s.family, e.Message)
			}
		}
	}
	t.Logf("distinct fingerprints: %d, collisions: %d", len(owner), collisions)
}

// TestCorpusMessages pins what normalisation leaves behind for one sample per
// toolchain. The set comparisons above would stay green if every rule grew
// sloppier at once; these do not.
func TestCorpusMessages(t *testing.T) {
	want := map[string][]Event{
		"go/undefined-identifier": {
			{Command: "go", Message: "undefined: greet"},
		},
		"go/panic-nil-map": {
			{Command: "go", Message: "panic: assignment to entry in nil map"},
		},
		"go/test-failure": {
			{Command: "go", Message: "--- FAIL: TestAnswer"},
			{Command: "go", Message: "answer = 41, want 42"},
		},
		"python/attribute-error": {
			{Command: "python", Message: "AttributeError: 'str' object has no attribute 'appendx'"},
		},
		"python/name-error": {
			{Command: "python", Message: "NameError: name 'total' is not defined"},
		},
		"typescript/cannot-find-name": {
			{Command: "tsc", Code: "TS2304", Message: "Cannot find name 'missingName'."},
		},
		"dotnet/name-does-not-exist": {
			{Command: "dotnet", Code: "CS0103", Message: "The name 'Greet' does not exist in the current context"},
		},
		"dotnet/nullreference-runtime": {
			{Command: "dotnet", Message: "System.NullReferenceException: Object reference not set to an instance of an object."},
		},
		"node/module-not-found": {
			{Command: "node", Code: "MODULE_NOT_FOUND", Message: "Error: Cannot find module 'no-such-package-here'"},
		},
	}

	byFamily := make(map[string]sample)
	for _, s := range loadCorpus(t) {
		if s.variant == "a" {
			byFamily[s.family] = s
		}
	}

	for _, family := range sortedKeys(want) {
		s, ok := byFamily[family]
		if !ok {
			t.Fatalf("%s is pinned here but missing from the corpus", family)
		}
		got := Scan(s.command, s.output)
		if len(got) != len(want[family]) {
			t.Errorf("%s: got %d diagnostics, want %d\n  got: %+v", family, len(got), len(want[family]), got)
			continue
		}
		for i, w := range want[family] {
			if got[i].Command != w.Command || got[i].Code != w.Code || got[i].Message != w.Message {
				t.Errorf("%s[%d]:\n  got  command=%q code=%q message=%q\n  want command=%q code=%q message=%q",
					family, i, got[i].Command, got[i].Code, got[i].Message, w.Command, w.Code, w.Message)
			}
		}
	}
}

// TestMultipleDiagnostics checks the decision that one output holds many
// errors. tsc reports three diagnostics for one broken file, and they reduce
// to two fingerprints: two of the three are "\':\' expected." at different
// columns, and once the column is gone they are the same error met twice.
// That is the intended answer rather than a lost diagnostic — the fix for
// a missing colon does not depend on which column it is missing from — but it
// is the reason this test counts two and not three.
func TestMultipleDiagnostics(t *testing.T) {
	for _, s := range loadCorpus(t) {
		if s.family != "typescript/syntax-error" || s.variant != "a" {
			continue
		}
		if want := 3; strings.Count(s.output, ": error TS") != want {
			t.Fatalf("the sample no longer prints %d diagnostics; this test is checking nothing", want)
		}
		got := Scan(s.command, s.output)
		if len(got) != 2 {
			t.Fatalf("got %d fingerprints, want 2: %+v", len(got), got)
		}
		if got[0].Sum == got[1].Sum {
			t.Errorf("the two distinct diagnostics share a fingerprint: %+v", got)
		}
		return
	}
	t.Fatal("typescript/syntax-error is missing from the corpus")
}

// TestRepeatedDiagnosticCountsOnce checks the other half of that decision:
// MSBuild prints every error a second time in its summary, and one error
// printed twice is still one error.
func TestRepeatedDiagnosticCountsOnce(t *testing.T) {
	for _, s := range loadCorpus(t) {
		if s.family != "dotnet/cannot-convert" || s.variant != "a" {
			continue
		}
		if strings.Count(s.output, "error CS0029") != 2 {
			t.Fatalf("the sample no longer prints the error twice; this test is checking nothing")
		}
		if got := Scan(s.command, s.output); len(got) != 1 {
			t.Fatalf("got %d diagnostics, want 1: %+v", len(got), got)
		}
		return
	}
	t.Fatal("dotnet/cannot-convert is missing from the corpus")
}

func sums(events []Event) []string {
	out := make([]string, 0, len(events))
	for _, e := range events {
		out = append(out, e.Sum)
	}
	sort.Strings(out)
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestOneErrorThroughTwoCommands is the other half of family separation, and
// the reason the verb is not in a fingerprint: `go build` and `go vet`
// report the same undefined identifier in different shapes, and a memory
// recorded under one has to be found under the other.
func TestOneErrorThroughTwoCommands(t *testing.T) {
	samples := make(map[string]sample)
	for _, s := range loadCorpus(t) {
		if s.variant == "a" {
			samples[s.family] = s
		}
	}

	for _, group := range sameError {
		// The shared error is the intersection of what each family
		// produces, not the first event of the first one: go vet prints a
		// package line that go build does not, so the families differ in
		// how many diagnostics they yield.
		var shared map[string]bool
		for _, family := range group {
			s, ok := samples[family]
			if !ok {
				t.Fatalf("%s is declared equivalent but is missing from the corpus", family)
			}
			sums := map[string]bool{}
			for _, e := range Scan(s.command, s.output) {
				sums[e.Sum] = true
			}
			if len(sums) == 0 {
				t.Errorf("%s: nothing extracted", family)
				continue
			}
			if shared == nil {
				shared = sums
				continue
			}
			for sum := range shared {
				if !sums[sum] {
					delete(shared, sum)
				}
			}
		}
		if len(shared) == 0 {
			t.Errorf("%v produce no fingerprint in common, so a memory recorded "+
				"under one would not be found under the others", group)
			continue
		}
		t.Logf("%v share %v", group, sortedSums(shared))
	}
}

func sortedSums(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for sum := range set {
		out = append(out, sum)
	}
	sort.Strings(out)
	return out
}
