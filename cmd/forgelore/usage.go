package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/forgeprint/forgelore/internal/measure"
)

// statusLine is the part of Claude Code's status line payload that bears on
// measurement. Verified against the official reference on 2026-10-02:
// https://code.claude.com/docs/en/statusline
//
// Only cost.total_cost_usd is cumulative over the session — the documentation
// calls it "the estimated cost of all API calls in the current session", and
// it resets when /clear starts a new one. The context_window token counts are
// explicitly "tokens currently in the context window, from the most recent
// API response", so they are recorded for context and never added up.
type statusLine struct {
	SessionID string `json:"session_id"`
	Version   string `json:"version"`
	Cost      struct {
		TotalCostUSD    float64 `json:"total_cost_usd"`
		TotalDurationMS int64   `json:"total_duration_ms"`
	} `json:"cost"`
	ContextWindow *struct {
		TotalInputTokens  int64 `json:"total_input_tokens"`
		TotalOutputTokens int64 `json:"total_output_tokens"`
	} `json:"context_window"`
}

const usageHelp = `forgelore usage records what a session has spent so far.

Claude Code hands this data to the status line command and to nothing else:
hooks do not receive it. So wire forgelore into your status line and let it
read the same JSON, in .claude/settings.json:

    "statusLine": {
      "type": "command",
      "command": "tee >(forgelore usage >/dev/null) | your-status-script"
    }

Nothing is printed on success, so it composes in a pipe.

For an agent that reports usage some other way, pass the figures directly
with --session and --cost-usd and leave stdin empty.`

func cmdUsage(e env, args []string) error {
	fs := newFlags(e, "usage")
	agent := fs.String("agent", "claude-code", "which agent the reading came from")
	format := fs.String("format", "claude-code-statusline", "the shape of the JSON on stdin, or 'none' to use the flags alone")
	session := fs.String("session", "", "override the session id")
	costUSD := fs.Float64("cost-usd", -1, "the session's cumulative cost, when not reading JSON")
	dir := fs.String("dir", "", "work on the project containing this directory")
	explain := fs.Bool("help-wiring", false, "print how to wire this into an agent")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *explain {
		_, err := fmt.Fprintln(e.stdout, usageHelp)
		return err
	}

	u := measure.Usage{
		Time:    time.Now().UTC().Truncate(time.Second),
		Agent:   *agent,
		Session: *session,
		CostUSD: *costUSD,
	}

	switch *format {
	case "none":
		if *costUSD < 0 {
			return fmt.Errorf("--format none needs --cost-usd")
		}
	case "claude-code-statusline":
		raw, err := readFileOrStdin(e, "-")
		if err != nil {
			return err
		}
		if strings.TrimSpace(raw) == "" {
			return fmt.Errorf("no JSON on stdin (try: forgelore usage --help-wiring)")
		}
		var sl statusLine
		if err := json.Unmarshal([]byte(raw), &sl); err != nil {
			return fmt.Errorf("that is not the status line JSON: %w", err)
		}
		if u.Session == "" {
			u.Session = sl.SessionID
		}
		u.AgentVersion = sl.Version
		u.DurationMS = sl.Cost.TotalDurationMS
		if *costUSD < 0 {
			u.CostUSD = sl.Cost.TotalCostUSD
		}
		if sl.ContextWindow != nil {
			u.ContextIn = sl.ContextWindow.TotalInputTokens
			u.ContextOut = sl.ContextWindow.TotalOutputTokens
		}
	default:
		return fmt.Errorf("unknown --format %q (claude-code-statusline, none)", *format)
	}

	if u.Session == "" {
		return fmt.Errorf("no session id, so this reading could not be attributed to a run")
	}
	if u.CostUSD < 0 {
		return fmt.Errorf("no cost in that payload")
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	cfg, _ := loadConfig(e, s, nil)
	if !cfg.Bool("measure.ledger") {
		return nil
	}
	u.Group = measure.Assign(
		cfg.String("measure.ab.salt"),
		u.Session,
		int(cfg.Int("measure.ab.control_percent")),
	)

	if err := measure.AppendUsage(ledgerRoot(s), u); err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(e.stdout, u)
	}
	return nil
}
