package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/graph"
)

func TestStackChrome_includesSummaryAndFlows(t *testing.T) {
	g := &graph.Graph{
		AppName:     "app",
		Environment: "production",
		Nodes: []graph.Node{
			{Name: "web", Provider: "vercel", Deployable: true},
			{Name: "api", Provider: "openai", Deployable: false},
		},
		Edges: []config.Edge{{From: "web", To: "api"}},
	}
	s := StackChrome(g, 120, true, nil, "")
	if !strings.Contains(s, "2 nodes") || !strings.Contains(s, "1 deployable") {
		t.Fatalf("missing summary: %q", s)
	}
	if !strings.Contains(s, "web→api") {
		t.Fatalf("missing flow: %q", s)
	}
	if !strings.Contains(s, "perch edge") || !strings.Contains(s, "status --json") || !strings.Contains(s, "? palette") {
		t.Fatalf("missing tips: %q", s)
	}
}

func TestStackModel_cycleEnvOnE(t *testing.T) {
	g0 := &graph.Graph{
		AppName:     "x",
		Environment: "dev",
		Nodes:       []graph.Node{{Name: "n", Provider: "custom", Deployable: true}},
	}
	sw := &EnvSwitcher{
		Names: []string{"dev", "staging"},
		Index: 0,
		Build: func(e string) (*graph.Graph, error) {
			return &graph.Graph{
				AppName:     "x",
				Environment: e,
				Nodes:       []graph.Node{{Name: "n", Provider: "custom", Deployable: true}},
			}, nil
		},
	}
	m := NewStackModelWithEnvs(g0, nil, true, sw)
	m.width, m.height = 80, 24
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'E'}})
	mm := next.(*StackModel)
	if mm.g.Environment != "staging" {
		t.Fatalf("want staging after e, got %q", mm.g.Environment)
	}
}
