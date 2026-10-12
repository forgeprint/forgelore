package measure

import (
	"fmt"
	"testing"
	"time"
)

// TestDiscoveryComesFromSessionLinesOnly keeps the proxy honest in both
// directions: a session nobody harvested is not a session that looked nothing
// up, and an injection line is not a measurement of a session.
func TestDiscoveryComesFromSessionLinesOnly(t *testing.T) {
	at := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	var entries []Entry

	// Four sessions per arm, so both clear minArmSessions.
	for i := 0; i < 4; i++ {
		entries = append(entries,
			Entry{Time: at, Session: fmt.Sprintf("t%d", i), Group: GroupTreatment,
				Event: EventSession, Calls: 100, Discovery: 10},
			Entry{Time: at, Session: fmt.Sprintf("c%d", i), Group: GroupControl,
				Event: EventSession, Calls: 100, Discovery: 40},
		)
	}
	// Noise that must not become a data point: an injection in a session
	// with no session line, and a harvested session with no calls.
	entries = append(entries,
		Entry{Time: at, Session: "x", Group: GroupTreatment, Event: EventInject, Bytes: 40},
		Entry{Time: at, Session: "y", Group: GroupControl, Event: EventSession, Calls: 0, Discovery: 0},
	)

	r := Build(&Ledger{Entries: entries}, at.Add(-time.Hour), at.Add(time.Hour))
	if r.Discovery == nil {
		t.Fatal("no discovery comparison was built")
	}
	if n := r.Discovery.Treatment.Sessions; n != 4 {
		t.Errorf("treatment sessions = %d, want 4", n)
	}
	if n := r.Discovery.Control.Sessions; n != 4 {
		t.Errorf("control sessions = %d, want 4 (the one with no calls is not a measurement)", n)
	}
	// Control looked four times as much, so the difference is positive:
	// control minus treatment, the direction Forgelore claims.
	if r.Discovery.Difference <= 0 {
		t.Errorf("difference = %v, want control higher", r.Discovery.Difference)
	}
}
