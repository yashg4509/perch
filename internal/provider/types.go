package provider

// Node is the runtime view of a stack node (from perch.yaml) passed into provider calls.
type Node struct {
	Name     string
	Provider string
	Fields   map[string]string
}

// LogOpts configures log streaming.
type LogOpts struct {
	Follow bool
}

// LogLine is one line of log output.
type LogLine struct {
	Text string
}

// DeploySnapshot is optional deploy metadata in status / context JSON (spec examples).
type DeploySnapshot struct {
	SHA string `json:"sha"`
	Ago string `json:"ago"`
}

// NodeStatus is unmarshaled from provider status API responses and test fixtures.
type NodeStatus struct {
	Healthy      bool            `json:"healthy"`
	ErrorRate    *float64        `json:"error_rate,omitempty"`
	LastDeploy   *DeploySnapshot `json:"last_deploy,omitempty"`
	DailyTokens  *int64          `json:"daily_tokens,omitempty"`
	DailyCostUSD *float64        `json:"daily_cost_usd,omitempty"`
	RecentErrors []string        `json:"recent_errors,omitempty"`
}

// DeployResult identifies a deployment action.
type DeployResult struct {
	ID string
}

// EnvVar is a single environment variable.
type EnvVar struct {
	Key, Value string
}

// Resource is a selectable remote resource (e.g. project picker during init).
type Resource struct {
	Name string
	ID   string
}
