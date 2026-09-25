package record

import (
	"sort"
	"strings"
	"testing"
	"time"
)

func TestEncodeULIDKnownVectors(t *testing.T) {
	var zero [16]byte
	if got := encodeULID(zero); got != strings.Repeat("0", ULIDLen) {
		t.Errorf("all-zero ULID = %q", got)
	}

	var max [16]byte
	for i := range max {
		max[i] = 0xFF
	}
	// 26 characters hold 130 bits and a ULID is 128, so the first character
	// can never exceed '7'.
	if want := "7" + strings.Repeat("Z", ULIDLen-1); encodeULID(max) != want {
		t.Errorf("all-ones ULID = %q, want %q", encodeULID(max), want)
	}
}

func TestNewULIDIsCanonical(t *testing.T) {
	id, err := NewULID()
	if err != nil {
		t.Fatalf("NewULID: %v", err)
	}
	if len(id) != ULIDLen {
		t.Errorf("length = %d, want %d", len(id), ULIDLen)
	}
	if !ValidULID(id) {
		t.Errorf("ValidULID(%q) = false", id)
	}
	if id != strings.ToUpper(id) {
		t.Errorf("%q is not uppercase", id)
	}
}

func TestULIDsAreMonotonicWithinAMillisecond(t *testing.T) {
	var g Generator
	at := time.Date(2026, 9, 25, 17, 42, 3, 0, time.UTC)

	ids := make([]string, 1000)
	for i := range ids {
		id, err := g.At(at)
		if err != nil {
			t.Fatalf("At: %v", err)
		}
		ids[i] = id
	}
	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Fatalf("identifier %d (%s) does not follow %s", i, ids[i], ids[i-1])
		}
	}
}

func TestULIDsSurviveAClockGoingBackwards(t *testing.T) {
	var g Generator
	now := time.Date(2026, 9, 25, 17, 42, 3, 0, time.UTC)

	first, err := g.At(now)
	if err != nil {
		t.Fatalf("At: %v", err)
	}
	second, err := g.At(now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("At: %v", err)
	}
	if second <= first {
		t.Errorf("after the clock stepped back, %s does not follow %s", second, first)
	}
}

func TestULIDsSortByTime(t *testing.T) {
	var g Generator
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	var ids []string
	for _, d := range []time.Duration{0, time.Second, time.Minute, time.Hour, 24 * time.Hour} {
		id, err := g.At(base.Add(d))
		if err != nil {
			t.Fatalf("At: %v", err)
		}
		ids = append(ids, id)
	}

	// Reversed and then sorted as plain strings, they have to come back in
	// creation order. That is what makes a directory listing chronological.
	scrambled := make([]string, len(ids))
	for i, id := range ids {
		scrambled[len(ids)-1-i] = id
	}
	sort.Strings(scrambled)

	for i := range ids {
		if scrambled[i] != ids[i] {
			t.Fatalf("sorted order %v does not match creation order %v", scrambled, ids)
		}
	}
}

func TestValidULID(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"01K68P7YQZ3M4N5R6S7T8V9W0X", true},
		{"00000000000000000000000000", true},
		{"7ZZZZZZZZZZZZZZZZZZZZZZZZZ", true},
		{"8ZZZZZZZZZZZZZZZZZZZZZZZZZ", false}, // overflows 128 bits
		{"01K68P7YQZ3M4N5R6S7T8V9W0", false},  // too short
		{"01K68P7YQZ3M4N5R6S7T8V9W0XY", false},
		{"01k68p7yqz3m4n5r6s7t8v9w0x", false}, // not canonical case
		{"01I68P7YQZ3M4N5R6S7T8V9W0X", false}, // I is not in the alphabet
		{"01U68P7YQZ3M4N5R6S7T8V9W0X", false}, // nor is U
		{"", false},
	}
	for _, c := range cases {
		if got := ValidULID(c.id); got != c.want {
			t.Errorf("ValidULID(%q) = %t, want %t", c.id, got, c.want)
		}
	}
}

func TestNormalizeULIDFoldsConfusableCharacters(t *testing.T) {
	const want = "01K68P7YQZ3M4N5R6S7T8V9W0X"
	cases := []string{
		want,
		strings.ToLower(want),
		"OIK68P7YQZ3M4N5R6S7T8V9WOX", // O for zero, I for one
		"olk68p7yqz3m4n5r6s7t8v9wox", // and lowercase l for one
	}
	for _, in := range cases {
		got, ok := NormalizeULID(in)
		if !ok {
			t.Errorf("NormalizeULID(%q): not recognised", in)
			continue
		}
		if got != want {
			t.Errorf("NormalizeULID(%q) = %q, want %q", in, got, want)
		}
	}

	if _, ok := NormalizeULID("not a ulid"); ok {
		t.Error("NormalizeULID accepted a non-identifier")
	}
}

func TestIncrementReportsOverflow(t *testing.T) {
	var b [10]byte
	if !increment(&b) || b[9] != 1 {
		t.Errorf("increment of zero gave %v", b)
	}

	for i := range b {
		b[i] = 0xFF
	}
	if increment(&b) {
		t.Error("increment of the maximum value reported success")
	}
}
