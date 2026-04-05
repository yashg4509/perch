package config

import (
	"fmt"
	"strings"
)

// Validate checks semantic rules for a parsed [Config]. It does not re-parse YAML.
func Validate(c *Config) error {
	if c == nil {
		return fmt.Errorf("config: nil config")
	}
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("config: name is required")
	}
	if len(c.Environments) == 0 {
		return fmt.Errorf("config: environments is required")
	}
	for envName, nodes := range c.Environments {
		if len(nodes) == 0 {
			return fmt.Errorf("config: environment %q must declare at least one node", envName)
		}
		seen := make(map[string]string, len(nodes))
		for nodeName, node := range nodes {
			if strings.TrimSpace(node.Provider) == "" {
				return fmt.Errorf("config: node %q in environment %q: provider is required", nodeName, envName)
			}
			if node.Provider == "custom" && strings.TrimSpace(node.Status) == "" {
				return fmt.Errorf("config: node %q in environment %q: custom provider requires status command", nodeName, envName)
			}
			key := strings.ToLower(nodeName)
			if prev, ok := seen[key]; ok {
				return fmt.Errorf("config: environment %q: duplicate node name %q (conflicts with %q)", envName, nodeName, prev)
			}
			seen[key] = nodeName
		}
	}
	for _, e := range c.Edges {
		for envName, nodes := range c.Environments {
			if _, ok := nodes[e.From]; !ok {
				return fmt.Errorf("config: edge %q -> %q: unknown node %q in environment %q", e.From, e.To, e.From, envName)
			}
			if _, ok := nodes[e.To]; !ok {
				return fmt.Errorf("config: edge %q -> %q: unknown node %q in environment %q", e.From, e.To, e.To, envName)
			}
		}
	}
	return nil
}
