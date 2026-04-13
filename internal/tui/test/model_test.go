package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/tui"
)

func TestStackModel_quitOnQ(t *testing.T) {
	g := &graph.Graph{
		AppName:     "x",
		Environment: "production",
		Nodes:       []graph.Node{{Name: "a", Provider: "vercel", Deployable: true}},
		Edges:       nil,
	}
	m := tui.NewStackModel(g, nil, true)
	m.SetWindowSize(80, 24)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if next == nil {
		t.Fatal("expected model back")
	}
}

func TestStackModel_windowSizeUpdatesDimensions(t *testing.T) {
	m := tui.NewStackModel(&graph.Graph{AppName: "z", Environment: "dev", Nodes: []graph.Node{{Name: "n", Provider: "openai"}}}, nil, true)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	mm, ok := next.(*tui.StackModel)
	if !ok {
		t.Fatalf("type %T", next)
	}
	if mm.WindowWidth() != 100 || mm.WindowHeight() != 30 {
		t.Fatalf("got %d x %d", mm.WindowWidth(), mm.WindowHeight())
	}
}
