package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Marshal encodes cfg as perch.yaml bytes.
func Marshal(cfg *Config) ([]byte, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config: nil")
	}
	edgeStrs := make([]string, len(cfg.Edges))
	for i, e := range cfg.Edges {
		edgeStrs[i] = fmt.Sprintf("%s -> %s", e.From, e.To)
	}
	f := fileConfig{
		Name:         cfg.Name,
		Environments: cfg.Environments,
		Edges:        edgeStrs,
	}
	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(&f); err != nil {
		return nil, err
	}
	return []byte(b.String()), nil
}
