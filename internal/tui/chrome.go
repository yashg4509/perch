package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/yashg4509/perch/internal/graph"
)

// StackChrome returns footer lines below the graph: optional status, summary, flows, CLI tip, key bar.
func StackChrome(g *graph.Graph, width int, noColor bool, env *EnvSwitcher, statusNote string) string {
	if g == nil {
		return ""
	}
	var lines []string
	if strings.TrimSpace(statusNote) != "" {
		lines = append(lines, styleChromeRow(statusNote, noColor, true))
	}
	lines = append(lines, styleChromeRow(graphSummaryLine(g), noColor, false))
	lines = append(lines, styleChromeRow(graphFlowsLine(g), noColor, true))
	lines = append(lines, styleChromeRow(cliTipLine(), noColor, true))
	lines = append(lines, styleChromeRow(keyBarLine(env), noColor, false))

	out := strings.Join(lines, "\n")
	if width <= 0 {
		return out
	}
	var clipped []string
	for _, ln := range strings.Split(out, "\n") {
		clipped = append(clipped, ansi.Truncate(ln, width, "…"))
	}
	return strings.Join(clipped, "\n")
}

func graphSummaryLine(g *graph.Graph) string {
	dep, ro := 0, 0
	prov := make(map[string]int)
	for _, n := range g.Nodes {
		if n.Deployable {
			dep++
		} else {
			ro++
		}
		prov[n.Provider]++
	}
	var pb strings.Builder
	first := true
	keys := make([]string, 0, len(prov))
	for k := range prov {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, p := range keys {
		c := prov[p]
		if !first {
			pb.WriteString(" · ")
		}
		first = false
		if c == 1 {
			pb.WriteString(p)
		} else {
			pb.WriteString(fmt.Sprintf("%s×%d", p, c))
		}
	}
	s := fmt.Sprintf("%d nodes (%d deployable · %d read-only) · %d edges · env %s",
		len(g.Nodes), dep, ro, len(g.Edges), g.Environment)
	if pb.Len() > 0 {
		s += " · providers: " + pb.String()
	}
	return s
}

func graphFlowsLine(g *graph.Graph) string {
	if len(g.Edges) == 0 {
		return "Flows: (none — add edges: in perch.yaml under edges:)"
	}
	parts := make([]string, 0, len(g.Edges))
	for _, e := range g.Edges {
		parts = append(parts, e.From+"→"+e.To)
	}
	return "Flows: " + strings.Join(parts, "  ·  ")
}

func cliTipLine() string {
	return "Tip: perch status --json · perch graph --json · perch context --for-agent"
}

func keyBarLine(env *EnvSwitcher) string {
	s := "arrows focus  ·  Enter detail  ·  ? palette  ·  r refresh  ·  l e d t actions  ·  q quit"
	if env != nil && len(env.Names) > 1 {
		s += "  ·  E env"
	}
	return s
}

func styleChromeRow(s string, noColor, faint bool) string {
	if noColor || s == "" {
		return s
	}
	st := lipgloss.NewStyle()
	if faint {
		st = st.Faint(true).Foreground(lipgloss.AdaptiveColor{Light: "#4B5563", Dark: "#9CA3AF"})
	} else {
		st = st.Foreground(lipgloss.AdaptiveColor{Light: "#374151", Dark: "#D1D5DB"})
	}
	return st.Render(s)
}

// StackChromeLines returns footer height for [StackChrome] with the same statusNote (must match View).
func StackChromeLines(statusNote string) int {
	n := 4
	if strings.TrimSpace(statusNote) != "" {
		n++
	}
	return n
}
