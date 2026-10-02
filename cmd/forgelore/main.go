// Command forgelore is the command-line entry point for Forgelore, a local,
// team-shared memory layer for coding agents.
//
// The CLI is the whole tool. Hooks and MCP are layers over these commands and
// add no capability of their own (decision K6), so anything an agent can do
// through an integration it can also do with a shell.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
)

// version is stamped at link time with -X main.version=<tag>. An unstamped
// build reports "dev".
var version = "dev"

// env is everything a command touches outside itself. Passing it explicitly
// is what lets the tests drive the whole CLI without starting a process or
// changing the working directory of the test binary.
type env struct {
	args   []string
	wd     string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "forgelore:", err)
		os.Exit(1)
	}
	e := env{
		args:   os.Args[1:],
		wd:     wd,
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	if err := run(e); err != nil {
		fmt.Fprintln(os.Stderr, "forgelore:", err)
		os.Exit(1)
	}
}

func run(e env) error {
	if len(e.args) == 0 {
		return usage(e.stdout)
	}

	name, rest := e.args[0], e.args[1:]
	switch name {
	case "version", "--version", "-v":
		_, err := fmt.Fprintf(e.stdout, "forgelore %s %s/%s %s\n",
			version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		return err
	case "help", "--help", "-h":
		return usage(e.stdout)
	case "init":
		return cmdInit(e, rest)
	case "record":
		return cmdRecord(e, rest)
	case "recall":
		return cmdRecall(e, rest)
	case "search":
		return cmdSearch(e, rest)
	case "show":
		return cmdShow(e, rest)
	case "index":
		return cmdIndex(e, rest)
	case "doctor":
		return cmdDoctor(e, rest)
	case "stats":
		return cmdStats(e, rest)
	default:
		return fmt.Errorf("unknown command %q (try: forgelore help)", name)
	}
}

// newFlags builds a flag set that reports its own errors to stderr and leaves
// the exit decision to run.
func newFlags(e env, name string) *flag.FlagSet {
	fs := flag.NewFlagSet("forgelore "+name, flag.ContinueOnError)
	fs.SetOutput(e.stderr)
	return fs
}

// parseInterspersed parses flags that appear after positional arguments as
// well as before them. Go's flag package stops at the first word that is not
// a flag, which would make `forgelore search sqlite --json` search for
// "sqlite --json" — a surprise that costs more than the few lines it takes to
// avoid.
func parseInterspersed(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			return positional, nil
		}
		positional = append(positional, rest[0])
		args = rest[1:]
	}
}

func usage(out io.Writer) error {
	_, err := io.WriteString(out, `forgelore - local, team-shared memory for coding agents

Usage:
  forgelore <command> [flags]

Commands:
  init            Create .forgelore/ in this directory
  record          Write a record
  recall          Look up what is known about a command's error output
  search          Search the records, returning ids and titles only
  show            Print one record in full
  index rebuild   Rebuild the search index from the record files
  doctor          Report anything wrong with the store
  stats           Summarise what the store holds
  version         Print the version and build platform
  help            Print this message

Every command takes --json for machine-readable output, and --dir to work on
a project other than the current directory.

Run a command with -h for its own flags.
`)
	return err
}
