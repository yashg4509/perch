package tui_test

import (
	"testing"

	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/tui"
)

func TestNodeFocusOrder_leftToRightLayers(t *testing.T) {
	g := &graph.Graph{
		Nodes: []graph.Node{
			{Name: "web", Provider: "vercel", Deployable: true},
			{Name: "api", Provider: "render", Deployable: true},
			{Name: "db", Provider: "supabase", Deployable: true},
		},
		Edges: []config.Edge{{From: "web", To: "api"}, {From: "api", To: "db"}},
	}
	got := tui.NodeFocusOrder(g)
	want := []string{"web", "api", "db"}
	if len(got) != len(want) {
		t.Fatalf("len %d vs %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("idx %d: got %q want %q (full %v)", i, got[i], want[i], got)
		}
	}
}

func TestNodeFocusOrder_sameLayerSorted(t *testing.T) {
	g := &graph.Graph{
		Nodes: []graph.Node{
			{Name: "z", Provider: "custom", Deployable: true},
			{Name: "a", Provider: "custom", Deployable: true},
		},
		Edges: nil,
	}
	got := tui.NodeFocusOrder(g)
	// single column, sorted names
	if len(got) != 2 || got[0] != "a" || got[1] != "z" {
		t.Fatalf("got %v", got)
	}
}
