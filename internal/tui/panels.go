package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/stackstatus"
)

func renderCommandPalette(width, height int, noColor bool) string {
	body := `perch — command palette   Esc close

● colors (when health is loaded)
  green healthy · yellow degraded (error_rate ≥ 1%) · red error · grey unknown
  blue ● and label = focused node (overrides health tint on ●)

Navigation
  ↑ ↓ ← →     focus node
  Enter         node detail

Health
  r             refresh status (same path as perch status)

Actions
  l             logs (hint / configured command)
  e             env vars (CLI roadmap)
  d             deploy / rollback (roadmap)
  t             timeline (roadmap)

Environment
  E             next environment (when several)

Other
  ?             toggle this palette
  q             quit

Shell (edits perch.yaml)
  perch edge list | add FROM TO | remove FROM TO   (--dry-run to preview)
`
	return clipBlock(stylePanelFrame("Command palette", body, width, noColor), width, height)
}

func renderNodeDetail(g *graph.Graph, nodeName string, rep *stackstatus.NodeReport, width, height int, noColor bool) string {
	if g == nil {
		return clipBlock("(no graph)", width, height)
	}
	var n *graph.Node
	for i := range g.Nodes {
		if g.Nodes[i].Name == nodeName {
			n = &g.Nodes[i]
			break
		}
	}
	if n == nil {
		return clipBlock("Unknown node.\n\nEsc back", width, height)
	}
	var b strings.Builder
	kind := "read-only"
	if n.Deployable {
		kind = "deployable"
	}
	fmt.Fprintf(&b, "%s  (%s)\n", n.Name, n.Provider)
	fmt.Fprintf(&b, "Kind: %s\n", kind)
	if n.Project != "" {
		fmt.Fprintf(&b, "Project: %s\n", n.Project)
	}
	if n.Service != "" {
		fmt.Fprintf(&b, "Service: %s\n", n.Service)
	}
	if strings.TrimSpace(n.Status) != "" {
		fmt.Fprintf(&b, "Status cmd: %s\n", truncateRunes(n.Status, 72))
	}
	if strings.TrimSpace(n.Logs) != "" {
		fmt.Fprintf(&b, "Logs cmd: %s\n", truncateRunes(n.Logs, 72))
	}
	b.WriteByte('\n')
	if rep == nil {
		b.WriteString("Health: (no data yet — press r to refresh)\n")
	} else {
		st := "unhealthy"
		if rep.Healthy {
			st = "healthy"
		}
		fmt.Fprintf(&b, "Health: %s\n", st)
		if rep.LastDeploy != nil {
			sha := rep.LastDeploy.SHA
			ago := rep.LastDeploy.Ago
			if sha != "" || ago != "" {
				fmt.Fprintf(&b, "Last deploy: %s  %s\n", sha, ago)
			}
		}
		if rep.ErrorRate != nil {
			fmt.Fprintf(&b, "Error rate: %.4f\n", *rep.ErrorRate)
		}
		if rep.DailyCostUSD != nil {
			fmt.Fprintf(&b, "Daily cost (USD): %.2f\n", *rep.DailyCostUSD)
		}
		if len(rep.RecentErrors) > 0 {
			b.WriteString("\nRecent errors:\n")
			for i, ln := range rep.RecentErrors {
				if i >= 5 {
					break
				}
				b.WriteString("  • ")
				b.WriteString(truncateRunes(ln, 76))
				b.WriteByte('\n')
			}
		}
	}
	b.WriteString("\nKeys: l logs · e env · d deploy · t timeline · Esc back")
	return clipBlock(stylePanelFrame("Node detail", b.String(), width, noColor), width, height)
}

func renderStubPanel(title, body string, width, height int, noColor bool) string {
	return clipBlock(stylePanelFrame(title, body, width, noColor), width, height)
}

func stylePanelFrame(title, body string, width int, noColor bool) string {
	titleLine := "── " + title + " ──"
	if !noColor {
		titleLine = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#1D4ED8", Dark: "#93C5FD"}).Render(titleLine)
	}
	lines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	out := []string{ansi.Truncate(titleLine, width, "…"), ""}
	for _, ln := range lines {
		out = append(out, ansi.Truncate(ln, width, "…"))
	}
	return strings.Join(out, "\n")
}
