package record

import (
	"strings"
	"testing"
	"time"
)

// canonical is a record exactly as the writer emits it, including two fields
// this version does not know about.
const canonical = `---
schema: 1
id: 01K68P7YQZ3M4N5R6S7T8V9W0X
type: fix
scope: team
title: "Set CGO_ENABLED=0 before cross-building; the arm64 link step fails otherwise"
created: 2026-09-25T17:42:03Z
source: hook
tainted: false
fingerprint: 9f2c4a1e7b3d0856
tags: [go, build, cross-compile]
related: [01K68P9AB2C3D4E5F6G7H8J9K0]
weight: 3
observed_on:
  - windows
  - linux
---

## Symptom

` + "`GOARCH=arm64 go build ./...`" + ` fails in the link step.
`

func TestDecodeReadsEveryField(t *testing.T) {
	r, err := Decode([]byte(canonical))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if r.Schema != 1 {
		t.Errorf("Schema = %d, want 1", r.Schema)
	}
	if r.ID != "01K68P7YQZ3M4N5R6S7T8V9W0X" {
		t.Errorf("ID = %q", r.ID)
	}
	if r.Type != TypeFix {
		t.Errorf("Type = %q, want %q", r.Type, TypeFix)
	}
	if r.Scope != ScopeTeam {
		t.Errorf("Scope = %q, want %q", r.Scope, ScopeTeam)
	}
	if !strings.HasPrefix(r.Title, "Set CGO_ENABLED=0") {
		t.Errorf("Title = %q", r.Title)
	}
	if want := time.Date(2026, 9, 25, 17, 42, 3, 0, time.UTC); !r.Created.Equal(want) {
		t.Errorf("Created = %v, want %v", r.Created, want)
	}
	if r.Source != SourceHook {
		t.Errorf("Source = %q, want %q", r.Source, SourceHook)
	}
	if r.Tainted {
		t.Error("Tainted = true, want false")
	}
	if r.Fingerprint != "9f2c4a1e7b3d0856" {
		t.Errorf("Fingerprint = %q", r.Fingerprint)
	}
	if got := strings.Join(r.Tags, ","); got != "go,build,cross-compile" {
		t.Errorf("Tags = %q", got)
	}
	if len(r.Related) != 1 || r.Related[0] != "01K68P9AB2C3D4E5F6G7H8J9K0" {
		t.Errorf("Related = %v", r.Related)
	}
	if len(r.Unknown) != 2 {
		t.Fatalf("Unknown holds %d fields, want 2", len(r.Unknown))
	}
	if !strings.Contains(r.Body, "fails in the link step") {
		t.Errorf("Body = %q", r.Body)
	}
}

func TestRoundTripIsByteIdentical(t *testing.T) {
	// The property the whole format rests on: a canonical record read and
	// written again is the same bytes, including fields this version has never
	// heard of (ADR-0011, ADR-0017).
	r, err := Decode([]byte(canonical))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	out, err := r.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if string(out) != canonical {
		t.Errorf("round trip changed the file.\n--- got ---\n%s\n--- want ---\n%s", out, canonical)
	}
}

func TestEncodingIsIdempotentForNonCanonicalInput(t *testing.T) {
	// Input written by hand may not be canonical. The first write normalises
	// it; every write after that changes nothing.
	const handWritten = `---
schema: 1
id: 01K68P7YQZ3M4N5R6S7T8V9W0X
type: note
scope: local
title: unquoted title
created: 2026-09-25T17:42:03Z
source: user
tainted: false
---
body
`
	r, err := Decode([]byte(handWritten))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	once, err := r.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if string(once) == handWritten {
		t.Fatal("expected the writer to normalise the unquoted title")
	}

	again, err := Decode(once)
	if err != nil {
		t.Fatalf("Decode of normalised record: %v", err)
	}
	twice, err := again.Encode()
	if err != nil {
		t.Fatalf("Encode of normalised record: %v", err)
	}
	if string(twice) != string(once) {
		t.Errorf("second write differs from the first.\n--- first ---\n%s\n--- second ---\n%s", once, twice)
	}
}

func TestEncodeOmitsEmptyOptionalFields(t *testing.T) {
	r := validRecord()
	r.Type = TypeNote
	r.Fingerprint = ""
	out, err := r.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	for _, key := range []string{"fingerprint:", "tags:", "related:", "superseded_by:", "[]"} {
		if strings.Contains(string(out), key) {
			t.Errorf("output contains %q, want the field omitted:\n%s", key, out)
		}
	}
}

func TestValidateRejects(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*Record)
		wantMsg string
	}{
		{"no schema", func(r *Record) { r.Schema = 0 }, "schema"},
		{"bad id", func(r *Record) { r.ID = "not-a-ulid" }, "id"},
		{"lowercase id", func(r *Record) { r.ID = strings.ToLower(r.ID) }, "id"},
		{"no type", func(r *Record) { r.Type = "" }, "type"},
		{"no title", func(r *Record) { r.Title = "   " }, "title"},
		{"bad scope", func(r *Record) { r.Scope = "everyone" }, "scope"},
		{"bad source", func(r *Record) { r.Source = "telepathy" }, "source"},
		{"no created", func(r *Record) { r.Created = time.Time{} }, "created"},
		{"fix without fingerprint", func(r *Record) { r.Type = TypeFix; r.Fingerprint = "" }, "fingerprint"},
		{"uppercase fingerprint", func(r *Record) { r.Fingerprint = "9F2C" }, "fingerprint"},
		{"shouting tag", func(r *Record) { r.Tags = []string{"Go"} }, "tags"},
		{"bad related id", func(r *Record) { r.Related = []string{"nope"} }, "related"},
		{"bad superseded_by", func(r *Record) { r.SupersededBy = "nope" }, "superseded_by"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := validRecord()
			c.mutate(r)
			err := r.Validate()
			if err == nil {
				t.Fatalf("Validate: got nil error, want one about %q", c.wantMsg)
			}
			if !strings.Contains(err.Error(), c.wantMsg) {
				t.Errorf("Validate error = %q, want it to name %q", err, c.wantMsg)
			}
		})
	}
}

func TestDecodeRejectsMalformedFiles(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"no frontmatter", "just a body\n"},
		{"unclosed frontmatter", "---\nschema: 1\n"},
		{"missing required field", "---\nschema: 1\nid: 01K68P7YQZ3M4N5R6S7T8V9W0X\n---\n"},
		{"wrong type for schema", "---\nschema: one\n---\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Decode([]byte(c.input)); err == nil {
				t.Error("Decode: got nil error, want one")
			}
		})
	}
}

func TestInjectionBehaviourByType(t *testing.T) {
	// The table in ADR-0016, as executable form.
	cases := []struct {
		typ        string
		superseded bool
		index      bool
		onMatch    bool
	}{
		{typ: TypeFix, onMatch: true},
		{typ: TypeDeadEnd, onMatch: true},
		{typ: TypeCommand, index: true},
		{typ: TypeDecision, index: true},
		{typ: TypeNote},
		{typ: "from_a_later_version"},
		{typ: TypeFix, superseded: true},
		{typ: TypeCommand, superseded: true},
	}
	for _, c := range cases {
		r := &Record{Type: c.typ}
		if c.superseded {
			r.SupersededBy = "01K68P9AB2C3D4E5F6G7H8J9K0"
		}
		if got := r.InSessionIndex(); got != c.index {
			t.Errorf("type %q superseded=%t: InSessionIndex = %t, want %t", c.typ, c.superseded, got, c.index)
		}
		if got := r.InjectableOnMatch(); got != c.onMatch {
			t.Errorf("type %q superseded=%t: InjectableOnMatch = %t, want %t", c.typ, c.superseded, got, c.onMatch)
		}
	}
}

func TestUnknownTypeSurvivesARoundTrip(t *testing.T) {
	r := validRecord()
	r.Type = "from_a_later_version"
	out, err := r.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	back, err := Decode(out)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if back.Type != "from_a_later_version" {
		t.Errorf("Type = %q, want it preserved", back.Type)
	}
	if back.InSessionIndex() || back.InjectableOnMatch() {
		t.Error("an unrecognised type must behave like a note: never injected")
	}
}

func validRecord() *Record {
	return &Record{
		Schema:      CurrentSchema,
		ID:          "01K68P7YQZ3M4N5R6S7T8V9W0X",
		Type:        TypeFix,
		Scope:       ScopeTeam,
		Title:       "Set CGO_ENABLED=0 before cross-building",
		Created:     time.Date(2026, 9, 25, 17, 42, 3, 0, time.UTC),
		Source:      SourceHook,
		Fingerprint: "9f2c4a1e7b3d0856",
		Body:        "\nbody\n",
	}
}
