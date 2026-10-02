package fingerprint

import "strings"

// runners are wrappers that run some other tool. The diagnostic belongs to the
// tool, not to the wrapper, so the wrapper is stepped over: "npx --yes -p
// typescript tsc" is tsc.
var runners = map[string]bool{
	"npx":  true,
	"bunx": true,
	"pnpm": true,
	"yarn": true,
	"npm":  true,
}

// runnerFlagsWithValue are the wrapper flags that consume the token after
// them, which would otherwise be mistaken for the tool. The list is short on
// purpose and covers what the corpus uses; an unlisted flag that takes a value
// makes NormalizeCommand pick the wrong tool, which splits a fingerprint
// rather than merging two, so the failure is a miss and not a wrong hint.
var runnerFlagsWithValue = map[string]bool{
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
	words := strings.Fields(command)

	// Leading VAR=value assignments belong to the environment, not the
	// command.
	for len(words) > 0 && isAssignment(words[0]) {
		words = words[1:]
	}

	// Step over a wrapper and its own flags.
	if len(words) > 0 && runners[toolName(words[0])] {
		words = words[1:]
		for len(words) > 0 && strings.HasPrefix(words[0], "-") {
			takesValue := runnerFlagsWithValue[words[0]]
			words = words[1:]
			if takesValue && len(words) > 0 {
				words = words[1:]
			}
		}
	}

	if len(words) == 0 {
		return ""
	}

	return toolName(words[0])
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
