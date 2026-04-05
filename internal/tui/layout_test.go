package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/graph"
)

func TestRenderLayout_fixed8080(t *testing.T) {
	g := &graph.Graph{
		AppName:     "fixture-app",
		Environment: "production",
		Nodes: []graph.Node{
			{Name: "frontend", Provider: "vercel", Deployable: true, Project: "demo"},
			{Name: "llm", Provider: "openai", Deployable: false},
		},
		Edges: []config.Edge{{From: "frontend", To: "llm"}},
	}
	const width, height = 80, 24
	s := RenderLayout(g, width, height, true)
	if !strings.Contains(s, "fixture-app") || !strings.Contains(s, "production") {
		t.Fatalf("missing header: %q", s)
	}
	if !strings.Contains(s, "frontend") || !strings.Contains(s, "vercel") {
		t.Fatalf("missing node: %q", s)
	}
	if !strings.Contains(s, "llm") || !strings.Contains(s, "openai") {
		t.Fatalf("missing node: %q", s)
	}
	if !strings.Contains(s, "llm") || !strings.Contains(s, "●") {
		t.Fatalf("missing circle node or llm label: %q", s)
	}
	if !strings.Contains(s, "►") && !strings.Contains(s, "─") {
		t.Fatalf("missing graph edge: %q", s)
	}
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	if len(lines) > height {
		t.Fatalf("layout height %d exceeds cap %d", len(lines), height)
	}
	for i, line := range lines {
		if w := visibleWidth(line); w > width {
			t.Fatalf("line %d width %d > %d: %q", i, w, width, line)
		}
	}
}

func TestRenderLayout_coloredContainsANSI(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	g := &graph.Graph{
		AppName:     "c",
		Environment: "p",
		Nodes: []graph.Node{
			{Name: "a", Provider: "vercel", Deployable: true},
			{Name: "b", Provider: "openai", Deployable: false},
		},
		Edges: []config.Edge{{From: "a", To: "b"}},
	}
	s := RenderLayout(g, 80, 24, false)
	if !strings.Contains(s, "\x1b[") {
		t.Fatalf("expected ANSI color sequences when noColor=false: %q", s)
	}
}

func TestRenderLayout_cycleFallsBackToList(t *testing.T) {
	g := &graph.Graph{
		AppName:     "cyclic",
		Environment: "dev",
		Nodes: []graph.Node{
			{Name: "a", Provider: "custom", Deployable: true},
			{Name: "b", Provider: "custom", Deployable: true},
		},
		Edges: []config.Edge{{From: "a", To: "b"}, {From: "b", To: "a"}},
	}
	s := RenderLayout(g, 80, 24, true)
	if !strings.Contains(s, "Nodes") || !strings.Contains(s, "Edges") {
		t.Fatalf("expected list fallback for cycle, got: %q", s)
	}
}
