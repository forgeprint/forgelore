package fingerprint

import "strings"

// executors do nothing but run a binary their own command line names. The
// diagnostic belongs to that binary, so the executor is stepped over:
// "npx --yes -p typescript tsc" is tsc.
var executors = map[string]bool{
	"npx":  true,
	"bunx": true,
}

// packageManagers run their own subcommands. "npm run build" names a script,
// not a tool, and what that script invokes is not on the command line at
// all — so the package manager is the tool, exactly as "go" is the tool for
// "go build" and "go test". Treating the subcommand as the tool used to
// fingerprint a build failure under "run" and a test failure under "test",
// splitting one error across the ways it can be provoked, which is what this
// whole function exists to prevent.
var packageManagers = map[string]bool{
	"npm":  true,
	"yarn": true,
	"pnpm": true,
	"bun":  true,
}

// handsOver are the package manager subcommands that do name a binary, and
// so behave like an executor.
var handsOver = map[string]bool{
	"exec": true,
	"dlx":  true,
	"x":    true,
}

// wrapperFlagsWithValue are the wrapper flags that consume the token after
// them, which would otherwise be mistaken for the tool. The list is short on
// purpose and covers what the corpus uses; an unlisted flag that takes a value
// makes NormalizeCommand pick the wrong tool, which splits a fingerprint
// rather than merging two, so the failure is a miss and not a wrong hint.
var wrapperFlagsWithValue = map[string]bool{
	"-p":        true,
	"--package": true,
	"-c":        true,
	"--call":    true,
}

// NormalizeCommand reduces a command line to the tool that ran.
//
// Everything else goes: leading environment assignments, the verb, flags,
// paths, file arguments. Keeping any of it would split one error across every
// way it can be provoked, and the verb splits the most: the same compile
// error reaches the user through go build, go test, go run and go vet, and a
// fix recorded under one of them has to be found under the others. What the
// verb would have told the fingerprint apart, the message already does.
func NormalizeCommand(command string) string {
	words := strings.Fields(lastInChain(command))

	// Leading VAR=value assignments belong to the environment, not the
	// command.
	for len(words) > 0 && isAssignment(words[0]) {
		words = words[1:]
	}

	if len(words) == 0 {
		return ""
	}

	switch head := toolName(words[0]); {
	case executors[head]:
		words = skipWrapper(words[1:])

	case packageManagers[head]:
		// Only a subcommand that hands over to a named binary is stepped
		// through. Everything else leaves the package manager as the tool.
		rest := skipWrapper(words[1:])
		if len(rest) > 0 && handsOver[rest[0]] {
			words = skipWrapper(rest[1:])
		} else {
			return head
		}
	}

	if len(words) == 0 {
		return ""
	}

	return toolName(words[0])
}

// lastInChain returns the final command of a shell chain.
//
// An agent rarely runs a build on its own. It writes `cd web && npm run
// build` or `ls -a && go build ./...`, and taking the first word makes the
// compiler's diagnostic belong to `cd` or to `ls` — a fingerprint that
// matches nothing, silently, because every lookup still succeeds and simply
// finds no record. This was watched happening in a real session.
//
// The last link is a guess, and it is wrong for `go build ./... && echo ok`,
// where the error comes from the first. It is the better guess: a chain is
// written to reach its last command, and the earlier links are the setup.
//
// Pipes are deliberately not split on. `go build ./... 2>&1 | head -40`
// produces its diagnostic in the first element and feeds it to the rest, so
// there the first word is already the right answer.
func lastInChain(command string) string {
	last := command
	for _, sep := range []string{"&&", "||", ";"} {
		for {
			i := strings.LastIndex(last, sep)
			if i < 0 {
				break
			}
			// `find . -exec rm {} \;` ends in an escaped semicolon that
			// belongs to find's arguments, and splitting there would
			// leave nothing at all.
			if i > 0 && last[i-1] == '\\' {
				break
			}
			tail := last[i+len(sep):]
			if strings.TrimSpace(tail) == "" {
				break
			}
			last = tail
			break
		}
	}
	return last
}

// skipWrapper drops the leading flags of a wrapper, and the values of the
// flags known to take one.
func skipWrapper(words []string) []string {
	for len(words) > 0 && strings.HasPrefix(words[0], "-") {
		takesValue := wrapperFlagsWithValue[words[0]]
		words = words[1:]
		if takesValue && len(words) > 0 {
			words = words[1:]
		}
	}
	return words
}

// isAssignment reports whether a word is a VAR=value environment assignment
// rather than the command. A word that starts with '=' is not one.
func isAssignment(word string) bool {
	i := strings.IndexByte(word, '=')
	return i > 0 && !strings.ContainsAny(word[:i], `/\`)
}

// toolName strips the directory and the Windows executable suffix, so that
// /usr/local/go/bin/go, go and go.exe are one tool.
func toolName(word string) string {
	if i := strings.LastIndexAny(word, `/\`); i >= 0 {
		word = word[i+1:]
	}
	return strings.TrimSuffix(strings.TrimSuffix(word, ".exe"), ".EXE")
}
