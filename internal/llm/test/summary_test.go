package llm_test

import (
	"testing"

	"github.com/yashg4509/perch/internal/llm"
)

func TestSummarize_fromNodeHints(t *testing.T) {
	s := llm.Summarize([]llm.NodeSpend{
		{Name: "openai", DailyTokens: ip(1000), DailyCostUSD: fp(0.01)},
		{Name: "anthropic", DailyTokens: ip(2000), DailyCostUSD: fp(0.02)},
	})
	if s.NodeCount != 2 {
		t.Fatal(s)
	}
	if s.TotalTokens != 3000 {
		t.Fatal(s.TotalTokens)
	}
	if s.TotalCostUSD < 0.029 || s.TotalCostUSD > 0.031 {
		t.Fatal(s.TotalCostUSD)
	}
}

func TestSummarize_skipsNilMetrics(t *testing.T) {
	s := llm.Summarize([]llm.NodeSpend{{Name: "x"}, {Name: "y", DailyTokens: ip(5)}})
	if s.TotalTokens != 5 || s.NodeCount != 1 {
		t.Fatalf("%+v", s)
	}
}

func ip(v int64) *int64     { return &v }
func fp(v float64) *float64 { return &v }
