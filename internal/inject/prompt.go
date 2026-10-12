// Package inject decides what a prompt is about.
//
// Retrieval at prompt time is the one path that spends tokens on every prompt
// rather than on an event, so what it spends them on has to be narrow. It
// matches on the things a person names when they mean a particular part of a
// codebase — a path and an identifier — and never on the prose around them.
// Sending whole prompts to a search index would return something for almost
// anything, which is the same as returning nothing useful at a cost per turn.
package inject

import (
	"regexp"
	"strings"
)

// maxTerms bounds the work one prompt can cause. A prompt naming thirty files
// is a prompt where the agent is not looking for a reminder.
const maxTerms = 8

var (
	// A path: something with a separator, or something with a file
	// extension. Both forms appear in the way people write about code —
	// "internal/store/index.go" and "index.go" mean the same file.
	pathish = regexp.MustCompile(`[A-Za-z0-9_.-]*(?:/[A-Za-z0-9_.-]+)+|[A-Za-z0-9_-]+\.[A-Za-z][A-Za-z0-9]{0,8}`)

	// An identifier: snake_case, kebab-case, or camel case either way up.
	// A single lowercase word is not one, because it is also just a word —
	// an internal capital or separator is what makes it a name.
	identish = regexp.MustCompile(`\b[a-z0-9]+(?:_[a-z0-9]+)+\b|\b[a-z0-9]+(?:-[a-z0-9]+)+\b|\b[A-Za-z][a-z0-9]+(?:[A-Z][a-z0-9]*)+\b`)

	// Pasted content arrives expanded and marked, and what it contains is
	// not what the person said. Documented for UserPromptSubmit, verified
	// against the documentation only.
	pasted = regexp.MustCompile(`(?s)<pasted_content[^>]*>.*?</pasted_content[^>]*>`)
)

// Terms returns what a prompt names, lowercased and in the order it named
// them, or nothing when it names nothing in particular.
func Terms(prompt string) []string {
	prompt = pasted.ReplaceAllString(prompt, " ")

	seen := map[string]bool{}
	var out []string
	add := func(term string) {
		term = strings.ToLower(strings.Trim(term, "./-_"))
		if len(term) < 3 || seen[term] || len(out) >= maxTerms {
			return
		}
		seen[term] = true
		out = append(out, term)
	}
	for _, m := range pathish.FindAllString(prompt, -1) {
		add(m)
	}
	for _, m := range identish.FindAllString(prompt, -1) {
		add(m)
	}
	return out
}
