package measure

import (
	"crypto/sha256"
	"encoding/binary"
)

// Assign puts a session in the treatment or the control arm.
//
// The assignment is derived from the session identifier rather than drawn at
// random and stored, because Forgelore has no process to store it in: every
// command is a fresh invocation that exits (ADR-0008). Hashing means the
// hundredth lookup in a session lands in the same arm as the first, with no
// state to keep, no file to corrupt and nothing to go stale.
//
// controlPercent outside 0..100 is clamped. Zero — the default — puts every
// session in treatment, which is to say the trial is off.
func Assign(salt, sessionID string, controlPercent int) string {
	switch {
	case controlPercent <= 0:
		return GroupTreatment
	case controlPercent >= 100:
		return GroupControl
	}

	// The separator keeps salt "ab"+session "c" from colliding with salt
	// "a"+session "bc"; a null byte cannot appear in either.
	h := sha256.Sum256([]byte(salt + "\x00" + sessionID))
	if int(binary.BigEndian.Uint64(h[:8])%100) < controlPercent {
		return GroupControl
	}
	return GroupTreatment
}

// EstimateTokens approximates how many tokens a string costs.
//
// It is four bytes per token, rounded up. That is the crude rule of thumb for
// English prose and it is wrong for code, which packs more tokens into the
// same bytes. The accurate answer needs the model's tokeniser, which is a
// dependency this project does not take (K3), and a model call, which it does
// not make (K8).
//
// So the estimate stays crude and stays labelled: `report` prints token
// figures as estimates, and compares the two arms on cost in dollars, which
// the agent reports and nobody has to estimate.
func EstimateTokens(s string) int { return EstimateTokensFromBytes(len(s)) }

// EstimateTokensFromBytes is the same estimate for callers that have counted
// the bytes but no longer hold the text.
func EstimateTokensFromBytes(n int) int {
	if n <= 0 {
		return 0
	}
	return (n + 3) / 4
}
