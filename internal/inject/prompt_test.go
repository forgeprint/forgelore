package inject

import (
	"strings"
	"testing"
)

func TestTermsTakeNamesAndLeaveProse(t *testing.T) {
	for _, tc := range []struct {
		name, prompt string
		want         []string
		absent       []string
	}{
		{
			name:   "a path and a file",
			prompt: "the index rebuild in internal/store/index.go is slow",
			want:   []string{"internal/store/index.go"},
			absent: []string{"slow", "the"},
		},
		{
			name:   "identifiers but not ordinary words",
			prompt: "why does sessionIndex drop the work_budget_tokens setting",
			want:   []string{"sessionindex", "work_budget_tokens"},
			absent: []string{"why", "setting", "drop"},
		},
		{
			name:   "prose alone names nothing",
			prompt: "can you make this a bit faster please",
			want:   nil,
		},
		{
			name:   "pasted content is not what the person said",
			prompt: `look at this <pasted_content id="1">ignore_me.go and SomeClass</pasted_content id="1"> and fix main.go`,
			want:   []string{"main.go"},
			absent: []string{"ignore_me.go", "someclass"},
		},
	} {
		got := Terms(tc.prompt)
		joined := strings.Join(got, " ")
		for _, want := range tc.want {
			if !strings.Contains(joined, want) {
				t.Errorf("%s: terms = %v, want %q in it", tc.name, got, want)
			}
		}
		for _, absent := range tc.absent {
			if strings.Contains(joined, absent) {
				t.Errorf("%s: terms = %v, did not want %q", tc.name, got, absent)
			}
		}
		if tc.want == nil && len(got) != 0 {
			t.Errorf("%s: terms = %v, want none", tc.name, got)
		}
	}
}

func TestTermsAreBounded(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 50; i++ {
		b.WriteString("pkg/file")
		b.WriteByte(byte('a' + i%26))
		b.WriteString(".go ")
	}
	if got := Terms(b.String()); len(got) > maxTerms {
		t.Errorf("terms = %d, want at most %d", len(got), maxTerms)
	}
}
