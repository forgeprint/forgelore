// Package config resolves settings from the five sources of ADR-0018.
//
// Highest first: a command-line flag, an environment variable, the
// individual's config, the team's config, the built-in default. A setting's
// effective value is therefore not visible in any one file, which is why
// every value carries where it came from and `doctor` prints it.
//
// The file dialect is the restricted YAML of ADR-0017, with dotted keys for
// grouping. An unknown key is a warning and never an error (ADR-0011): a
// config written for a newer Forgelore has to keep working on an older one.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/forgeprint/forgelore/internal/record"
)

// Source is where a value came from.
type Source string

const (
	SourceDefault Source = "default"
	SourceTeam    Source = "team"
	SourceLocal   Source = "local"
	SourceEnv     Source = "env"
	SourceFlag    Source = "flag"
)

// File names inside a .forgelore directory.
const (
	teamFile  = "config.yaml"
	localFile = "local/config.yaml"
)

// envPrefix turns a key into an environment variable:
// measure.ab.control_percent becomes FORGELORE_MEASURE_AB_CONTROL_PERCENT.
const envPrefix = "FORGELORE_"

// A spec is a setting this version understands. Only settings with a consumer
// are listed: a key that nothing reads is a promise the tool has not made.
type spec struct {
	key  string
	kind record.Kind
	def  record.Value
	doc  string
}

var known = []spec{
	{
		key:  "measure.ledger",
		kind: record.KindBool,
		def:  record.Value{Kind: record.KindBool, Bool: true},
		doc:  "record injections in the local ledger",
	},
	{
		key:  "measure.ab.control_percent",
		kind: record.KindInt,
		def:  record.Value{Kind: record.KindInt, Int: 0},
		doc:  "percentage of sessions assigned to the control group; 0 disables A/B",
	},
	{
		key:  "measure.ab.salt",
		kind: record.KindString,
		def:  record.Value{Kind: record.KindString, Str: ""},
		doc:  "changes every session's A/B assignment, for starting a fresh trial",
	},
	{
		key:  "report.period_days",
		kind: record.KindInt,
		def:  record.Value{Kind: record.KindInt, Int: 30},
		doc:  "how far back `report` looks by default",
	},
}

func specFor(key string) (spec, bool) {
	for _, s := range known {
		if s.key == key {
			return s, true
		}
	}
	return spec{}, false
}

// Keys returns every setting this version understands, with its default and
// what it does. `doctor` prints these.
func Keys() []struct{ Key, Default, Doc string } {
	out := make([]struct{ Key, Default, Doc string }, 0, len(known))
	for _, s := range known {
		out = append(out, struct{ Key, Default, Doc string }{s.key, formatValue(s.def), s.doc})
	}
	return out
}

// Problem is something wrong with a config file that was not bad enough to
// stop. Every one of these is reported by `doctor` and ignored otherwise.
type Problem struct {
	Path string
	Line int
	Msg  string
}

func (p Problem) String() string {
	if p.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", p.Path, p.Line, p.Msg)
	}
	return fmt.Sprintf("%s: %s", p.Path, p.Msg)
}

// resolved is one setting with its provenance.
type resolved struct {
	value  record.Value
	source Source
	from   string
}

// Config holds the effective settings.
type Config struct {
	values map[string]resolved
}

// Load resolves the settings for the project rooted at a .forgelore
// directory. lookupEnv and overrides are passed in rather than read from the
// process, so that a test can exercise all five sources.
//
// A file that cannot be parsed is skipped with a problem, not an error: a
// broken config must not stop a recall (ADR-0007).
func Load(root string, lookupEnv func(string) (string, bool), overrides map[string]string) (*Config, []Problem) {
	c := &Config{values: make(map[string]resolved, len(known))}
	var problems []Problem

	for _, s := range known {
		c.values[s.key] = resolved{value: s.def, source: SourceDefault, from: "built in"}
	}

	// Lowest precedence first, so each source overwrites the one before it.
	for _, f := range []struct {
		path   string
		source Source
	}{
		{filepath.Join(root, teamFile), SourceTeam},
		{filepath.Join(root, localFile), SourceLocal},
	} {
		problems = append(problems, c.applyFile(f.path, f.source)...)
	}

	if lookupEnv != nil {
		for _, s := range known {
			name := EnvName(s.key)
			raw, ok := lookupEnv(name)
			if !ok {
				continue
			}
			value, err := coerce(raw, s.kind)
			if err != nil {
				problems = append(problems, Problem{Path: name, Msg: err.Error()})
				continue
			}
			c.values[s.key] = resolved{value: value, source: SourceEnv, from: name}
		}
	}

	for key, raw := range overrides {
		s, ok := specFor(key)
		if !ok {
			problems = append(problems, Problem{Path: "flag", Msg: fmt.Sprintf("unknown setting %q", key)})
			continue
		}
		value, err := coerce(raw, s.kind)
		if err != nil {
			problems = append(problems, Problem{Path: "flag " + key, Msg: err.Error()})
			continue
		}
		c.values[key] = resolved{value: value, source: SourceFlag, from: "flag"}
	}

	return c, problems
}

// applyFile reads one config file. A missing file is the normal case and not
// a problem.
func (c *Config) applyFile(path string, source Source) []Problem {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return []Problem{{Path: path, Msg: err.Error()}}
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	doc, err := record.ParseDoc(lines, 1)
	if err != nil {
		return []Problem{{Path: path, Msg: err.Error()}}
	}

	var problems []Problem
	for _, f := range doc.Fields {
		s, ok := specFor(f.Key)
		if !ok {
			problems = append(problems, Problem{
				Path: path,
				Msg:  fmt.Sprintf("unknown setting %q, ignored", f.Key),
			})
			continue
		}
		if f.Value.Kind != s.kind {
			problems = append(problems, Problem{
				Path: path,
				Msg:  fmt.Sprintf("%s wants %s, got %s; ignored", f.Key, kindName(s.kind), kindName(f.Value.Kind)),
			})
			continue
		}
		c.values[f.Key] = resolved{value: f.Value, source: source, from: path}
	}
	return problems
}

// EnvName returns the environment variable that overrides a setting.
func EnvName(key string) string {
	return envPrefix + strings.ToUpper(strings.NewReplacer(".", "_", "-", "_").Replace(key))
}

// Int returns a setting this version understands. Asking for a key that is
// not in the table is a programming error, not a user error, so it panics
// rather than returning a zero that would quietly change behaviour.
func (c *Config) Int(key string) int64 {
	return c.must(key, record.KindInt).Int
}

func (c *Config) Bool(key string) bool {
	return c.must(key, record.KindBool).Bool
}

func (c *Config) String(key string) string {
	return c.must(key, record.KindString).Str
}

func (c *Config) must(key string, kind record.Kind) record.Value {
	r, ok := c.values[key]
	if !ok {
		panic("config: no setting named " + key)
	}
	if r.value.Kind != kind {
		panic("config: " + key + " is not " + kindName(kind))
	}
	return r.value
}

// Origin returns where a setting's value came from, for `doctor`.
func (c *Config) Origin(key string) (Source, string) {
	r, ok := c.values[key]
	if !ok {
		return "", ""
	}
	return r.source, r.from
}

// Effective returns every setting with its value and provenance, in the order
// they are declared.
func (c *Config) Effective() []struct{ Key, Value, Source, From string } {
	out := make([]struct{ Key, Value, Source, From string }, 0, len(known))
	for _, s := range known {
		r := c.values[s.key]
		out = append(out, struct{ Key, Value, Source, From string }{
			s.key, formatValue(r.value), string(r.source), r.from,
		})
	}
	return out
}

func coerce(raw string, kind record.Kind) (record.Value, error) {
	switch kind {
	case record.KindInt:
		n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil {
			return record.Value{}, fmt.Errorf("%q is not an integer", raw)
		}
		return record.Value{Kind: record.KindInt, Int: n}, nil
	case record.KindBool:
		switch strings.TrimSpace(raw) {
		case "true":
			return record.Value{Kind: record.KindBool, Bool: true}, nil
		case "false":
			return record.Value{Kind: record.KindBool, Bool: false}, nil
		}
		return record.Value{}, fmt.Errorf("%q is not true or false", raw)
	default:
		return record.Value{Kind: record.KindString, Str: raw}, nil
	}
}

func formatValue(v record.Value) string {
	switch v.Kind {
	case record.KindInt:
		return strconv.FormatInt(v.Int, 10)
	case record.KindBool:
		return strconv.FormatBool(v.Bool)
	case record.KindList:
		return "[" + strings.Join(v.List, ", ") + "]"
	default:
		if v.Str == "" {
			return `""`
		}
		return v.Str
	}
}

func kindName(k record.Kind) string {
	switch k {
	case record.KindInt:
		return "an integer"
	case record.KindBool:
		return "a boolean"
	case record.KindList:
		return "a list"
	default:
		return "a string"
	}
}
