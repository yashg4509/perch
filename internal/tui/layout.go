package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/stackstatus"
)

// RenderLayout renders a fixed-size stack view for tests and [StackModel.View].
// Primary mode: horizontal graph with ● nodes and labels; color when noColor is false.
// Falls back to a node/edge list if the graph has a cycle or cannot fit.
func RenderLayout(g *graph.Graph, width, height int, noColor bool) string {
	return RenderLayoutState(g, width, height, noColor, "", nil)
}

// RenderLayoutState is like [RenderLayout] but highlights selected and colors ● from status rows.
func RenderLayoutState(g *graph.Graph, width, height int, noColor bool, selected string, status map[string]stackstatus.NodeReport) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	if g == nil {
		return clipBlock("(no graph)", width, height)
	}
	if !noColor && !isatty.IsTerminal(os.Stdout.Fd()) {
		// Lipgloss omits ANSI when stdout is not a TTY (tests, pipes). Interactive TTY keeps auto-detected profile.
		lipgloss.SetColorProfile(termenv.TrueColor)
	}

	title := fmt.Sprintf("%s  [%s]", g.AppName, g.Environment)
	sep := strings.Repeat("─", min(60, width-2))
	head := padLinePlain(title, width) + "\n" + padLinePlain(sep, width)
	headLines := 2
	graphH := height - headLines
	if graphH < 4 {
		graphH = height
		head = ""
		headLines = 0
	}

	viz := renderHorizontalGraph(g, width, graphH, noColor, selected, status)
	if viz == "" {
		return renderListFallback(g, width, height, noColor, selected, status)
	}
	if headLines > 0 {
		return clipBlock(head+"\n"+viz, width, height)
	}
	return clipBlock(viz, width, height)
}

func clipBlock(s string, width, height int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		t := ansi.Truncate(ln, width, "…")
		pad := width - ansi.StringWidth(t)
		if pad > 0 {
			t += strings.Repeat(" ", pad)
		}
		out = append(out, t)
	}
	return strings.Join(out, "\n")
}

// padLinePlain pads plain text (no ANSI) to width runes.
func padLinePlain(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return string(r[:width])
	}
	return s + strings.Repeat(" ", width-len(r))
}

// padLineVisual pads a string that may contain ANSI to visible width.
func padLineVisual(s string, width int) string {
	sw := ansi.StringWidth(s)
	if sw >= width {
		return ansi.Truncate(s, width, "…")
	}
	return s + strings.Repeat(" ", width-sw)
}

// padLine is an alias for plain footers and simple lines.
func padLine(s string, width int) string {
	return padLinePlain(s, width)
}

func truncateRunes(s string, width int) string {
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return "…"
	}
	return string(r[:width-1]) + "…"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func visibleWidth(s string) int {
	return ansi.StringWidth(s)
}

// VisibleLineWidth returns the rendered width of s accounting for ANSI styling.
func VisibleLineWidth(s string) int {
	return visibleWidth(s)
}
