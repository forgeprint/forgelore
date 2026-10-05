package measure

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func at(day int) time.Time {
	return time.Date(2026, 10, day, 12, 0, 0, 0, time.UTC)
}

// TestAssignIsStable is the whole reason assignment is hashed rather than
// drawn: every invocation is a new process, and the hundredth lookup in a
// session has to land where the first one did.
func TestAssignIsStable(t *testing.T) {
	first := Assign("", "session-abc", 20)
	for i := 0; i < 100; i++ {
		if got := Assign("", "session-abc", 20); got != first {
			t.Fatalf("call %d moved the session from %q to %q", i, first, got)
		}
	}
}

func TestAssignEdges(t *testing.T) {
	for _, percent := range []int{0, -1} {
		if got := Assign("", "s", percent); got != GroupTreatment {
			t.Errorf("control_percent %d: got %q, want everything in treatment", percent, got)
		}
	}
	for _, percent := range []int{100, 101} {
		if got := Assign("", "s", percent); got != GroupControl {
			t.Errorf("control_percent %d: got %q, want everything in control", percent, got)
		}
	}
}

// TestAssignHitsTheRequestedShare: the split has to be roughly what was asked
// for, or the two arms are not comparable.
func TestAssignHitsTheRequestedShare(t *testing.T) {
	const n = 10000
	for _, percent := range []int{10, 20, 50} {
		control := 0
		for i := 0; i < n; i++ {
			if Assign("", fmt.Sprintf("session-%d", i), percent) == GroupControl {
				control++
			}
		}
		got := float64(control) / n * 100
		if diff := got - float64(percent); diff < -2 || diff > 2 {
			t.Errorf("control_percent %d produced %.1f%%, want within two points", percent, got)
		}
	}
}

// TestSaltStartsAFreshTrial: changing the salt has to reshuffle sessions,
// otherwise a second experiment reuses the first one's assignment.
func TestSaltReshuffles(t *testing.T) {
	moved := 0
	for i := 0; i < 1000; i++ {
		session := fmt.Sprintf("session-%d", i)
		if Assign("first", session, 50) != Assign("second", session, 50) {
			moved++
		}
	}
	if moved < 400 || moved > 600 {
		t.Errorf("%d of 1000 sessions moved arm; want roughly half", moved)
	}
}

func TestEstimateTokens(t *testing.T) {
	cases := map[string]int{"": 0, "a": 1, "abcd": 1, "abcde": 2, "12345678": 2}
	for in, want := range cases {
		if got := EstimateTokens(in); got != want {
			t.Errorf("EstimateTokens(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestLedgerRoundTrip(t *testing.T) {
	root := t.TempDir()
	want := Entry{
		Time: at(3), Session: "s1", Group: GroupTreatment, Event: EventInject,
		Fingerprint: "abcd1234abcd1234", RecordID: "01K68P9AB2C3D4E5F6G7H8J9K0",
		Bytes: 120, EstTokens: 30,
	}
	if err := AppendEntry(root, want); err != nil {
		t.Fatal(err)
	}
	if err := AppendUsage(root, Usage{Time: at(3), Session: "s1", Group: GroupTreatment, Agent: "claude-code", CostUSD: 0.5}); err != nil {
		t.Fatal(err)
	}

	l, err := Read(root, at(1), at(31))
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Entries) != 1 || l.Entries[0] != want {
		t.Errorf("entries = %+v, want [%+v]", l.Entries, want)
	}
	if len(l.Usage) != 1 || l.Usage[0].CostUSD != 0.5 {
		t.Errorf("usage = %+v", l.Usage)
	}
}

func TestLedgerSplitsByMonth(t *testing.T) {
	root := t.TempDir()
	for _, when := range []time.Time{at(3), time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)} {
		if err := AppendEntry(root, Entry{Time: when, Session: "s", Event: EventMiss}); err != nil {
			t.Fatal(err)
		}
	}
	names, err := os.ReadDir(filepath.Join(root, Dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("got %d files, want one per month: %v", len(names), names)
	}
}

func TestLedgerHonoursThePeriod(t *testing.T) {
	root := t.TempDir()
	for _, day := range []int{1, 10, 20} {
		if err := AppendEntry(root, Entry{Time: at(day), Session: "s", Event: EventMiss}); err != nil {
			t.Fatal(err)
		}
	}
	l, err := Read(root, at(5), at(15))
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Entries) != 1 {
		t.Errorf("got %d entries in a ten day window, want 1", len(l.Entries))
	}
}

// TestTruncatedLineIsCountedNotFatal: a process killed mid-write leaves half
// a line, and that must not cost the whole report.
func TestTruncatedLineIsCountedNotFatal(t *testing.T) {
	root := t.TempDir()
	if err := AppendEntry(root, Entry{Time: at(3), Session: "s", Event: EventMiss}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, Dir, "injections-2026-10.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"time":"2026-10-03T12:00`); err != nil {
		t.Fatal(err)
	}
	f.Close()

	l, err := Read(root, at(1), at(31))
	if err != nil {
		t.Fatalf("a truncated line stopped the read: %v", err)
	}
	if len(l.Entries) != 1 {
		t.Errorf("got %d good entries, want 1", len(l.Entries))
	}
	if l.Unreadable != 1 {
		t.Errorf("unreadable = %d, want 1", l.Unreadable)
	}
}

func TestReadingAnAbsentLedgerIsEmpty(t *testing.T) {
	l, err := Read(t.TempDir(), at(1), at(31))
	if err != nil {
		t.Fatalf("a project that has never recorded anything errored: %v", err)
	}
	if len(l.Entries) != 0 || len(l.Usage) != 0 {
		t.Errorf("got %+v, want nothing", l)
	}
}

func TestReportCountsTheSpendingSide(t *testing.T) {
	l := &Ledger{Entries: []Entry{
		{Time: at(2), Session: "s1", Group: GroupTreatment, Event: EventInject, Bytes: 100, EstTokens: 25},
		{Time: at(2), Session: "s1", Group: GroupTreatment, Event: EventInject, Bytes: 60, EstTokens: 15},
		{Time: at(2), Session: "s2", Group: GroupControl, Event: EventControl},
		{Time: at(2), Session: "s3", Group: GroupTreatment, Event: EventMiss},
	}}
	r := Build(l, at(1), at(31))
	if r.Injections != 2 || r.Controls != 1 || r.Misses != 1 {
		t.Errorf("injections=%d controls=%d misses=%d", r.Injections, r.Controls, r.Misses)
	}
	if r.InjectedBytes != 160 || r.InjectedTokens != 40 {
		t.Errorf("bytes=%d tokens=%d", r.InjectedBytes, r.InjectedTokens)
	}
}

// TestCostTakesTheHighestReadingNotTheSum: CostUSD is cumulative and the
// agent reports it repeatedly. Summing it would multiply a session's cost by
// how often its status line refreshed.
func TestCostTakesTheHighestReading(t *testing.T) {
	l := &Ledger{Usage: []Usage{
		{Time: at(2), Session: "s1", Group: GroupTreatment, CostUSD: 0.10},
		{Time: at(2), Session: "s1", Group: GroupTreatment, CostUSD: 0.40},
		{Time: at(2), Session: "s1", Group: GroupTreatment, CostUSD: 0.90},
	}}
	r := Build(l, at(1), at(31))
	if r.Cost == nil {
		t.Fatal("no cost comparison")
	}
	if got := r.Cost.Treatment.Total; got != 0.90 {
		t.Errorf("session cost = %v, want the highest reading 0.90", got)
	}
	if r.Cost.Treatment.Sessions != 1 {
		t.Errorf("sessions = %d, want 1", r.Cost.Treatment.Sessions)
	}
}

func TestLoopsCountRepeatedErrors(t *testing.T) {
	l := &Ledger{Entries: []Entry{
		// s1 meets the same error three times: two repeats.
		{Time: at(2), Session: "s1", Group: GroupControl, Event: EventMiss, Fingerprint: "aaaa"},
		{Time: at(2), Session: "s1", Group: GroupControl, Event: EventMiss, Fingerprint: "aaaa"},
		{Time: at(2), Session: "s1", Group: GroupControl, Event: EventMiss, Fingerprint: "aaaa"},
		// s2 meets two different errors once each: no repeats.
		{Time: at(2), Session: "s2", Group: GroupTreatment, Event: EventInject, Fingerprint: "aaaa"},
		{Time: at(2), Session: "s2", Group: GroupTreatment, Event: EventMiss, Fingerprint: "bbbb"},
	}}
	r := Build(l, at(1), at(31))
	if r.Loops == nil {
		t.Fatal("no loop comparison")
	}
	if r.Loops.Control.Total != 2 {
		t.Errorf("control loops = %v, want 2", r.Loops.Control.Total)
	}
	if r.Loops.Treatment.Total != 0 {
		t.Errorf("treatment loops = %v, want 0", r.Loops.Treatment.Total)
	}
}

// TestSmallArmsAreNotCompared: three sessions a side is the floor, and below
// it the report says so rather than quoting an interval nobody should read.
func TestSmallArmsAreNotCompared(t *testing.T) {
	var usage []Usage
	for i := 0; i < 2; i++ {
		usage = append(usage,
			Usage{Time: at(2), Session: fmt.Sprintf("t%d", i), Group: GroupTreatment, CostUSD: 1},
			Usage{Time: at(2), Session: fmt.Sprintf("c%d", i), Group: GroupControl, CostUSD: 2})
	}
	r := Build(&Ledger{Usage: usage}, at(1), at(31))
	if r.Cost.Significant {
		t.Error("a two-session arm was called significant")
	}
	if !strings.Contains(r.Cost.Reason, "not enough sessions") {
		t.Errorf("reason = %q", r.Cost.Reason)
	}
}

// TestNoDifferenceIsNotSignificant: two arms drawn from the same population
// must come back as not meaningful. This is the test that stops the report
// from flattering the tool.
func TestNoDifferenceIsNotSignificant(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	var usage []Usage
	for i := 0; i < 40; i++ {
		usage = append(usage,
			Usage{Time: at(2), Session: fmt.Sprintf("t%d", i), Group: GroupTreatment, CostUSD: rng.Float64()},
			Usage{Time: at(2), Session: fmt.Sprintf("c%d", i), Group: GroupControl, CostUSD: rng.Float64()})
	}
	r := Build(&Ledger{Usage: usage}, at(1), at(31))
	if r.Cost.Significant {
		t.Errorf("identical populations were called significant: diff=%v CI=[%v, %v]",
			r.Cost.Difference, r.Cost.LowCI, r.Cost.HighCI)
	}
	if !strings.Contains(r.Cost.Reason, "covers zero") {
		t.Errorf("reason = %q", r.Cost.Reason)
	}
}

// TestARealDifferenceIsFound is the other half: a clear separation has to be
// detected, or the report would never report anything.
func TestARealDifferenceIsFound(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	var usage []Usage
	for i := 0; i < 40; i++ {
		usage = append(usage,
			Usage{Time: at(2), Session: fmt.Sprintf("t%d", i), Group: GroupTreatment, CostUSD: 1 + rng.Float64()},
			Usage{Time: at(2), Session: fmt.Sprintf("c%d", i), Group: GroupControl, CostUSD: 3 + rng.Float64()})
	}
	r := Build(&Ledger{Usage: usage}, at(1), at(31))
	if !r.Cost.Significant {
		t.Errorf("a two dollar separation was missed: diff=%v CI=[%v, %v] %s",
			r.Cost.Difference, r.Cost.LowCI, r.Cost.HighCI, r.Cost.Reason)
	}
	if r.Cost.Difference < 1.5 || r.Cost.Difference > 2.5 {
		t.Errorf("difference = %v, want about 2", r.Cost.Difference)
	}
}

// TestTheIntervalIsReproducible: a figure that moves between two runs on the
// same data is not one anybody will believe.
func TestTheIntervalIsReproducible(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	var usage []Usage
	for i := 0; i < 20; i++ {
		usage = append(usage,
			Usage{Time: at(2), Session: fmt.Sprintf("t%d", i), Group: GroupTreatment, CostUSD: rng.Float64()},
			Usage{Time: at(2), Session: fmt.Sprintf("c%d", i), Group: GroupControl, CostUSD: 2 * rng.Float64()})
	}
	first := Build(&Ledger{Usage: usage}, at(1), at(31)).Cost
	second := Build(&Ledger{Usage: usage}, at(1), at(31)).Cost
	if first.LowCI != second.LowCI || first.HighCI != second.HighCI {
		t.Errorf("two runs gave [%v, %v] and [%v, %v]", first.LowCI, first.HighCI, second.LowCI, second.HighCI)
	}
}

// TestNoUsageMeansNoCostComparison is the honesty boundary: without the
// agent's usage there is nothing to compare, and the report says nothing
// rather than guessing.
func TestNoUsageMeansNoCostComparison(t *testing.T) {
	l := &Ledger{Entries: []Entry{
		{Time: at(2), Session: "s1", Group: GroupTreatment, Event: EventInject, Bytes: 50, EstTokens: 13},
	}}
	r := Build(l, at(1), at(31))
	if r.Cost != nil {
		t.Errorf("a cost comparison appeared without any usage: %+v", r.Cost)
	}
	if r.Loops == nil {
		t.Error("the loop comparison needs no usage and should still be there")
	}
	if r.InjectedBytes != 50 {
		t.Errorf("the spending side was lost: %d", r.InjectedBytes)
	}
}

func TestMedianAndPercentile(t *testing.T) {
	if got := median([]float64{3, 1, 2}); got != 2 {
		t.Errorf("odd median = %v, want 2", got)
	}
	if got := median([]float64{4, 1, 3, 2}); got != 2.5 {
		t.Errorf("even median = %v, want 2.5", got)
	}
	if got := median(nil); got != 0 {
		t.Errorf("empty median = %v, want 0", got)
	}

	// median must not reorder the caller's slice.
	xs := []float64{3, 1, 2}
	median(xs)
	if xs[0] != 3 {
		t.Errorf("median sorted its argument in place: %v", xs)
	}

	sorted := []float64{0, 1, 2, 3, 4}
	if got := percentile(sorted, 50); got != 2 {
		t.Errorf("p50 = %v, want 2", got)
	}
	if got := percentile(sorted, 25); got != 1 {
		t.Errorf("p25 = %v, want 1", got)
	}
	if got := percentile([]float64{7}, 95); got != 7 {
		t.Errorf("single value = %v, want 7", got)
	}
}

// TestUnattributedLinesDoNotInventLoops: a line written without a session id
// still counts as spending, but lumping every one of them under the empty
// session would make unrelated invocations look like one looping run.
func TestUnattributedLinesDoNotInventLoops(t *testing.T) {
	l := &Ledger{
		Entries: []Entry{
			{Time: at(2), Session: "", Event: EventInject, Fingerprint: "aaaa", Bytes: 10, EstTokens: 3},
			{Time: at(2), Session: "", Event: EventInject, Fingerprint: "aaaa", Bytes: 10, EstTokens: 3},
			{Time: at(2), Session: "", Event: EventInject, Fingerprint: "aaaa", Bytes: 10, EstTokens: 3},
		},
		Usage: []Usage{{Time: at(2), Session: "", CostUSD: 5}},
	}
	r := Build(l, at(1), at(31))
	if r.InjectedBytes != 30 {
		t.Errorf("the spending side lost the unattributed lines: %d", r.InjectedBytes)
	}
	if r.Loops != nil {
		t.Errorf("loops were invented from unattributed lines: %+v", r.Loops)
	}
	if r.Cost != nil {
		t.Errorf("a cost arm was built from an unattributed session: %+v", r.Cost)
	}
}
