package stackcontext

import (
	"sort"
	"time"

	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/provider"
	"github.com/yashg4509/perch/internal/stackstatus"
)

// Report is the loose JSON shape for `perch context --json` (extensible fields).
type Report struct {
	GeneratedAt  string                  `json:"generated_at"`
	Stack        string                  `json:"stack"`
	Environment  string                  `json:"environment"`
	Summary      string                  `json:"summary,omitempty"`
	Nodes        []Node                  `json:"nodes"`
	StatusReport *stackstatus.EnvReport  `json:"-"` // for agent text formatting only
}

// Node merges topology from the graph with live status rows.
type Node struct {
	Name         string                   `json:"name"`
	Provider     string                   `json:"provider"`
	Deployable   bool                     `json:"deployable"`
	Healthy      bool                     `json:"healthy"`
	StatusSource string                   `json:"status_source,omitempty"`
	Detail       string                   `json:"detail,omitempty"`
	Configured   bool                     `json:"configured"`
	ErrorRate    *float64                 `json:"error_rate,omitempty"`
	LastDeploy   *provider.DeploySnapshot `json:"last_deploy,omitempty"`
	DailyTokens  *int64                   `json:"daily_tokens,omitempty"`
	DailyCostUSD *float64                 `json:"daily_cost_usd,omitempty"`
	RecentErrors []string                 `json:"recent_errors,omitempty"`
}

// Build merges graph topology with `perch status`-style rows for agent-facing JSON.
func Build(at time.Time, g *graph.Graph, rep *stackstatus.EnvReport) *Report {
	out := &Report{
		GeneratedAt: at.UTC().Format(time.RFC3339),
	}
	if g == nil {
		return out
	}
	out.Stack = g.AppName
	out.Environment = g.Environment

	if rep == nil {
		rep = &stackstatus.EnvReport{Env: g.Environment, Nodes: nil}
	}

	byName := make(map[string]stackstatus.NodeReport, len(rep.Nodes))
	for _, n := range rep.Nodes {
		byName[n.Name] = n
	}

	gnByName := make(map[string]graph.Node, len(g.Nodes))
	for _, n := range g.Nodes {
		gnByName[n.Name] = n
	}
	names := make([]string, 0, len(gnByName))
	for n := range gnByName {
		names = append(names, n)
	}
	sort.Strings(names)

	out.Nodes = make([]Node, 0, len(names))
	for _, name := range names {
		gn := gnByName[name]
		sn, ok := byName[name]
		if !ok {
			sn = stackstatus.NodeReport{Name: name, Provider: gn.Provider, Healthy: false}
		}
		out.Nodes = append(out.Nodes, Node{
			Name:         gn.Name,
			Provider:     gn.Provider,
			Deployable:   gn.Deployable,
			Healthy:      sn.Healthy,
			StatusSource: sn.StatusSource,
			Detail:       sn.Detail,
			Configured:   sn.Configured,
			ErrorRate:    sn.ErrorRate,
			LastDeploy:   sn.LastDeploy,
			DailyTokens:  sn.DailyTokens,
			DailyCostUSD: sn.DailyCostUSD,
			RecentErrors: sn.RecentErrors,
		})
	}
	out.StatusReport = rep
	c := stackstatus.CountGroups(rep)
	out.Summary = stackstatus.FormatSummaryLine(c)
	return out
}
