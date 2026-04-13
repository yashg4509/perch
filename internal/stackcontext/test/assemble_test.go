package stackcontext_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/stackcontext"
	"github.com/yashg4509/perch/internal/stackstatus"
)

func TestBuild_mergesGraphAndStatus(t *testing.T) {
	at := time.Date(2026, 4, 3, 14, 22, 0, 0, time.UTC)
	g := &graph.Graph{
		AppName:     "my-app",
		Environment: "production",
		Nodes: []graph.Node{
			{Name: "frontend", Provider: "vercel", Deployable: true, Project: "p"},
			{Name: "llm", Provider: "openai", Deployable: false},
		},
		Edges: []config.Edge{{From: "frontend", To: "llm"}},
	}
	rep := &stackstatus.EnvReport{
		Env: "production",
		Nodes: []stackstatus.NodeReport{
			{Name: "frontend", Provider: "vercel", Healthy: true},
			{Name: "llm", Provider: "openai", Healthy: true, DailyTokens: ptrI64(142000), DailyCostUSD: ptrF(0.43)},
		},
	}
	r := stackcontext.Build(at, g, rep)
	if r.Stack != "my-app" || r.Environment != "production" {
		t.Fatalf("%+v", r)
	}
	if r.GeneratedAt != "2026-04-03T14:22:00Z" {
		t.Fatal(r.GeneratedAt)
	}
	if len(r.Nodes) != 2 {
		t.Fatal(len(r.Nodes))
	}
	var llm *stackcontext.Node
	for i := range r.Nodes {
		if r.Nodes[i].Name == "llm" {
			llm = &r.Nodes[i]
			break
		}
	}
	if llm == nil || !llm.Healthy || llm.DailyTokens == nil || *llm.DailyTokens != 142000 {
		t.Fatalf("%+v", llm)
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["nodes"]; !ok {
		t.Fatal(m)
	}
}

func ptrI64(v int64) *int64   { return &v }
func ptrF(v float64) *float64 { return &v }
