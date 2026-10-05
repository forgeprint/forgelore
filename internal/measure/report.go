package measure

import (
	"math"
	"math/rand/v2"
	"sort"
	"time"
)

// bootstrapRounds is how many resamples the confidence interval is built
// from. Two thousand is enough for a 95% interval to settle and cheap enough
// that `report` stays instant.
const bootstrapRounds = 2000

// minArmSessions is the smallest arm a comparison is attempted on. Below it
// the interval is so wide that quoting it would dress up noise as a finding.
const minArmSessions = 3

// Report is everything `report` prints for a period.
type Report struct {
	Since time.Time
	Until time.Time

	// The spending side. Always available: it is what Forgelore itself did.
	Injections     int
	Controls       int
	Misses         int
	InjectedBytes  int
	InjectedTokens int

	// The outcome side. Loops are measured from Forgelore's own ledger and
	// are therefore always available. Cost needs the agent to have reported
	// its usage, and is nil when it did not.
	Loops *Comparison
	Cost  *Comparison

	// Hook latency, in milliseconds. Nil when no hook has run: the CLI path
	// records nothing here, because nobody is waiting on it.
	Hook *Latency

	Unreadable int
}

// Latency is what a hook cost the user in wall time.
type Latency struct {
	Samples int     `json:"samples"`
	P50     float64 `json:"p50_ms"`
	P95     float64 `json:"p95_ms"`
	Max     float64 `json:"max_ms"`
}

// Arm is one side of the trial.
type Arm struct {
	Sessions int
	Median   float64
	Mean     float64
	Total    float64
}

// Comparison is control measured against treatment.
//
// Difference is control minus treatment, so a positive number means the
// treatment arm came out lower, which for both cost and repeated loops is
// the direction Forgelore claims.
type Comparison struct {
	Unit       string
	Treatment  Arm
	Control    Arm
	Difference float64
	LowCI      float64
	HighCI     float64

	// Significant is false whenever the interval covers zero, and also
	// whenever an arm was too small to compare. Reason says which.
	Significant bool
	Reason      string
}

// Build turns a period of ledger lines into a report.
func Build(l *Ledger, since, until time.Time) *Report {
	r := &Report{Since: since, Until: until, Unreadable: l.Unreadable}

	for _, e := range l.Entries {
		switch e.Event {
		case EventInject:
			r.Injections++
			r.InjectedBytes += e.Bytes
			r.InjectedTokens += e.EstTokens
		case EventControl:
			r.Controls++
		case EventMiss:
			r.Misses++
		}
	}

	r.Hook = hookLatency(l.Entries)
	r.Loops = compare("repeated errors per session", loopsPerSession(l.Entries))
	r.Cost = compare("USD per session", costPerSession(l.Usage))
	return r
}

// loopsPerSession counts, for each session, how many errors it met more than
// once. A repeated error is the thing memory is supposed to prevent, and it
// is visible in Forgelore's own ledger without the agent reporting anything.
func loopsPerSession(entries []Entry) map[string][]float64 {
	type key struct{ session, fingerprint string }
	seen := make(map[key]int)
	group := make(map[string]string)

	for _, e := range entries {
		// A line with no session is still a real injection and still
		// counts on the spending side, but it cannot be attributed to a
		// run, and lumping every such line together would invent loops
		// between unrelated invocations.
		if e.Session == "" {
			continue
		}
		seen[key{e.Session, e.Fingerprint}]++
		if e.Group != "" {
			group[e.Session] = e.Group
		}
	}

	loops := make(map[string]int, len(group))
	for session := range group {
		loops[session] = 0
	}
	for k, n := range seen {
		if n > 1 {
			loops[k.session] += n - 1
		}
	}

	byArm := make(map[string][]float64)
	for session, n := range loops {
		arm := group[session]
		byArm[arm] = append(byArm[arm], float64(n))
	}
	return byArm
}

// costPerSession reduces the usage log to one figure per session.
//
// CostUSD is a cumulative counter that the agent reports over and over as the
// session runs, so the session's cost is the largest reading, never the sum
// of them. Summing would multiply a session's cost by how often its status
// line happened to refresh.
func costPerSession(usage []Usage) map[string][]float64 {
	highest := make(map[string]float64)
	group := make(map[string]string)
	for _, u := range usage {
		if u.Session == "" {
			continue
		}
		if u.CostUSD > highest[u.Session] {
			highest[u.Session] = u.CostUSD
		}
		if u.Group != "" {
			group[u.Session] = u.Group
		}
	}

	byArm := make(map[string][]float64)
	for session, cost := range highest {
		byArm[group[session]] = append(byArm[group[session]], cost)
	}
	return byArm
}

func compare(unit string, byArm map[string][]float64) *Comparison {
	treatment, control := byArm[GroupTreatment], byArm[GroupControl]
	if len(treatment) == 0 && len(control) == 0 {
		return nil
	}

	// Both arms are built by ranging over a map, and Go randomises that
	// order. Sorting here is what makes the resampling below land on the
	// same values every run; without it the interval moves between two
	// reports on identical data.
	sort.Float64s(treatment)
	sort.Float64s(control)

	c := &Comparison{
		Unit:      unit,
		Treatment: summarise(treatment),
		Control:   summarise(control),
	}
	c.Difference = c.Control.Median - c.Treatment.Median

	if len(treatment) < minArmSessions || len(control) < minArmSessions {
		c.Reason = "not enough sessions in both arms to compare"
		return c
	}

	c.LowCI, c.HighCI = bootstrapDifference(treatment, control)
	if c.LowCI > 0 || c.HighCI < 0 {
		c.Significant = true
	} else {
		c.Reason = "the interval covers zero, so the difference is not meaningful"
	}
	return c
}

func summarise(xs []float64) Arm {
	a := Arm{Sessions: len(xs)}
	if len(xs) == 0 {
		return a
	}
	for _, x := range xs {
		a.Total += x
	}
	a.Mean = a.Total / float64(len(xs))
	a.Median = median(xs)
	return a
}

// bootstrapDifference returns a 95% interval for the difference of medians.
//
// A bootstrap rather than a t-test: session costs are skewed and bounded
// below by zero, nothing about them is normal, and the arms are usually
// small. Resampling assumes only that the sessions are a sample of
// themselves.
//
// The generator is seeded with a constant, so running `report` twice on the
// same ledger gives the same interval. A measurement that moves when you look
// at it twice is not one anybody will believe.
func bootstrapDifference(treatment, control []float64) (low, high float64) {
	rng := rand.New(rand.NewPCG(0x666f7267, 0x656c6f72))

	diffs := make([]float64, bootstrapRounds)
	for i := range diffs {
		diffs[i] = median(resample(rng, control)) - median(resample(rng, treatment))
	}
	sort.Float64s(diffs)
	return percentile(diffs, 2.5), percentile(diffs, 97.5)
}

func resample(rng *rand.Rand, xs []float64) []float64 {
	out := make([]float64, len(xs))
	for i := range out {
		out[i] = xs[rng.IntN(len(xs))]
	}
	return out
}

// median sorts a copy, so the caller's slice keeps its order.
func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sorted := append([]float64(nil), xs...)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}

// percentile reads a value out of an already sorted slice, interpolating
// between the two neighbours.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	pos := p / 100 * float64(len(sorted)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return sorted[lo]
	}
	return sorted[lo] + (pos-float64(lo))*(sorted[hi]-sorted[lo])
}

// hookLatency summarises how long the agent waited on Forgelore.
func hookLatency(entries []Entry) *Latency {
	var ms []float64
	for _, e := range entries {
		if e.HookMS > 0 {
			ms = append(ms, float64(e.HookMS))
		}
	}
	if len(ms) == 0 {
		return nil
	}
	sort.Float64s(ms)
	return &Latency{
		Samples: len(ms),
		P50:     percentile(ms, 50),
		P95:     percentile(ms, 95),
		Max:     ms[len(ms)-1],
	}
}
