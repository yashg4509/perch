package stackstatus

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/customstatus"
	"github.com/yashg4509/perch/internal/provider"
)

// EnvReport is the JSON shape for `perch status --json` (structure-first milestone).
type EnvReport struct {
	Env   string       `json:"env"`
	Nodes []NodeReport `json:"nodes"`
}

// NodeReport is one row in [EnvReport.Nodes] (extends as status fixtures grow).
type NodeReport struct {
	Name         string                   `json:"name"`
	Provider     string                   `json:"provider"`
	Healthy      bool                     `json:"healthy"`
	ErrorRate    *float64                 `json:"error_rate,omitempty"`
	LastDeploy   *provider.DeploySnapshot `json:"last_deploy,omitempty"`
	DailyTokens  *int64                   `json:"daily_tokens,omitempty"`
	DailyCostUSD *float64                 `json:"daily_cost_usd,omitempty"`
	RecentErrors []string                 `json:"recent_errors,omitempty"`
}

// Collect walks every node in cfg.Environments[env] in sorted name order and resolves status.
// Read-only providers use [provider.ReadOnlyStub] with canned healthy JSON; deployable providers
// return a placeholder status until the runtime is wired (healthy=false).
func Collect(ctx context.Context, cfg *config.Config, env string, reg *provider.Registry) (*EnvReport, error) {
	if cfg == nil {
		return nil, fmt.Errorf("stackstatus: nil config")
	}
	if reg == nil {
		return nil, fmt.Errorf("stackstatus: nil registry")
	}
	nodes, ok := cfg.Environments[env]
	if !ok {
		return nil, fmt.Errorf("stackstatus: unknown environment %q", env)
	}
	names := make([]string, 0, len(nodes))
	for name := range nodes {
		names = append(names, name)
	}
	sort.Strings(names)

	out := &EnvReport{Env: env, Nodes: make([]NodeReport, 0, len(names))}
	for _, name := range names {
		n := nodes[name]
		st, err := statusForNode(ctx, reg, name, n)
		if err != nil {
			return nil, err
		}
		out.Nodes = append(out.Nodes, nodeReportFromStatus(name, n.Provider, st))
	}
	return out, nil
}

func nodeReportFromStatus(name, prov string, st provider.NodeStatus) NodeReport {
	return NodeReport{
		Name:         name,
		Provider:     prov,
		Healthy:      st.Healthy,
		ErrorRate:    st.ErrorRate,
		LastDeploy:   st.LastDeploy,
		DailyTokens:  st.DailyTokens,
		DailyCostUSD: st.DailyCostUSD,
		RecentErrors: st.RecentErrors,
	}
}

func statusForNode(ctx context.Context, reg *provider.Registry, name string, n config.Node) (provider.NodeStatus, error) {
	if n.Provider == "custom" {
		if strings.TrimSpace(n.Status) == "" {
			return provider.NodeStatus{}, fmt.Errorf("stackstatus: node %q: custom provider needs status command", name)
		}
		return customstatus.Run(ctx, n.Status)
	}
	spec := reg.ByName[n.Provider]
	if spec == nil {
		return provider.NodeStatus{}, fmt.Errorf("stackstatus: unknown provider %q for node %q", n.Provider, name)
	}
	pn := provider.Node{
		Name:     name,
		Provider: n.Provider,
		Fields:   nodeFields(n),
	}
	if !spec.Deployable {
		stub := provider.NewReadOnlyStub(spec, provider.ReadOnlyStubOptions{
			StatusBody: []byte(`{"healthy":true}`),
		})
		return stub.Status(ctx, pn)
	}
	return provider.NodeStatus{Healthy: false}, nil
}

func nodeFields(n config.Node) map[string]string {
	m := make(map[string]string)
	if n.Project != "" {
		m["project"] = n.Project
	}
	if n.Service != "" {
		m["service"] = n.Service
	}
	return m
}
