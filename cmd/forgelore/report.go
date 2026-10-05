package main

import (
	"fmt"
	"time"

	"github.com/forgeprint/forgelore/internal/measure"
)

func cmdReport(e env, args []string) error {
	fs := newFlags(e, "report")
	days := fs.Int("days", 0, "how far back to look; the default comes from report.period_days")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	cfg, _ := loadConfig(e, s, nil)

	period := int(cfg.Int("report.period_days"))
	if *days > 0 {
		period = *days
	}
	until := time.Now().UTC()
	since := until.AddDate(0, 0, -period)

	l, err := measure.Read(ledgerRoot(s), since, until)
	if err != nil {
		return err
	}
	r := measure.Build(l, since, until)

	if *asJSON {
		return writeJSON(e.stdout, r)
	}
	return printReport(e, r, period)
}

func printReport(e env, r *measure.Report, days int) error {
	out := e.stdout
	fmt.Fprintf(out, "Last %d days, to %s\n\n", days, r.Until.Format("2006-01-02"))

	fmt.Fprintln(out, "What Forgelore spent")
	fmt.Fprintf(out, "  errors looked up    %d\n", r.Injections+r.Controls+r.Misses)
	fmt.Fprintf(out, "  hints injected      %d\n", r.Injections)
	fmt.Fprintf(out, "  withheld (control)  %d\n", r.Controls)
	fmt.Fprintf(out, "  nothing known       %d\n", r.Misses)
	fmt.Fprintf(out, "  bytes injected      %d\n", r.InjectedBytes)
	fmt.Fprintf(out, "  tokens injected     %d  (estimated at four bytes per token)\n", r.InjectedTokens)

	printComparison(e, "Repeated errors per session", r.Loops,
		"Measured from Forgelore's own ledger, so it needs nothing from the agent.",
		func(v float64) string { return fmt.Sprintf("%.2f", v) })

	printComparison(e, "Cost per session", r.Cost,
		"Read from the agent's own reporting. The agent computes it client-side at list price, so it is an estimate of your bill, not your bill.",
		func(v float64) string { return fmt.Sprintf("$%.4f", v) })

	if r.Cost == nil {
		fmt.Fprintln(out, "\nCost per session")
		fmt.Fprintln(out, "  No usage was reported for this period, so there is nothing to")
		fmt.Fprintln(out, "  compare. Forgelore will not claim a saving it has not measured.")
		fmt.Fprintln(out, "  To wire it up: forgelore usage --help-wiring")
	}

	if r.Unreadable > 0 {
		fmt.Fprintf(out, "\n%d ledger line(s) could not be read and were left out.\n", r.Unreadable)
	}
	return nil
}

func printComparison(e env, title string, c *measure.Comparison, note string, format func(float64) string) {
	if c == nil {
		return
	}
	out := e.stdout
	fmt.Fprintf(out, "\n%s\n", title)
	fmt.Fprintf(out, "  %-10s %8s  %8s  %8s\n", "", "sessions", "median", "mean")
	fmt.Fprintf(out, "  %-10s %8d  %8s  %8s\n", "treatment", c.Treatment.Sessions, format(c.Treatment.Median), format(c.Treatment.Mean))
	fmt.Fprintf(out, "  %-10s %8d  %8s  %8s\n", "control", c.Control.Sessions, format(c.Control.Median), format(c.Control.Mean))

	switch {
	case c.Reason != "":
		fmt.Fprintf(out, "\n  Not meaningful: %s.\n", c.Reason)
	case c.Significant:
		direction := "lower"
		if c.Difference < 0 {
			direction = "higher"
		}
		fmt.Fprintf(out, "\n  Treatment is %s by %s (95%% interval %s to %s).\n",
			direction, format(abs(c.Difference)), format(c.LowCI), format(c.HighCI))
	}
	fmt.Fprintf(out, "  %s\n", note)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
