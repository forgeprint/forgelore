package record

import "strings"

// Canonical writing (ADR-0017). A value that looks like a bare token is never
// quoted; anything else is double-quoted; lists use flow form. These rules are
// what make a written record round-trip byte for byte, which in turn is what
// keeps a one-field change to a one-line diff in review.

// bareToken reports whether s can be written unquoted: one or more characters
// from [A-Za-z0-9_.:+-]. That covers enums, ULIDs, fingerprints, timestamps,
// integers and booleans, and nothing with a space or a comment marker in it.
func bareToken(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case c == '_', c == '.', c == ':', c == '+', c == '-':
		default:
			return false
		}
	}
	return true
}

// quote writes a double-quoted string, escaping the only two characters that
// need it.
func quote(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' || s[i] == '"' {
			b.WriteByte('\\')
		}
		b.WriteByte(s[i])
	}
	b.WriteByte('"')
	return b.String()
}

// encodeScalar quotes only when it has to.
func encodeScalar(s string) string {
	if bareToken(s) {
		return s
	}
	return quote(s)
}

// encodeList writes flow form. An empty list is never written at all, so the
// caller omits the key instead of calling this.
func encodeList(items []string) string {
	parts := make([]string, len(items))
	for i, it := range items {
		parts[i] = encodeScalar(it)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
