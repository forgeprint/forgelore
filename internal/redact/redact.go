// Package redact masks credentials and drops private sections before a record
// is written.
//
// This runs at write time, not at display time (ADR-0013). A secret that
// reaches disk is already a problem: it is in a file, possibly in a git
// object, and removing it later does not remove it from history.
//
// Pattern matching both misses things and occasionally masks something
// harmless. It is a layer, not a guarantee; team records are scanned again
// before they can be committed.
package redact

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Mark is what replaces a masked value. It names the pattern that matched, so
// a reader can tell a redaction from the original text.
const markPrefix = "[redacted:"

// Private marks a section that is never stored at all.
const (
	openTag  = "<private>"
	closeTag = "</private>"
)

// Finding counts what a redaction pass removed.
type Finding struct {
	Kind  string
	Count int
}

type pattern struct {
	kind string
	re   *regexp.Regexp
	// group is the submatch to mask. 0 masks the whole match.
	group int
}

// patterns are ordered: the most specific first, so that a GitHub token is
// reported as one rather than as a generic assignment.
var patterns = []pattern{
	{kind: "private-key-block", re: regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`)},
	{kind: "github-token", re: regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{20,}`)},
	{kind: "slack-token", re: regexp.MustCompile(`xox[abprs]-[A-Za-z0-9-]{10,}`)},
	{kind: "aws-access-key-id", re: regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{kind: "google-api-key", re: regexp.MustCompile(`AIza[0-9A-Za-z_-]{35}`)},
	{kind: "stripe-key", re: regexp.MustCompile(`[sr]k_live_[0-9A-Za-z]{16,}`)},
	{kind: "anthropic-key", re: regexp.MustCompile(`sk-ant-[A-Za-z0-9_-]{20,}`)},
	{kind: "openai-key", re: regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`)},
	{kind: "jwt", re: regexp.MustCompile(`eyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}`)},
	{kind: "bearer-token", re: regexp.MustCompile(`(?i)(bearer\s+)([A-Za-z0-9._~+/=-]{16,})`), group: 2},
	{kind: "url-credentials", re: regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+.-]*://[^\s:/@]+:)([^\s@]+)(@)`), group: 2},
	// Last: an assignment whose left-hand side names a secret. Broad on
	// purpose, because the specific patterns cannot cover every provider.
	// The leading run of name characters matters: DB_PASSWORD has no word
	// boundary before PASSWORD, and that spelling is the common one.
	{kind: "assigned-secret", re: regexp.MustCompile(`(?i)[A-Za-z0-9_.-]*(password|passwd|secret|token|api[_-]?key|access[_-]?key|auth[_-]?token|client[_-]?secret)\b\s*[:=]\s*("[^"\n]{4,}"|'[^'\n]{4,}'|[^\s,;)"']{4,})`), group: 2},
}

// Text removes private sections and masks credentials. It returns the cleaned
// text and what it found.
func Text(s string) (string, []Finding) {
	counts := map[string]int{}

	s = stripPrivate(s, counts)
	for _, p := range patterns {
		s = maskPattern(s, p, counts)
	}
	return s, findings(counts)
}

// stripPrivate removes every <private>...</private> section. An unclosed tag
// removes everything after it: when it is unclear where the private section
// ends, the safe reading is "all of it".
func stripPrivate(s string, counts map[string]int) string {
	var b strings.Builder
	for {
		start := strings.Index(s, openTag)
		if start < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:start])
		counts["private-section"]++

		rest := s[start+len(openTag):]
		end := strings.Index(rest, closeTag)
		if end < 0 {
			// Unclosed: drop the remainder.
			return b.String()
		}
		s = rest[end+len(closeTag):]
	}
}

func maskPattern(s string, p pattern, counts map[string]int) string {
	return p.re.ReplaceAllStringFunc(s, func(match string) string {
		if p.group == 0 {
			counts[p.kind]++
			return mark(p.kind)
		}
		sub := p.re.FindStringSubmatchIndex(match)
		if sub == nil || len(sub) <= 2*p.group+1 || sub[2*p.group] < 0 {
			return match
		}
		// Already masked once by a more specific pattern: leave it alone.
		value := match[sub[2*p.group]:sub[2*p.group+1]]
		if strings.Contains(value, markPrefix) {
			return match
		}
		counts[p.kind]++
		return match[:sub[2*p.group]] + mark(p.kind) + match[sub[2*p.group+1]:]
	})
}

func mark(kind string) string { return markPrefix + kind + "]" }

func findings(counts map[string]int) []Finding {
	if len(counts) == 0 {
		return nil
	}
	out := make([]Finding, 0, len(counts))
	for kind, n := range counts {
		out = append(out, Finding{Kind: kind, Count: n})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind < out[j].Kind })
	return out
}

// Merge combines finding lists, which a caller redacting several strings needs
// in order to report once.
func Merge(lists ...[]Finding) []Finding {
	counts := map[string]int{}
	for _, list := range lists {
		for _, f := range list {
			counts[f.Kind] += f.Count
		}
	}
	return findings(counts)
}

// Summary renders findings for a person, for example
// "github-token x1, private-section x2".
func Summary(fs []Finding) string {
	if len(fs) == 0 {
		return ""
	}
	parts := make([]string, len(fs))
	for i, f := range fs {
		parts[i] = f.Kind + " x" + strconv.Itoa(f.Count)
	}
	return strings.Join(parts, ", ")
}
