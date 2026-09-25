// Package record defines the on-disk record format: the restricted YAML
// dialect of ADR-0017, the record fields themselves, and ULID generation.
//
// docs/record-format.md is the normative description. Where this package and
// that document disagree, the document is the bug report.
package record

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// CurrentSchema is the schema version this package writes. A record from a
// later version is read as far as its known fields allow rather than rejected
// (ADR-0011).
const CurrentSchema = 1

// Record types (ADR-0016). A type outside this set is preserved and behaves
// like a note: searchable, never injected.
const (
	TypeFix      = "fix"
	TypeDeadEnd  = "dead_end"
	TypeDecision = "decision"
	TypeCommand  = "command"
	TypeNote     = "note"
)

// Scope is where a record lives. The directory decides it; the field is
// validated against the directory (ADR-0010).
type Scope string

const (
	ScopeTeam  Scope = "team"
	ScopeLocal Scope = "local"
)

// How a record came to exist.
const (
	SourceUser   = "user"
	SourceHook   = "hook"
	SourceImport = "import"
)

// timeLayout is ISO-8601 UTC at second precision. Parsing with this layout
// rejects offsets, which is deliberate: every record is stored in UTC.
const timeLayout = "2006-01-02T15:04:05Z"

const delimiter = "---"

// fieldOrder is the canonical order the writer emits. Unknown fields follow,
// in the order they were read.
var fieldOrder = []string{
	"schema", "id", "type", "scope", "title", "created",
	"source", "tainted", "fingerprint", "tags", "related", "superseded_by",
}

// Record is one memory: metadata that decides when it is injected, and a
// markdown body that is only ever fetched on request.
type Record struct {
	Schema       int
	ID           string
	Type         string
	Scope        Scope
	Title        string
	Created      time.Time
	Source       string
	Tainted      bool
	Fingerprint  string
	Tags         []string
	Related      []string
	SupersededBy string

	// Body is everything after the closing delimiter, verbatim.
	Body string

	// Unknown holds fields this version does not recognise. They are written
	// back exactly as they were read (ADR-0011).
	Unknown []Field
}

// ValidationError names the field that is wrong and why.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("field %q: %s", e.Field, e.Msg)
}

// InSessionIndex reports whether this record's title belongs in the budgeted
// session-start index (ADR-0016).
func (r *Record) InSessionIndex() bool {
	if r.SupersededBy != "" {
		return false
	}
	return r.Type == TypeCommand || r.Type == TypeDecision
}

// InjectableOnMatch reports whether this record may be injected when its
// fingerprint matches a failing command (ADR-0016).
func (r *Record) InjectableOnMatch() bool {
	if r.SupersededBy != "" {
		return false
	}
	return r.Type == TypeFix || r.Type == TypeDeadEnd
}

// Decode reads a record file. It is strict: anything outside the format is an
// error, and the caller skips the file rather than repairing it (ADR-0007).
func Decode(data []byte) (*Record, error) {
	front, body, err := splitFile(data)
	if err != nil {
		return nil, err
	}
	doc, err := ParseDoc(front, 2)
	if err != nil {
		return nil, err
	}

	r := &Record{Body: body}
	for _, f := range doc.Fields {
		if err := r.setField(f); err != nil {
			return nil, err
		}
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return r, nil
}

// splitFile separates the frontmatter lines from the body. The body is kept
// verbatim so that encoding returns it untouched.
func splitFile(data []byte) ([]string, string, error) {
	text := string(bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")))
	rest, ok := strings.CutPrefix(text, delimiter+"\n")
	if !ok {
		return nil, "", &SyntaxError{1, "file does not start with a --- frontmatter delimiter"}
	}
	var front []string
	for line := 2; ; line++ {
		next, more, ok := strings.Cut(rest, "\n")
		if !ok {
			return nil, "", &SyntaxError{line, "frontmatter is not closed with ---"}
		}
		if strings.TrimRight(next, " ") == delimiter {
			return front, more, nil
		}
		front = append(front, next)
		rest = more
	}
}

func (r *Record) setField(f Field) error {
	str := func() (string, error) {
		if f.Value.Kind != KindString {
			return "", &ValidationError{f.Key, "expected a string"}
		}
		return f.Value.Str, nil
	}

	var err error
	switch f.Key {
	case "schema":
		if f.Value.Kind != KindInt {
			return &ValidationError{f.Key, "expected an integer"}
		}
		r.Schema = int(f.Value.Int)
	case "id":
		r.ID, err = str()
	case "type":
		r.Type, err = str()
	case "scope":
		var s string
		if s, err = str(); err == nil {
			r.Scope = Scope(s)
		}
	case "title":
		r.Title, err = str()
	case "created":
		var s string
		if s, err = str(); err != nil {
			return err
		}
		t, perr := time.Parse(timeLayout, s)
		if perr != nil {
			return &ValidationError{f.Key, "expected an ISO-8601 UTC timestamp like 2026-09-25T17:42:03Z"}
		}
		r.Created = t
	case "source":
		r.Source, err = str()
	case "tainted":
		if f.Value.Kind != KindBool {
			return &ValidationError{f.Key, "expected true or false"}
		}
		r.Tainted = f.Value.Bool
	case "fingerprint":
		r.Fingerprint, err = str()
	case "tags", "related":
		if f.Value.Kind != KindList {
			return &ValidationError{f.Key, "expected a list"}
		}
		if f.Key == "tags" {
			r.Tags = f.Value.List
		} else {
			r.Related = f.Value.List
		}
	case "superseded_by":
		r.SupersededBy, err = str()
	default:
		r.Unknown = append(r.Unknown, f)
	}
	return err
}

// Validate checks everything docs/record-format.md requires.
func (r *Record) Validate() error {
	if r.Schema < 1 {
		return &ValidationError{"schema", "missing or not a positive integer"}
	}
	if !ValidULID(r.ID) {
		return &ValidationError{"id", "not a canonical uppercase ULID"}
	}
	if r.Type == "" {
		return &ValidationError{"type", "missing"}
	}
	if !bareToken(r.Type) {
		return &ValidationError{"type", "must be a plain token"}
	}
	switch r.Scope {
	case ScopeTeam, ScopeLocal:
	default:
		return &ValidationError{"scope", "must be team or local"}
	}
	if strings.TrimSpace(r.Title) == "" {
		return &ValidationError{"title", "missing"}
	}
	if r.Created.IsZero() {
		return &ValidationError{"created", "missing"}
	}
	switch r.Source {
	case SourceUser, SourceHook, SourceImport:
	default:
		return &ValidationError{"source", "must be user, hook or import"}
	}
	switch r.Type {
	case TypeFix, TypeDeadEnd:
		if r.Fingerprint == "" {
			return &ValidationError{"fingerprint", "required for fix and dead_end records"}
		}
	}
	if r.Fingerprint != "" && !lowerHex(r.Fingerprint) {
		return &ValidationError{"fingerprint", "must be lowercase hexadecimal"}
	}
	for _, tag := range r.Tags {
		if !kebab(tag) {
			return &ValidationError{"tags", fmt.Sprintf("%q is not lowercase kebab-case", tag)}
		}
	}
	for _, id := range r.Related {
		if !ValidULID(id) {
			return &ValidationError{"related", fmt.Sprintf("%q is not a canonical ULID", id)}
		}
	}
	if r.SupersededBy != "" && !ValidULID(r.SupersededBy) {
		return &ValidationError{"superseded_by", "not a canonical ULID"}
	}
	return nil
}

// Encode writes the record canonically. Encoding a decoded record returns the
// same bytes, which is the property the round-trip test pins down.
func (r *Record) Encode() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}

	var b strings.Builder
	b.WriteString(delimiter + "\n")
	for _, key := range fieldOrder {
		if line, ok := r.encodeField(key); ok {
			b.WriteString(line + "\n")
		}
	}
	for _, f := range r.Unknown {
		for _, raw := range f.Raw {
			b.WriteString(raw + "\n")
		}
	}
	b.WriteString(delimiter + "\n")

	body := r.Body
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	b.WriteString(body)
	return []byte(b.String()), nil
}

// encodeField returns the line for a known key, or false when the field is
// optional and empty. An empty list is never written as [].
func (r *Record) encodeField(key string) (string, bool) {
	switch key {
	case "schema":
		return fmt.Sprintf("schema: %d", r.Schema), true
	case "id":
		return "id: " + r.ID, true
	case "type":
		return "type: " + r.Type, true
	case "scope":
		return "scope: " + string(r.Scope), true
	case "title":
		// Always quoted: the title is free text, and quoting it unconditionally
		// means its rendering never depends on what it happens to contain.
		return "title: " + quote(r.Title), true
	case "created":
		return "created: " + r.Created.UTC().Format(timeLayout), true
	case "source":
		return "source: " + r.Source, true
	case "tainted":
		return fmt.Sprintf("tainted: %t", r.Tainted), true
	case "fingerprint":
		if r.Fingerprint == "" {
			return "", false
		}
		return "fingerprint: " + r.Fingerprint, true
	case "tags":
		if len(r.Tags) == 0 {
			return "", false
		}
		return "tags: " + encodeList(r.Tags), true
	case "related":
		if len(r.Related) == 0 {
			return "", false
		}
		return "related: " + encodeList(r.Related), true
	case "superseded_by":
		if r.SupersededBy == "" {
			return "", false
		}
		return "superseded_by: " + r.SupersededBy, true
	}
	return "", false
}

func lowerHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return s != ""
}

// kebab reports whether s is lowercase alphanumeric words joined by hyphens.
func kebab(s string) bool {
	if s == "" {
		return false
	}
	for _, part := range strings.Split(s, "-") {
		if part == "" {
			return false
		}
		for i := 0; i < len(part); i++ {
			c := part[i]
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
				return false
			}
		}
	}
	return true
}
