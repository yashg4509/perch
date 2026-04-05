package graph

import (
	"fmt"
	"sort"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/provider"
)

// Graph is the resolved stack view for one environment.
type Graph struct {
	AppName     string
	Environment string
	Nodes       []Node
	Edges       []config.Edge
}

// Node augments config nodes with provider-registry metadata.
type Node struct {
	Name       string
	Provider   string
	Deployable bool
	Project    string
	Service    string
	Status     string
	Logs       string
}

// Build constructs a graph for the named environment using perch.yaml and the provider registry.
func Build(cfg *config.Config, reg *provider.Registry, env string) (*Graph, error) {
	if reg == nil {
		return nil, fmt.Errorf("graph: nil registry")
	}
	nodes, err := cfg.EnvironmentNodes(env)
	if err != nil {
		return nil, err
	}
	g := &Graph{
		AppName:     cfg.Name,
		Environment: env,
		Edges:       append([]config.Edge(nil), cfg.Edges...),
	}
	names := make([]string, 0, len(nodes))
	for name := range nodes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		cn := nodes[name]
		d, err := deployable(reg, cn.Provider)
		if err != nil {
			return nil, fmt.Errorf("graph: node %q: %w", name, err)
		}
		g.Nodes = append(g.Nodes, Node{
			Name:       name,
			Provider:   cn.Provider,
			Deployable: d,
			Project:    cn.Project,
			Service:    cn.Service,
			Status:     cn.Status,
			Logs:       cn.Logs,
		})
	}
	return g, nil
}

func deployable(reg *provider.Registry, providerID string) (bool, error) {
	if providerID == "custom" {
		return true, nil
	}
	s, ok := reg.ByName[providerID]
	if !ok {
		return false, fmt.Errorf("unknown provider %q", providerID)
	}
	return s.Deployable, nil
}
