package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/forgeprint/forgelore/internal/agent"
	"github.com/forgeprint/forgelore/internal/candidate"
	"github.com/forgeprint/forgelore/internal/measure"
	"github.com/forgeprint/forgelore/internal/record"
	"github.com/forgeprint/forgelore/internal/redact"
	"github.com/forgeprint/forgelore/internal/store"
)

// instruction is what the agent's own CLI is asked for.
//
// One line, and only the line. The body it is given is already in the review
// queue; the single thing missing is the sentence that would be injected, and
// `review --accept` asks a person for it. A model is good at that sentence and
// a parser is not, which is the whole case for spending anything here
// (ADR-0027).
const instruction = `Below is a note about one coding session in a project.
Write ONE line, at most 100 characters, that states what was learned.

Rules:
- State the answer, not the question: "The index is rebuilt on every open",
  not "Why is the index slow".
- Use only what the note says. Invent nothing, and do not guess a cause.
- No quotes, no markdown, no trailing period, no preamble. The line only.
- If the note says nothing worth keeping, reply with exactly: SKIP`

// maxTitle is where a title stops being a title.
const maxTitle = 120

// cmdDistill asks the agent's own CLI for the one line a work candidate is
// missing, and writes it to the candidate.
//
// It is a command before it is a hook (K6), and the hook only ever starts this
// command in a child it does not wait for, so nothing a model does can slow an
// agent down (ADR-0027).
//
// There is no counter enforcing one call per session, because none is needed:
// distillation fills an empty title and skips a candidate that has one, so a
// second call for the same session has nothing to do.
func cmdDistill(e env, args []string) error {
	fs := newFlags(e, "distill")
	key := fs.String("key", "", "the candidate to title, as `review` lists it")
	dir := fs.String("dir", "", "work on the project containing this directory")
	adapter := fs.String("adapter", "claude-code", "which agent's CLI to run")
	mappingFile := fs.String("mapping", "", "read the agent mapping from this file")
	dryRun := fs.Bool("dry-run", false, "print what would be sent and run nothing")
	asJSON := fs.Bool("json", false, "machine-readable output")
	if _, err := parseInterspersed(fs, args); err != nil {
		return err
	}
	if *key == "" {
		return fmt.Errorf("distill needs --key (forgelore review lists them)")
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	cfg, _ := loadConfig(e, s, nil)
	if !cfg.Bool("distill.enabled") && !*dryRun {
		return fmt.Errorf("distillation is off; turn it on with distill.enabled in .forgelore/config.yaml")
	}

	mapping, err := loadMapping(*adapter, *mappingFile)
	if err != nil {
		return err
	}
	if mapping.Distill == nil {
		return fmt.Errorf("%s has no non-interactive CLI in its mapping, so there is nothing to ask", *adapter)
	}

	waiting, err := candidate.Read(s.Root())
	if err != nil {
		return err
	}
	var chosen *candidate.Candidate
	for i := range waiting {
		if waiting[i].Key() == *key {
			chosen = &waiting[i]
			break
		}
	}
	if chosen == nil {
		return fmt.Errorf("no candidate with key %s", *key)
	}
	if strings.TrimSpace(chosen.Title) != "" {
		if !*asJSON {
			fmt.Fprintf(e.stdout, "%s already has a title\n", *key)
		}
		return nil
	}
	if strings.TrimSpace(chosen.Body) == "" {
		return fmt.Errorf("%s has nothing to read", *key)
	}

	if *dryRun {
		fmt.Fprintf(e.stdout, "%s %s\n\n%s\n\n--- on stdin ---\n%s\n",
			mapping.Distill.Command,
			strings.Join(renderArgs(mapping.Distill, instruction), " "),
			instruction, chosen.Body)
		return nil
	}

	reply, took, err := runDistill(mapping.Distill, chosen.Body)
	title := oneTitle(reply)

	if cfg.Bool("measure.ledger") {
		entry := measure.Entry{
			Time: time.Now().UTC().Truncate(time.Second), Session: chosen.Session,
			Event: measure.EventDistill, Bytes: len(instruction) + len(chosen.Body),
			HookMS: int(took / time.Millisecond),
		}
		entry.EstTokens = measure.EstimateTokensFromBytes(entry.Bytes + len(reply))
		_ = measure.AppendEntry(s.Root(), entry)
	}
	if err != nil {
		return err
	}
	if title == "" {
		if !*asJSON {
			fmt.Fprintf(e.stdout, "%s: nothing worth a title\n", *key)
		}
		return nil
	}

	chosen.Title = title
	chosen.Source = record.SourceDistill
	if err := candidate.Write(s.Root(), waiting); err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(e.stdout, struct {
			Key   string `json:"key"`
			Title string `json:"title"`
		}{*key, title})
	}
	fmt.Fprintf(e.stdout, "%s  %s\n", *key, title)
	return nil
}

// executable is where the child comes from. It is a variable so that a test
// can run the real child without needing this binary to be installed.
var executable = os.Executable

// renderArgs substitutes the instruction into the mapping's arguments.
func renderArgs(d *agent.Distill, text string) []string {
	out := make([]string, 0, len(d.Args))
	for _, a := range d.Args {
		out = append(out, strings.ReplaceAll(a, "{{instruction}}", text))
	}
	return out
}

// runDistill starts the agent's own CLI and waits for one line.
func runDistill(d *agent.Distill, material string) (string, time.Duration, error) {
	timeout := time.Duration(d.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := renderArgs(d, instruction)
	if d.Material == "arg" {
		args = append(args, material)
	}
	cmd := exec.CommandContext(ctx, d.Command, args...)
	if d.Material != "arg" {
		cmd.Stdin = strings.NewReader(material)
	}
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut

	started := time.Now()
	err := cmd.Run()
	took := time.Since(started)
	if err != nil {
		return "", took, fmt.Errorf("%s: %w: %s", d.Command, err,
			strings.TrimSpace(firstLines(errOut.String(), 3)))
	}
	return out.String(), took, nil
}

// oneTitle takes a model's answer and keeps only what a title may be.
//
// Everything here is distrust of the shape, not of the content: a model asked
// for one line sometimes sends three, or wraps it in quotes, or explains
// itself first. Redaction runs again because a title is about to be written,
// and a write is where redaction belongs (ADR-0013).
func oneTitle(reply string) string {
	line := strings.TrimSpace(reply)
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = strings.TrimSpace(line[:i])
	}
	line = strings.Trim(line, "\"'` ")
	line = strings.TrimSuffix(line, ".")
	if line == "" || strings.EqualFold(line, "SKIP") {
		return ""
	}
	if len(line) > maxTitle {
		return ""
	}
	clean, _ := redact.Text(line)
	return strings.TrimSpace(clean)
}

// startDistill launches cmdDistill in a child nothing waits for.
//
// The hook must exit inside its deadline (K7), and a model call is seconds at
// best. So the child is started and abandoned: its streams are detached,
// because inheriting the agent's pipes would keep them open and an agent
// waiting on EOF would hang — the one way this feature could break the thing
// it is meant to help.
func startDistill(s *store.Store, key string) {
	self, err := executable()
	if err != nil {
		return
	}
	cmd := exec.Command(self, "distill", "--key", key, "--dir", s.Root())
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	if err := cmd.Start(); err != nil {
		return
	}
	// Released rather than waited for. The child outlives this process and
	// is reparented; its result reaches the person through the review queue.
	_ = cmd.Process.Release()
}
