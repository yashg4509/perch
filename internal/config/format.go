package config

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// FormatYAML serializes cfg to perch.yaml bytes (canonical ordering; comments are not preserved).
func FormatYAML(cfg *Config) ([]byte, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config: nil config")
	}
	envKeys := sortedEnvKeys(cfg.Environments)
	outEnvs := make(map[string]map[string]Node, len(envKeys))
	for _, ek := range envKeys {
		inner := cfg.Environments[ek]
		nodeKeys := sortedNodeKeys(inner)
		m := make(map[string]Node, len(nodeKeys))
		for _, nk := range nodeKeys {
			m[nk] = inner[nk]
		}
		outEnvs[ek] = m
	}
	edgeStrs := make([]string, 0, len(cfg.Edges))
	for _, e := range cfg.Edges {
		edgeStrs = append(edgeStrs, fmt.Sprintf("%s -> %s", e.From, e.To))
	}
	sort.Strings(edgeStrs)
	doc := struct {
		Name         string                     `yaml:"name"`
		Environments map[string]map[string]Node `yaml:"environments"`
		Edges        []string                   `yaml:"edges"`
	}{
		Name:         cfg.Name,
		Environments: outEnvs,
		Edges:        edgeStrs,
	}
	return yaml.Marshal(&doc)
}

func sortedEnvKeys(m map[string]map[string]Node) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func sortedNodeKeys(m map[string]Node) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
