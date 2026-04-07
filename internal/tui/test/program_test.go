package tui_test

import (
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/tui"
)

func TestProgram_quitWithQ(t *testing.T) {
	m := tui.NewStackModel(&graph.Graph{
		AppName:     "x",
		Environment: "p",
		Nodes:       []graph.Node{{Name: "n", Provider: "openai", Deployable: false}},
	}, nil, true)
	p := tea.NewProgram(m, tea.WithInput(strings.NewReader("q")), tea.WithOutput(io.Discard))
	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
}
