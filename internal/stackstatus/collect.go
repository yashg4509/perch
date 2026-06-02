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

// Status source values for [NodeReport.StatusSource].
const (
	SourceShell          = "shell"
	SourcePlaceholder  = "placeholder"
	SourceUnconfigured = "unconfigured"
	SourceUnchecked    = "unchecked"
)

// EnvReport is the JSON shape for `perch status --json` (structure-first milestone).
type EnvReport struct {
	Env   string       `json:"env"`
	Nodes []NodeReport `json:"nodes"`
}

// NodeReport is one row in [EnvReport.Nodes].
type NodeReport struct {
	Name         string                   `json:"name"`
	Provider     string                   `json:"provider"`
	Healthy      bool                     `json:"healthy"`
	StatusSource string                   `json:"status_source"`
	Detail       string                   `json:"detail,omitempty"`
	Configured   bool                     `json:"configured"`
	ErrorRate    *float64                 `json:"error_rate,omitempty"`
	LastDeploy   *provider.DeploySnapshot `json:"last_deploy,omitempty"`
	DailyTokens  *int64                   `json:"daily_tokens,omitempty"`
	DailyCostUSD *float64                 `json:"daily_cost_usd,omitempty"`
	RecentErrors []string                 `json:"recent_errors,omitempty"`
}

// Collect walks every node in cfg.Environments[env] in sorted name order and resolves status.
func Collect(ctx context.Context, cfg *config.Config, env string, reg *provider.Registry, opts CollectOptions) (*EnvReport, error) {
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
	var jobs []probeJob
	for i, name := range names {
		n := nodes[name]
		row, job, err := statusRow(ctx, reg, opts, name, n)
		if err != nil {
			return nil, err
		}
		out.Nodes = append(out.Nodes, row)
		if job != nil {
			job.index = i
			jobs = append(jobs, *job)
		}
	}
	if len(jobs) > 0 {
		probeJobsParallel(ctx, jobs, out.Nodes, opts.ProbeConcurrency)
	}
	return out, nil
}

func statusRow(ctx context.Context, reg *provider.Registry, opts CollectOptions, name string, n config.Node) (NodeReport, *probeJob, error) {
	spec := reg.ByName[n.Provider]
	configured, cfgDetail := nodeConfigured(n, spec, opts)
	row := NodeReport{
		Name:       name,
		Provider:   n.Provider,
		Configured: configured,
	}
	hasCred := credChecker(opts, spec)

	if n.Provider == "custom" {
		if strings.TrimSpace(n.Status) == "" {
			return row, nil, fmt.Errorf("stackstatus: node %q: custom provider needs status command", name)
		}
		st, err := customstatus.Run(ctx, n.Status)
		if err != nil {
			return row, nil, err
		}
		row.Healthy = st.Healthy
		row.StatusSource = SourceShell
		row.ErrorRate = st.ErrorRate
		row.LastDeploy = st.LastDeploy
		row.DailyTokens = st.DailyTokens
		row.DailyCostUSD = st.DailyCostUSD
		row.RecentErrors = st.RecentErrors
		if !row.Healthy {
			row.Detail = "custom health command failed"
		}
		return row, nil, nil
	}

	if spec == nil {
		return row, nil, fmt.Errorf("stackstatus: unknown provider %q for node %q", n.Provider, name)
	}

	if !configured {
		row.Healthy = false
		row.StatusSource = SourceUnconfigured
		row.Detail = cfgDetail
		return row, nil, nil
	}

	if spec.Deployable {
		row.Healthy = false
		row.StatusSource = SourcePlaceholder
		row.Detail = "deployable host status API not implemented yet"
		return row, nil, nil
	}

	env := opts.ProjectEnv
	if env == nil {
		env = map[string]string{}
	}
	if configuredViaAppEnv(spec, env, hasCred) {
		row.Healthy = true
		row.StatusSource = SourceAppEnv
		row.Detail = appEnvDetail(spec)
		return row, nil, nil
	}

	if token, ok := shouldProbeAPI(spec, opts, hasCred); ok {
		row.Healthy = false
		row.StatusSource = SourceUnchecked
		return row, &probeJob{
			name:       name,
			spec:       spec,
			node:       n,
			token:      token,
			projectEnv: env,
		}, nil
	}

	if spec.StatusProbeMode() == "none" {
		row.Healthy = true
		row.StatusSource = SourceUnchecked
		row.Detail = "no HTTP status probe (use CLI or app .env)"
		return row, nil, nil
	}
	row.Healthy = false
	row.StatusSource = SourceUnchecked
	row.Detail = "credential present; probe not scheduled"
	return row, nil, nil
}

func credChecker(opts CollectOptions, spec *provider.Spec) func(string) (bool, error) {
	if spec == nil || spec.Credentials.Key == "" {
		return nil
	}
	return func(key string) (bool, error) {
		if opts.CredStore == nil {
			return false, nil
		}
		v, ok, err := opts.CredStore.Get(key)
		if err != nil {
			return false, err
		}
		return ok && strings.TrimSpace(v) != "", nil
	}
}

func nodeConfigured(n config.Node, spec *provider.Spec, opts CollectOptions) (bool, string) {
	env := opts.ProjectEnv
	if env == nil {
		env = map[string]string{}
	}
	return config.ProviderConfigured(n, spec, env, credChecker(opts, spec))
}

// NodeReportFromStatus maps provider status into a row (used by tests and fixtures).
func NodeReportFromStatus(name, prov string, st provider.NodeStatus) NodeReport {
	return NodeReport{
		Name:         name,
		Provider:     prov,
		Healthy:      st.Healthy,
		StatusSource: SourceUnchecked,
		Configured:   true,
		ErrorRate:    st.ErrorRate,
		LastDeploy:   st.LastDeploy,
		DailyTokens:  st.DailyTokens,
		DailyCostUSD: st.DailyCostUSD,
		RecentErrors: st.RecentErrors,
	}
}
