package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yashg4509/perch/internal/graph"
)

func TestStackModel_quitOnQ(t *testing.T) {
	g := &graph.Graph{
		AppName:     "x",
		Environment: "production",
		Nodes:       []graph.Node{{Name: "a", Provider: "vercel", Deployable: true}},
		Edges:       nil,
	}
	m := NewStackModel(g, nil, true)
	m.width, m.height = 80, 24
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if next == nil {
		t.Fatal("expected model back")
	}
}

func TestStackModel_windowSizeUpdatesDimensions(t *testing.T) {
	m := NewStackModel(&graph.Graph{AppName: "z", Environment: "dev", Nodes: []graph.Node{{Name: "n", Provider: "openai"}}}, nil, true)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	mm, ok := next.(*StackModel)
	if !ok {
		t.Fatalf("type %T", next)
	}
	if mm.width != 100 || mm.height != 30 {
		t.Fatalf("got %d x %d", mm.width, mm.height)
	}
}
