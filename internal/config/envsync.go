package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadEnvMap reads perch.envmap.yaml: top-level keys are environment names.
func LoadEnvMap(path string) (map[string]map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]map[string]string
	if err := yaml.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("envmap: %w", err)
	}
	if out == nil {
		out = map[string]map[string]string{}
	}
	return out, nil
}

// ApplyEnvMap sets node.project or node.service from dotenv for one environment.
func ApplyEnvMap(cfg *Config, env string, dotenv map[string]string, mappings map[string]string) ([]string, error) {
	if cfg == nil {
		return nil, nil
	}
	nodes, ok := cfg.Environments[env]
	if !ok {
		return nil, fmt.Errorf("config: unknown environment %q", env)
	}
	var applied []string
	for path, envKey := range mappings {
		val := strings.TrimSpace(dotenv[envKey])
		if val == "" {
			continue
		}
		parts := strings.SplitN(path, ".", 2)
		if len(parts) != 2 {
			return applied, fmt.Errorf("envmap: invalid path %q", path)
		}
		nodeName, field := parts[0], parts[1]
		n, ok := nodes[nodeName]
		if !ok {
			return applied, fmt.Errorf("envmap: unknown node %q", nodeName)
		}
		switch field {
		case "project":
			n.Project = val
		case "service":
			n.Service = val
		default:
			return applied, fmt.Errorf("envmap: unsupported field %q", field)
		}
		nodes[nodeName] = n
		applied = append(applied, path)
	}
	cfg.Environments[env] = nodes
	return applied, nil
}

// DevLocalDefaults sets non-secret project ids for local dev when still placeholder.
var DevLocalDefaults = map[string]string{
	"auth":      "perch-brief",
	"openai":    "default",
	"anthropic": "default",
	"billing":   "local-dev",
	"email":     "local-dev",
	"errors":    "local-dev",
	"jobs":      "local-dev",
	"logs":      "local-dev",
	"analytics": "local-dev",
	"metrics":   "local-dev",
	"repo":      "local-dev",
	"traces":    "local-dev",
	"queue":     "local-dev",
	"redis":     "local-dev",
}

// ApplyDevDefaults fills placeholder projects on dev with local-dev style ids.
func ApplyDevDefaults(cfg *Config, env string) []string {
	if cfg == nil || env != "dev" {
		return nil
	}
	nodes, ok := cfg.Environments[env]
	if !ok {
		return nil
	}
	var applied []string
	for name, def := range DevLocalDefaults {
		n, ok := nodes[name]
		if !ok {
			continue
		}
		if IsPlaceholder(n.Project) {
			n.Project = def
			nodes[name] = n
			applied = append(applied, name+".project")
		}
	}
	cfg.Environments[env] = nodes
	return applied
}
