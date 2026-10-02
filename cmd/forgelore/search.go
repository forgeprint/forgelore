package main

import (
	"fmt"
	"strings"

	"github.com/forgeprint/forgelore/internal/record"
)

// cmdSearch returns ids, types and titles and nothing else. The body is not
// here and cannot be asked for in bulk: the only way to a body is `show` with
// an id that came from a search, which is what keeps a search from costing a
// whole store's worth of context (ADR-0006).
func cmdSearch(e env, args []string) error {
	fs := newFlags(e, "search")
	limit := fs.Int("limit", 20, "how many results at most")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	words, err := parseInterspersed(fs, args)
	if err != nil {
		return err
	}

	query := strings.Join(words, " ")
	if strings.TrimSpace(query) == "" {
		return fmt.Errorf("search needs something to search for")
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	idx, err := openIndex(s)
	if err != nil {
		return err
	}
	defer idx.Close()

	hits, err := idx.Search(query, *limit)
	if err != nil {
		return err
	}
	results := toHitJSON(hits)

	if *asJSON {
		return writeJSON(e.stdout, struct {
			Query   string    `json:"query"`
			Results []hitJSON `json:"results"`
		}{query, results})
	}

	if len(results) == 0 {
		fmt.Fprintf(e.stdout, "nothing matches %q\n", query)
		return nil
	}
	for _, h := range results {
		superseded := ""
		if h.Superseded {
			superseded = "  (superseded)"
		}
		fmt.Fprintf(e.stdout, "%s  %-9s %s%s\n", h.ID, h.Type, h.Title, superseded)
	}
	fmt.Fprintf(e.stdout, "\n%d result(s). Use: forgelore show <id>\n", len(results))
	return nil
}

// cmdShow is the second stage of staged access: one record, in full,
// identified by an id the caller already has.
func cmdShow(e env, args []string) error {
	fs := newFlags(e, "show")
	dir := fs.String("dir", "", "work on the project containing this directory")
	asJSON := fs.Bool("json", false, "machine-readable output")
	words, err := parseInterspersed(fs, args)
	if err != nil {
		return err
	}
	if len(words) != 1 {
		return fmt.Errorf("show needs exactly one record id")
	}

	id, ok := record.NormalizeULID(words[0])
	if !ok {
		return fmt.Errorf("%q is not a record id", words[0])
	}

	s, err := openStore(e, *dir)
	if err != nil {
		return err
	}
	r, err := s.Get(id)
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(e.stdout, struct {
			ID           string   `json:"id"`
			Type         string   `json:"type"`
			Scope        string   `json:"scope"`
			Title        string   `json:"title"`
			Created      string   `json:"created"`
			Source       string   `json:"source"`
			Tainted      bool     `json:"tainted"`
			Fingerprint  string   `json:"fingerprint,omitempty"`
			Tags         []string `json:"tags,omitempty"`
			Related      []string `json:"related,omitempty"`
			SupersededBy string   `json:"superseded_by,omitempty"`
			Body         string   `json:"body"`
		}{
			r.ID, r.Type, string(r.Scope), r.Title,
			r.Created.Format("2006-01-02T15:04:05Z"), r.Source, r.Tainted,
			r.Fingerprint, r.Tags, r.Related, r.SupersededBy, r.Body,
		})
	}

	encoded, err := r.Encode()
	if err != nil {
		return err
	}
	_, err = e.stdout.Write(encoded)
	return err
}
