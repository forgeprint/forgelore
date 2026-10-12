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
	// PromptSubmitted: the person named something, and the index may know
	// about it. The fifth canonical event, added by ADR-0020's amendment —
	// the only one that fires on a person speaking rather than on a
	// command or a session boundary.
	PromptSubmitted Kind = "prompt_submitted"
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

	// Transcript is the file the agent writes about this session, and
	// TranscriptFormat says which parser reads it. Both empty when the
	// agent sends no transcript or writes one nothing here understands.
	Transcript       string
	TranscriptFormat string

	// AgentEvent is what the agent called this, kept so a reply can echo
	// it back where the agent expects to see it.
	AgentEvent string

	// Reply is the channel this event answers on: "notice" means the
	// person, anything else means the model.
	Reply string

	// Prompt is the submitted text, set for PromptSubmitted. It is read and
	// never stored: what it names decides a lookup, and the prompt itself
	// goes nowhere (ADR-0013).
	Prompt string

	// Tool, Command and Output are set for the two command events.
	Tool    string
	Command string
	Output  string

	// Inferred marks a CommandFailed the agent did not report as one: the
	// mapping decided it from the shape of the output, because the agent
	// sends no exit status and its success event is all a masked failure
	// produces (ADR-0022).
	//
	// It is a statement about confidence, and the caller spends less on a
	// lookup that rests on it. Output that merely talks about an error —
	// a log being printed, a test fixture being written — looks exactly
	// like output that is one, and in one measured session every single
	// inferred failure was of that kind.
	Inferred bool

	// Untrusted marks an event whose output came from outside the project,
	// such as a web fetch. Anything derived from it is written tainted
	// (ADR-0013).
	Untrusted bool
}
