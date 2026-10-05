package main

import (
	"os/exec"
	"regexp"
	"sort"

	"github.com/forgeprint/forgelore/internal/agent"
)

// semver finds a version in whatever a CLI prints for --version. Each one
// says something different around it.
var semver = regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+`)

// agentStatus is one supported agent as it stands on this machine.
type agentStatus struct {
	Agent     string `json:"agent"`
	Installed string `json:"installed,omitempty"`
	Verified  string `json:"verified,omitempty"`
	Note      string `json:"note"`
}

// agentStatuses compares what is installed with what Forgelore's mappings
// have been checked against.
//
// Every outcome is a note and none is an error. A mapping written for an
// older version usually still works, and a tool that refuses to run because
// an agent moved would be worse than the drift it is reporting.
func agentStatuses() []agentStatus {
	var out []agentStatus
	for name, verified := range agent.Verified() {
		s := agentStatus{Agent: name, Verified: verified}

		bin := agent.Binary(name)
		if path, err := exec.LookPath(bin); err == nil {
			if output, err := exec.Command(path, "--version").Output(); err == nil {
				s.Installed = semver.FindString(string(output))
			}
		}

		switch {
		case s.Installed == "":
			s.Note = "not installed here"
		case verified == "":
			s.Note = "installed, but no mapping has been verified against any version"
		case verified == s.Installed:
			s.Note = "verified against this version"
		default:
			s.Note = "verified against " + verified + ", you have " + s.Installed
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Agent < out[j].Agent })
	return out
}
