package agent

import (
	_ "embed"
	"fmt"
)

//go:embed mappings/claude-code.json
var claudeCodeMapping []byte

// Built returns a mapping that ships with this binary.
//
// The mappings are data (K5) and can be replaced on disk without rebuilding,
// which is the point: the parts of an agent's format that documentation does
// not pin down are the parts most likely to need a correction, and a
// correction here is a JSON edit rather than a release.
func Built(name string) (*Mapping, error) {
	switch name {
	case "claude-code":
		return ParseMapping(claudeCodeMapping)
	}
	return nil, fmt.Errorf("agent: no built-in mapping for %q", name)
}
