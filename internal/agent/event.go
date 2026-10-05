// Package agent turns one coding agent's hook payload into the small set of
// events Forgelore acts on, using a mapping file rather than code (K5).
//
// The canonical events are deliberately not a copy of any agent's lifecycle.
// Claude Code alone fires more than thirty events; mirroring them would mean
// every new agent adds vocabulary to a model that only ever needed four
// questions answered: has a session begun, did a command fail, did one
// succeed, has the session ended.
package agent

import "time"

// Kind is a canonical event.
type Kind string

const (
	// SessionStarted: a budgeted index may be offered.
	SessionStarted Kind = "session_started"
	// CommandFailed: the moment Forgelore exists for.
	CommandFailed Kind = "command_failed"
	// CommandSucceeded: a command that failed earlier in this session now
	// works, so there may be a fix worth proposing.
	CommandSucceeded Kind = "command_succeeded"
	// SessionEnded: anything proposed during the session is summarised.
	SessionEnded Kind = "session_ended"
)

// Event is one thing that happened, in Forgelore's vocabulary.
type Event struct {
	Kind    Kind
	Time    time.Time
	Session string
	Cwd     string

	// Scratchpad is a directory the agent keeps for this session. Forgelore
	// remembers which errors have already failed in this run there, because
	// every hook is a separate process with no memory of the last one (K8).
	// Empty when the agent offers none, and then a directory under the
	// project's own cache is used instead.
	Scratchpad string

	// AgentEvent is what the agent called this, kept so a reply can echo
	// it back where the agent expects to see it.
	AgentEvent string

	// Tool, Command and Output are set for the two command events.
	Tool    string
	Command string
	Output  string

	// Untrusted marks an event whose output came from outside the project,
	// such as a web fetch. Anything derived from it is written tainted
	// (ADR-0013).
	Untrusted bool
}
