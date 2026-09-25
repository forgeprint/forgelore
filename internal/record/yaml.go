package record

import (
	"fmt"
	"strconv"
	"strings"
)

// The restricted YAML dialect of ADR-0017, used for record frontmatter and for
// config files. Supported: flat keys, strings, integers, booleans, string
// lists in flow and block form, and full-line or trailing comments. Not
// supported: nested maps, anchors, aliases, multi-line string styles, tabs.
//
// Reading is strict. A construct outside the dialect is an error naming the
// line and the rule, never a guess.

// Kind is the type of a parsed value.
type Kind int

const (
	KindString Kind = iota
	KindInt
	KindBool
	KindList
)

// Value is one value in the dialect. Nothing nests, so a flat struct is enough.
type Value struct {
	Kind Kind
	Str  string
	Int  int64
	Bool bool
	List []string
}

// Field is a key with its value. Raw holds the lines the field was read from,
// so a field this version does not recognise can be written back exactly as it
// arrived (ADR-0011).
type Field struct {
	Key   string
	Value Value
	Raw   []string
}

// Doc is a parsed block of fields, in the order they appeared.
type Doc struct {
	Fields []Field
}

// Get returns the field with the given key.
func (d *Doc) Get(key string) (Field, bool) {
	for _, f := range d.Fields {
		if f.Key == key {
			return f, true
		}
	}
	return Field{}, false
}

// SyntaxError reports the line and the rule that was broken. doctor prints
// these as they are, so the message has to name the unsupported construct
// rather than only saying that parsing failed.
type SyntaxError struct {
	Line int
	Msg  string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}

// ParseDoc parses lines of the dialect. firstLine is the file line number of
// lines[0], so errors point into the file rather than into the slice.
func ParseDoc(lines []string, firstLine int) (*Doc, error) {
	doc := &Doc{}
	seen := make(map[string]bool, len(lines))
	// listIdx is the index of the field currently collecting block-form items,
	// or -1 when no key is open.
	listIdx := -1

	for i, raw := range lines {
		n := firstLine + i

		if strings.ContainsRune(raw, '\t') {
			return nil, &SyntaxError{n, "tab character; this dialect indents with spaces only"}
		}

		line := strings.TrimRight(raw, " ")
		body := strings.TrimLeft(line, " ")
		indent := len(line) - len(body)

		switch {
		case body == "":
			listIdx = -1

		case strings.HasPrefix(body, "#"):
			// Full-line comment.

		case strings.HasPrefix(body, "- "), body == "-":
			if listIdx < 0 {
				return nil, &SyntaxError{n, "list item without a key above it"}
			}
			item, err := parseListItem(strings.TrimPrefix(body, "-"), n)
			if err != nil {
				return nil, err
			}
			f := &doc.Fields[listIdx]
			f.Value.List = append(f.Value.List, item)
			f.Raw = append(f.Raw, raw)

		default:
			if indent > 0 {
				return nil, &SyntaxError{n, "indented key; nested mappings are not supported"}
			}
			key, rest, ok := strings.Cut(body, ":")
			if !ok {
				return nil, &SyntaxError{n, "expected 'key: value'"}
			}
			if err := checkKey(key, n); err != nil {
				return nil, err
			}
			if seen[key] {
				return nil, &SyntaxError{n, fmt.Sprintf("duplicate key %q", key)}
			}
			seen[key] = true

			rest = strings.TrimLeft(rest, " ")
			field := Field{Key: key, Raw: []string{raw}}

			if rest == "" || isComment(rest) {
				// A key with no inline value opens a block list. With no items
				// below it, it stays an empty list.
				field.Value = Value{Kind: KindList}
				doc.Fields = append(doc.Fields, field)
				listIdx = len(doc.Fields) - 1
				continue
			}

			v, err := parseValue(rest, n)
			if err != nil {
				return nil, err
			}
			field.Value = v
			doc.Fields = append(doc.Fields, field)
			listIdx = -1
		}
	}
	return doc, nil
}

func isComment(s string) bool { return strings.HasPrefix(s, "#") }

// checkKey accepts dot-separated segments of letters, digits and underscores,
// which covers record fields and the dotted config keys of ADR-0018.
func checkKey(key string, line int) error {
	if key == "" {
		return &SyntaxError{line, "empty key"}
	}
	for _, seg := range strings.Split(key, ".") {
		if seg == "" {
			return &SyntaxError{line, fmt.Sprintf("malformed key %q", key)}
		}
		for i, r := range seg {
			switch {
			case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_':
			case r >= '0' && r <= '9' && i > 0:
			default:
				return &SyntaxError{line, fmt.Sprintf("malformed key %q", key)}
			}
		}
	}
	return nil
}

// parseValue parses everything after "key: ".
func parseValue(s string, line int) (Value, error) {
	switch s[0] {
	case '&', '*':
		return Value{}, &SyntaxError{line, "anchors and aliases are not supported"}
	case '|', '>':
		return Value{}, &SyntaxError{line, "multi-line strings are not supported; put prose in the body"}
	case '!':
		return Value{}, &SyntaxError{line, "type tags are not supported"}
	case '{':
		return Value{}, &SyntaxError{line, "inline maps are not supported"}
	case '\'':
		return Value{}, &SyntaxError{line, "single-quoted strings are not supported; use double quotes"}
	case '[':
		return parseFlowList(s, line)
	case '"':
		str, rest, err := parseQuoted(s, line)
		if err != nil {
			return Value{}, err
		}
		if rest = strings.TrimLeft(rest, " "); rest != "" && !isComment(rest) {
			return Value{}, &SyntaxError{line, "unexpected text after a quoted string"}
		}
		return Value{Kind: KindString, Str: str}, nil
	}
	return parseBare(s, line)
}

// parseBare handles an unquoted scalar. As in YAML, " #" starts a trailing
// comment; a value that needs to contain one has to be quoted.
func parseBare(s string, line int) (Value, error) {
	if i := strings.Index(s, " #"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimRight(s, " ")
	if s == "" {
		return Value{}, &SyntaxError{line, "empty value"}
	}
	switch s {
	case "true":
		return Value{Kind: KindBool, Bool: true}, nil
	case "false":
		return Value{Kind: KindBool, Bool: false}, nil
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return Value{Kind: KindInt, Int: n}, nil
	}
	return Value{Kind: KindString, Str: s}, nil
}

// parseQuoted reads a double-quoted string and returns it with the rest of the
// line. Only \\ and \" are escapes; there are no others to learn.
func parseQuoted(s string, line int) (string, string, error) {
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			if i+1 >= len(s) {
				return "", "", &SyntaxError{line, "string ends with a backslash"}
			}
			i++
			switch s[i] {
			case '\\', '"':
				b.WriteByte(s[i])
			default:
				return "", "", &SyntaxError{line, fmt.Sprintf("unsupported escape %q; this dialect has only backslash and quote", s[i-1:i+1])}
			}
		case '"':
			return b.String(), s[i+1:], nil
		default:
			b.WriteByte(s[i])
		}
	}
	return "", "", &SyntaxError{line, "unterminated quoted string"}
}

// parseFlowList reads the [a, b] form.
func parseFlowList(s string, line int) (Value, error) {
	end := strings.LastIndexByte(s, ']')
	if end < 0 {
		return Value{}, &SyntaxError{line, "unterminated list"}
	}
	if rest := strings.TrimLeft(s[end+1:], " "); rest != "" && !isComment(rest) {
		return Value{}, &SyntaxError{line, "unexpected text after a list"}
	}
	inner := strings.TrimSpace(s[1:end])
	v := Value{Kind: KindList}
	if inner == "" {
		return v, nil
	}
	for _, part := range splitFlowItems(inner) {
		item, err := parseListItem(part, line)
		if err != nil {
			return Value{}, err
		}
		v.List = append(v.List, item)
	}
	return v, nil
}

// splitFlowItems splits on commas that are not inside a quoted string.
func splitFlowItems(s string) []string {
	var out []string
	var cur strings.Builder
	quoted := false
	for i := 0; i < len(s); i++ {
		switch {
		case quoted && s[i] == '\\' && i+1 < len(s):
			cur.WriteByte(s[i])
			i++
			cur.WriteByte(s[i])
		case s[i] == '"':
			quoted = !quoted
			cur.WriteByte(s[i])
		case s[i] == ',' && !quoted:
			out = append(out, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(s[i])
		}
	}
	return append(out, cur.String())
}

// parseListItem parses one item of either list form. Items are strings.
func parseListItem(s string, line int) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", &SyntaxError{line, "empty list item"}
	}
	if s[0] == '[' || s[0] == '{' {
		return "", &SyntaxError{line, "nested collections are not supported"}
	}
	if s[0] == '"' {
		str, rest, err := parseQuoted(s, line)
		if err != nil {
			return "", err
		}
		if rest = strings.TrimLeft(rest, " "); rest != "" && !isComment(rest) {
			return "", &SyntaxError{line, "unexpected text after a quoted list item"}
		}
		return str, nil
	}
	v, err := parseBare(s, line)
	if err != nil {
		return "", err
	}
	switch v.Kind {
	case KindInt:
		return strconv.FormatInt(v.Int, 10), nil
	case KindBool:
		return strconv.FormatBool(v.Bool), nil
	default:
		return v.Str, nil
	}
}
