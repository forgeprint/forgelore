package record

import (
	"strings"
	"testing"
)

func TestParseDocAcceptsTheDialect(t *testing.T) {
	lines := strings.Split(strings.TrimPrefix(`
# a full-line comment
name: plain
quoted: "with spaces and a # hash"
number: 42
negative: -7
yes: true
no: false
flow: [a, b, c]
block:
  - one
  - two
empty: []
trailing: value # comment after a bare scalar
escaped: "a \"quoted\" word and a \\ backslash"
`, "\n"), "\n")

	doc, err := ParseDoc(lines, 1)
	if err != nil {
		t.Fatalf("ParseDoc: %v", err)
	}

	want := []struct {
		key  string
		kind Kind
		str  string
		num  int64
		b    bool
		list []string
	}{
		{key: "name", kind: KindString, str: "plain"},
		{key: "quoted", kind: KindString, str: "with spaces and a # hash"},
		{key: "number", kind: KindInt, num: 42},
		{key: "negative", kind: KindInt, num: -7},
		{key: "yes", kind: KindBool, b: true},
		{key: "no", kind: KindBool, b: false},
		{key: "flow", kind: KindList, list: []string{"a", "b", "c"}},
		{key: "block", kind: KindList, list: []string{"one", "two"}},
		{key: "empty", kind: KindList},
		{key: "trailing", kind: KindString, str: "value"},
		{key: "escaped", kind: KindString, str: `a "quoted" word and a \ backslash`},
	}

	if len(doc.Fields) != len(want) {
		t.Fatalf("parsed %d fields, want %d", len(doc.Fields), len(want))
	}
	for i, w := range want {
		got := doc.Fields[i]
		if got.Key != w.key {
			t.Errorf("field %d: key = %q, want %q", i, got.Key, w.key)
			continue
		}
		if got.Value.Kind != w.kind {
			t.Errorf("%s: kind = %v, want %v", w.key, got.Value.Kind, w.kind)
			continue
		}
		switch w.kind {
		case KindString:
			if got.Value.Str != w.str {
				t.Errorf("%s = %q, want %q", w.key, got.Value.Str, w.str)
			}
		case KindInt:
			if got.Value.Int != w.num {
				t.Errorf("%s = %d, want %d", w.key, got.Value.Int, w.num)
			}
		case KindBool:
			if got.Value.Bool != w.b {
				t.Errorf("%s = %t, want %t", w.key, got.Value.Bool, w.b)
			}
		case KindList:
			if strings.Join(got.Value.List, ",") != strings.Join(w.list, ",") {
				t.Errorf("%s = %v, want %v", w.key, got.Value.List, w.list)
			}
		}
	}
}

func TestParseDocRejectsWhatIsOutsideTheDialect(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantMsg string
	}{
		{"nested map", "outer:\n  inner: 1", "nested mappings"},
		{"tab", "key:\tvalue", "tab character"},
		{"anchor", "key: &anchor value", "anchors and aliases"},
		{"alias", "key: *anchor", "anchors and aliases"},
		{"literal block", "key: |", "multi-line strings"},
		{"folded block", "key: >", "multi-line strings"},
		{"type tag", "key: !!str value", "type tags"},
		{"inline map", "key: {a: 1}", "inline maps"},
		{"single quotes", "key: 'value'", "single-quoted"},
		{"duplicate key", "key: a\nkey: b", "duplicate key"},
		{"orphan list item", "- a", "list item without a key"},
		{"no colon", "just a line", "expected 'key: value'"},
		{"unterminated string", `key: "abc`, "unterminated quoted string"},
		{"unknown escape", `key: "a \n b"`, "unsupported escape"},
		{"empty key", ": value", "empty key"},
		{"nested list", "key: [[a]]", "nested collections"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseDoc(strings.Split(c.input, "\n"), 1)
			if err == nil {
				t.Fatalf("ParseDoc(%q): got nil error, want one mentioning %q", c.input, c.wantMsg)
			}
			if !strings.Contains(err.Error(), c.wantMsg) {
				t.Errorf("ParseDoc(%q) error = %q, want it to mention %q", c.input, err, c.wantMsg)
			}
		})
	}
}

func TestSyntaxErrorPointsAtTheFileLine(t *testing.T) {
	// firstLine is the file line of lines[0]; frontmatter starts at line 2.
	_, err := ParseDoc([]string{"ok: 1", "bad: |"}, 2)
	if err == nil {
		t.Fatal("got nil error, want a syntax error")
	}
	se, ok := err.(*SyntaxError)
	if !ok {
		t.Fatalf("error type = %T, want *SyntaxError", err)
	}
	if se.Line != 3 {
		t.Errorf("Line = %d, want 3", se.Line)
	}
}

func TestDottedKeysAreAccepted(t *testing.T) {
	// Config groups settings with dotted keys rather than nesting (ADR-0018).
	doc, err := ParseDoc([]string{"inject.budget_tokens: 400"}, 1)
	if err != nil {
		t.Fatalf("ParseDoc: %v", err)
	}
	f, ok := doc.Get("inject.budget_tokens")
	if !ok {
		t.Fatal("dotted key not found")
	}
	if f.Value.Int != 400 {
		t.Errorf("value = %d, want 400", f.Value.Int)
	}
}

func TestCanonicalEncoding(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"plain", "plain"},
		{"01K68P7YQZ3M4N5R6S7T8V9W0X", "01K68P7YQZ3M4N5R6S7T8V9W0X"},
		{"2026-09-25T17:42:03Z", "2026-09-25T17:42:03Z"},
		{"with space", `"with space"`},
		{"", `""`},
		{"has # hash", `"has # hash"`},
		{`a "quote"`, `"a \"quote\""`},
		{`back\slash`, `"back\\slash"`},
	}
	for _, c := range cases {
		if got := encodeScalar(c.in); got != c.want {
			t.Errorf("encodeScalar(%q) = %s, want %s", c.in, got, c.want)
		}
	}

	if got := encodeList([]string{"go", "cross compile"}); got != `[go, "cross compile"]` {
		t.Errorf("encodeList = %s", got)
	}
}

func TestQuotedValuesSurviveAReparse(t *testing.T) {
	// Anything the writer emits has to be readable again, including the
	// characters that forced the quoting in the first place.
	for _, want := range []string{"plain", "with space", `a "quote"`, `back\slash`, "trailing # hash", ""} {
		line := "key: " + encodeScalar(want)
		doc, err := ParseDoc([]string{line}, 1)
		if err != nil {
			t.Fatalf("ParseDoc(%q): %v", line, err)
		}
		f, _ := doc.Get("key")
		if f.Value.Str != want {
			t.Errorf("round trip of %q gave %q (line was %q)", want, f.Value.Str, line)
		}
	}
}
