package main

import (
	"github.com/forgeprint/forgelore/internal/mcp"
)

// cmdMCP serves the Model Context Protocol on stdin and stdout, so that any
// MCP client can search this project's memory without an agent-specific
// integration.
//
// Nothing may be printed to stdout but protocol messages, so every diagnostic
// this command has goes to stderr — including the error it exits with.
func cmdMCP(e env, args []string) error {
	fs := newFlags(e, "mcp")
	dir := fs.String("dir", "", "serve the project containing this directory")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	return mcp.New(s, version).Serve(e.stdin, e.stdout)
}
