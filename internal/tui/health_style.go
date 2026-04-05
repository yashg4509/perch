package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/stackstatus"
)

// healthTier maps status reports to spec colors (green / yellow / red / grey).
type healthTier int

const (
	tierUnknown healthTier = iota
	tierHealthy
	tierDegraded
	tierError
)

func tierFromReport(rep *stackstatus.NodeReport) healthTier {
	if rep == nil {
		return tierUnknown
	}
	if !rep.Healthy {
		return tierError
	}
	if rep.ErrorRate != nil && *rep.ErrorRate >= 0.01 {
		return tierDegraded
	}
	return tierHealthy
}

func healthBulletColor(tier healthTier) lipgloss.TerminalColor {
	switch tier {
	case tierHealthy:
		return lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#4ADE80"}
	case tierDegraded:
		return lipgloss.AdaptiveColor{Light: "#CA8A04", Dark: "#FACC15"}
	case tierError:
		return lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}
	default:
		return lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}
	}
}

func selectedBulletColor() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#60A5FA"}
}

func fallbackDeployableColor(n graph.Node) lipgloss.TerminalColor {
	if n.Deployable {
		return lipgloss.AdaptiveColor{Light: "#1D4ED8", Dark: "#93C5FD"}
	}
	return lipgloss.AdaptiveColor{Light: "#6D28D9", Dark: "#DDD6FE"}
}

func bulletStyle(n graph.Node, selected bool, rep *stackstatus.NodeReport, noColor bool) lipgloss.Style {
	st := lipgloss.NewStyle().Bold(true)
	if noColor {
		return st
	}
	if selected {
		return st.Foreground(selectedBulletColor())
	}
	if rep != nil {
		return st.Foreground(healthBulletColor(tierFromReport(rep)))
	}
	return st.Foreground(fallbackDeployableColor(n))
}

func nameLabelStyle(selected bool, tier healthTier, noColor bool) lipgloss.Style {
	st := lipgloss.NewStyle().Bold(true)
	if noColor {
		return st
	}
	if selected {
		return st.Foreground(selectedBulletColor())
	}
	switch tier {
	case tierHealthy:
		st = st.Foreground(lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F9FAFB"})
	case tierDegraded:
		st = st.Foreground(lipgloss.AdaptiveColor{Light: "#854D0E", Dark: "#FEF08A"})
	case tierError:
		st = st.Foreground(lipgloss.AdaptiveColor{Light: "#991B1B", Dark: "#FCA5A5"})
	default:
		st = st.Foreground(lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F9FAFB"})
	}
	return st
}

func providerLabelStyle(noColor bool) lipgloss.Style {
	st := lipgloss.NewStyle().Faint(true)
	if noColor {
		return st
	}
	return st.Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"})
}

// styleLabelCell centers styled "name·provider" within a fixed monospace cell width.
func styleLabelCell(inner string, sel bool, tier healthTier, noColor bool, cellRuneLen int) string {
	if noColor || inner == "" {
		return padCenterPlain(inner, cellRuneLen)
	}
	dot := strings.Index(inner, "·")
	namePart, provPart := inner, ""
	if dot >= 0 {
		namePart = inner[:dot]
		provPart = inner[dot:]
	}
	combined := nameLabelStyle(sel, tier, false).Render(namePart) + providerLabelStyle(false).Render(provPart)
	sw := ansi.StringWidth(combined)
	pad := cellRuneLen - sw
	if pad < 0 {
		return ansi.Truncate(combined, cellRuneLen, "…")
	}
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + combined + strings.Repeat(" ", right)
}

func padCenterPlain(s string, w int) string {
	r := []rune(s)
	if w <= 0 {
		return ""
	}
	if len(r) >= w {
		if len(r) > w {
			return string(r[:w])
		}
		return s
	}
	pad := w - len(r)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
