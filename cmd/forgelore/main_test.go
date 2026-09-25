package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"version"}, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := out.String(); !strings.HasPrefix(got, "forgelore ") {
		t.Errorf("version output = %q, want it to start with %q", got, "forgelore ")
	}
}

func TestRunNoArgsPrintsUsage(t *testing.T) {
	var out bytes.Buffer
	if err := run(nil, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("no-args output = %q, want it to contain %q", out.String(), "Usage:")
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"nope"}, &out)
	if err == nil {
		t.Fatal("run with unknown command: got nil error, want an error")
	}
	if out.Len() != 0 {
		t.Errorf("unknown command wrote %q to stdout, want nothing", out.String())
	}
}
