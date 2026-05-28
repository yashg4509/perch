package graph

import (
	"sort"

	"github.com/yashg4509/perch/internal/config"
)

// JSONReport is the `perch graph --json` payload (topology; health reserved for later).
type JSONReport struct {
	AppName     string           `json:"appName"`
	Environment string           `json:"environment"`
	Providers   []string         `json:"providers"`
	Nodes       []JSONNode       `json:"nodes"`
	Edges       []JSONEdge       `json:"edges"`
	Health      []JSONNodeHealth `json:"health,omitempty"`
}

// JSONNode is one vertex in [JSONReport.Nodes].
type JSONNode struct {
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	Deployable bool   `json:"deployable"`
	Configured bool   `json:"configured"`
	Project    string `json:"project,omitempty"`
	Service    string `json:"service,omitempty"`
	Status     string `json:"status,omitempty"`
	Logs       string `json:"logs,omitempty"`
}

// JSONEdge is a directed dependency.
type JSONEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// JSONNodeHealth is a placeholder for future status wiring (omitted when empty).
type JSONNodeHealth struct {
	Node    string `json:"node"`
	Healthy *bool  `json:"healthy,omitempty"`
}

// NewJSONReport builds a stable, JSON-serializable view of g.
func NewJSONReport(g *Graph) *JSONReport {
	if g == nil {
		return &JSONReport{}
	}
	seen := make(map[string]struct{})
	for _, n := range g.Nodes {
		seen[n.Provider] = struct{}{}
	}
	provs := make([]string, 0, len(seen))
	for p := range seen {
		provs = append(provs, p)
	}
	sort.Strings(provs)

	nodes := make([]JSONNode, len(g.Nodes))
	for i, n := range g.Nodes {
		proj, svc := n.Project, n.Service
		configured := nodeConfiguredForGraph(n)
		if config.IsPlaceholder(proj) {
			proj = ""
		}
		if config.IsPlaceholder(svc) {
			svc = ""
		}
		nodes[i] = JSONNode{
			Name:       n.Name,
			Provider:   n.Provider,
			Deployable: n.Deployable,
			Configured: configured,
			Project:    proj,
			Service:    svc,
			Status:     n.Status,
			Logs:       n.Logs,
		}
	}
	edges := make([]JSONEdge, len(g.Edges))
	for i, e := range g.Edges {
		edges[i] = JSONEdge{From: e.From, To: e.To}
	}
	return &JSONReport{
		AppName:     g.AppName,
		Environment: g.Environment,
		Providers:   provs,
		Nodes:       nodes,
		Edges:       edges,
	}
}
