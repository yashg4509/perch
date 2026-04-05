package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the in-memory representation of perch.yaml (parse-only in T1-001; validate in T1-002).
type Config struct {
	Name         string
	Environments map[string]map[string]Node
	Edges        []Edge
}

// Node is one service definition under an environment.
type Node struct {
	Provider string `yaml:"provider"`
	Project  string `yaml:"project,omitempty"`
	Service  string `yaml:"service,omitempty"`
	Status   string `yaml:"status,omitempty"` // shell command when provider is custom
	Logs     string `yaml:"logs,omitempty"`   // optional log tail command for custom
}

// Edge is a directed dependency between node names (same names as in environments.*).
type Edge struct {
	From string
	To   string
}

type fileConfig struct {
	Name         string                     `yaml:"name"`
	Environments map[string]map[string]Node `yaml:"environments"`
	Edges        []string                   `yaml:"edges"`
}

// Parse decodes perch.yaml bytes into a [Config]. It does not validate semantic rules (T1-002).
func Parse(data []byte) (*Config, error) {
	var f fileConfig
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("yaml: %w", err)
	}
	edges := make([]Edge, 0, len(f.Edges))
	for _, raw := range f.Edges {
		e, err := parseEdge(raw)
		if err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	return &Config{
		Name:         f.Name,
		Environments: f.Environments,
		Edges:        edges,
	}, nil
}

func parseEdge(s string) (Edge, error) {
	parts := strings.Split(s, "->")
	if len(parts) != 2 {
		return Edge{}, fmt.Errorf("config: edge %q: expected 'from -> to'", strings.TrimSpace(s))
	}
	from := strings.TrimSpace(parts[0])
	to := strings.TrimSpace(parts[1])
	if from == "" || to == "" {
		return Edge{}, fmt.Errorf("config: edge %q: empty endpoint", strings.TrimSpace(s))
	}
	return Edge{From: from, To: to}, nil
}
