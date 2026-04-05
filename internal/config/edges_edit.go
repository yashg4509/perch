package config

import (
	"fmt"
	"strings"
)

// AddEdge appends from -> to if it is not already present (case-sensitive node names).
// Runs [Validate] after mutation.
func AddEdge(cfg *Config, from, to string) error {
	if cfg == nil {
		return fmt.Errorf("config: nil config")
	}
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if from == "" || to == "" {
		return fmt.Errorf("config: edge: empty endpoint")
	}
	if from == to {
		return fmt.Errorf("config: edge: self-loop %q -> %q is not allowed", from, to)
	}
	for _, e := range cfg.Edges {
		if e.From == from && e.To == to {
			return Validate(cfg)
		}
	}
	cfg.Edges = append(cfg.Edges, Edge{From: from, To: to})
	return Validate(cfg)
}

// RemoveEdge deletes every edge matching from -> to. Returns how many were removed.
func RemoveEdge(cfg *Config, from, to string) int {
	if cfg == nil || len(cfg.Edges) == 0 {
		return 0
	}
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	var keep []Edge
	n := 0
	for _, e := range cfg.Edges {
		if e.From == from && e.To == to {
			n++
			continue
		}
		keep = append(keep, e)
	}
	cfg.Edges = keep
	return n
}
