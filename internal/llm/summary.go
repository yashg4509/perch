// Package llm aggregates token/cost hints from LLM nodes (spec slice; `perch llm` later).
package llm

// NodeSpend is token/cost data for one LLM-class node (from status or usage APIs).
type NodeSpend struct {
	Name         string
	DailyTokens  *int64
	DailyCostUSD *float64
}

// Summary rolls up spend across nodes (spec: context / `perch llm costs` precursor).
type Summary struct {
	TotalTokens  int64
	TotalCostUSD float64
	NodeCount    int
}

// Summarize adds metrics only where pointers are non-nil.
func Summarize(nodes []NodeSpend) Summary {
	var s Summary
	for _, n := range nodes {
		if n.DailyTokens == nil && n.DailyCostUSD == nil {
			continue
		}
		s.NodeCount++
		if n.DailyTokens != nil {
			s.TotalTokens += *n.DailyTokens
		}
		if n.DailyCostUSD != nil {
			s.TotalCostUSD += *n.DailyCostUSD
		}
	}
	return s
}
