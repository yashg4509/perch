package graph

import "github.com/yashg4509/perch/internal/config"

func nodeConfiguredForGraph(n Node) bool {
	if n.Provider == "custom" {
		return n.Status != ""
	}
	if n.Project != "" && !config.IsPlaceholder(n.Project) {
		return true
	}
	if n.Service != "" && !config.IsPlaceholder(n.Service) {
		return true
	}
	return false
}
