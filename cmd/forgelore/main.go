// Command forgelore is the command-line entry point for Forgelore, a local,
// team-shared memory layer for coding agents.
//
// Only the version command exists at this stage. The real command set is
// defined in phase 2 of docs/plan.md; every feature lands here as a CLI
// command first, and hooks and MCP are layers on top (decision K6).
package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

// version is stamped at link time with -X main.version=<tag>. An unstamped
// build reports "dev".
var version = "dev"

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "forgelore:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return usage(out)
	}

	switch args[0] {
	case "version", "--version", "-v":
		_, err := fmt.Fprintf(out, "forgelore %s %s/%s %s\n",
			version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		return err
	case "help", "--help", "-h":
		return usage(out)
	default:
		return fmt.Errorf("unknown command %q (try: forgelore help)", args[0])
	}
}

func usage(out io.Writer) error {
	_, err := io.WriteString(out, `forgelore - local, team-shared memory for coding agents

Usage:
  forgelore <command> [flags]

Commands:
  version   Print the version and build platform
  help      Print this message
`)
	return err
}
