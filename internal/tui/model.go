package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/stackstatus"
)

// EnvSwitcher lets the user press "E" to rotate perch.yaml environments (when more than one exists).
type EnvSwitcher struct {
	Names []string
	Index int
	Build func(env string) (*graph.Graph, error)
}

// Next advances to the next environment name and rebuilds the graph. Nil no-op if disabled.
func (sw *EnvSwitcher) Next() (*graph.Graph, error) {
	if sw == nil || len(sw.Names) <= 1 || sw.Build == nil {
		return nil, nil
	}
	sw.Index = (sw.Index + 1) % len(sw.Names)
	return sw.Build(sw.Names[sw.Index])
}

// StatusFetcher loads node health for the active environment (same work as perch status).
type StatusFetcher func(ctx context.Context, env string) (*stackstatus.EnvReport, error)

type stackPanel int

const (
	panelGraph stackPanel = iota
	panelPalette
	panelDetail
	panelLogs
	panelEnv
	panelDeploy
	panelTimeline
)

type statusLoadedMsg struct {
	rep *stackstatus.EnvReport
	err error
}

// StackModel is the Bubbletea model for the stack overview.
type StackModel struct {
	g       *graph.Graph
	loadErr error
	width   int
	height  int
	noColor bool
	env     *EnvSwitcher

	fetchStatus StatusFetcher
	statusRep   *stackstatus.EnvReport
	statusErr   error
	statusLoad  bool

	focusNames []string
	selIdx     int

	panel stackPanel
}

// NewStackModel returns a model without environment switching.
func NewStackModel(g *graph.Graph, loadErr error, noColor bool) *StackModel {
	return NewStackModelWithEnvs(g, loadErr, noColor, nil)
}

// NewStackModelWithEnvs returns a model; pass a non-nil env switcher to enable "E" when multiple environments exist.
func NewStackModelWithEnvs(g *graph.Graph, loadErr error, noColor bool, env *EnvSwitcher) *StackModel {
	return NewStackModelWithEnvsAndFetch(g, loadErr, noColor, env, nil)
}

// NewStackModelWithEnvsAndFetch enables live health colors (●) via the same collection path as perch status.
func NewStackModelWithEnvsAndFetch(g *graph.Graph, loadErr error, noColor bool, env *EnvSwitcher, fetch StatusFetcher) *StackModel {
	m := &StackModel{
		g:           g,
		loadErr:     loadErr,
		width:       80,
		height:      24,
		noColor:     noColor,
		env:         env,
		fetchStatus: fetch,
		panel:       panelGraph,
	}
	m.syncFocusOrder()
	return m
}

func (m *StackModel) syncFocusOrder() {
	if m.g == nil {
		m.focusNames = nil
		m.selIdx = 0
		return
	}
	order := NodeFocusOrder(m.g)
	if len(order) == 0 {
		order = SortedNodeNames(m.g)
	}
	m.focusNames = order
	if len(m.focusNames) == 0 {
		m.selIdx = 0
		return
	}
	if m.selIdx >= len(m.focusNames) {
		m.selIdx = len(m.focusNames) - 1
	}
	if m.selIdx < 0 {
		m.selIdx = 0
	}
}

func (m *StackModel) selectedName() string {
	if len(m.focusNames) == 0 || m.selIdx < 0 || m.selIdx >= len(m.focusNames) {
		return ""
	}
	return m.focusNames[m.selIdx]
}

func (m *StackModel) statusMap() map[string]stackstatus.NodeReport {
	if m.statusRep == nil {
		return nil
	}
	out := make(map[string]stackstatus.NodeReport, len(m.statusRep.Nodes))
	for _, row := range m.statusRep.Nodes {
		out[row.Name] = row
	}
	return out
}

func (m *StackModel) selectedReportPtr() *stackstatus.NodeReport {
	name := m.selectedName()
	if name == "" || m.statusRep == nil {
		return nil
	}
	for i := range m.statusRep.Nodes {
		if m.statusRep.Nodes[i].Name == name {
			return &m.statusRep.Nodes[i]
		}
	}
	return nil
}

func (m *StackModel) currentEnv() string {
	if m.g != nil {
		return m.g.Environment
	}
	return ""
}

func (m *StackModel) cmdRefreshStatus() tea.Cmd {
	if m.fetchStatus == nil {
		return nil
	}
	env := m.currentEnv()
	fetch := m.fetchStatus
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		rep, err := fetch(ctx, env)
		return statusLoadedMsg{rep: rep, err: err}
	}
}

func (m *StackModel) navigate(ks string) {
	if len(m.focusNames) == 0 {
		return
	}
	delta := 0
	switch ks {
	case "up", "left":
		delta = -1
	case "down", "right":
		delta = 1
	default:
		return
	}
	m.selIdx = (m.selIdx + delta + len(m.focusNames)) % len(m.focusNames)
}

// Init implements tea.Model.
func (m *StackModel) Init() tea.Cmd {
	if m.loadErr != nil || m.g == nil {
		return nil
	}
	m.syncFocusOrder()
	return m.cmdRefreshStatus()
}

// Update implements tea.Model.
func (m *StackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.width < 40 {
			m.width = 80
		}
		if m.height < 8 {
			m.height = 24
		}
		return m, nil
	case statusLoadedMsg:
		m.statusLoad = false
		if msg.err != nil {
			m.statusErr = msg.err
		} else {
			m.statusErr = nil
			m.statusRep = msg.rep
		}
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *StackModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ks := msg.String()

	if m.panel != panelGraph {
		switch ks {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc", "?":
			m.panel = panelGraph
			return m, nil
		default:
			return m, nil
		}
	}

	switch ks {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "?":
		m.panel = panelPalette
		return m, nil
	case "enter":
		if m.selectedName() != "" {
			m.panel = panelDetail
		}
		return m, nil
	case "up", "down", "left", "right":
		m.navigate(ks)
		return m, nil
	case "r":
		if m.fetchStatus != nil {
			m.statusLoad = true
			return m, m.cmdRefreshStatus()
		}
		return m, nil
	case "l":
		m.panel = panelLogs
		return m, nil
	case "e":
		m.panel = panelEnv
		return m, nil
	case "d":
		m.panel = panelDeploy
		return m, nil
	case "t":
		m.panel = panelTimeline
		return m, nil
	case "E", "shift+e":
		if m.env != nil {
			if ng, err := m.env.Next(); err == nil && ng != nil {
				m.g = ng
				m.syncFocusOrder()
				var cmd tea.Cmd
				if m.fetchStatus != nil {
					m.statusLoad = true
					cmd = m.cmdRefreshStatus()
				}
				return m, cmd
			}
		}
		return m, nil
	}
	return m, nil
}

// View implements tea.Model.
func (m *StackModel) View() string {
	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	if m.loadErr != nil {
		msg := fmt.Sprintf("perch: %v\n\nPress q to quit.", m.loadErr)
		return clipBlock(msg, w, h)
	}
	if m.g == nil {
		return clipBlock("Nothing to show.\n\nPress q to quit.", w, h)
	}

	switch m.panel {
	case panelPalette:
		return renderCommandPalette(w, h, m.noColor)
	case panelDetail:
		return renderNodeDetail(m.g, m.selectedName(), m.selectedReportPtr(), w, h, m.noColor)
	case panelLogs:
		body := logsStubBody(m.g, m.selectedName())
		return renderStubPanel("Logs", body, w, h, m.noColor)
	case panelEnv:
		body := envStubBody(m.selectedName())
		return renderStubPanel("Environment variables", body, w, h, m.noColor)
	case panelDeploy:
		return renderStubPanel("Deploy / rollback", "Roadmap: trigger deploy and rollback from the TUI.\n\nToday: use your platform CLI or perch deploy when implemented.\n\nEsc back", w, h, m.noColor)
	case panelTimeline:
		return renderStubPanel("Timeline", "Roadmap: chronological deploys, restarts, and errors across the stack.\n\nEsc back", w, h, m.noColor)
	}

	statusNote := ""
	if m.statusLoad {
		statusNote = "Refreshing health…"
	} else if m.statusErr != nil {
		statusNote = "Status error: " + m.statusErr.Error()
	}
	chromeH := StackChromeLines(statusNote)
	bodyH := h - chromeH
	if bodyH < 4 {
		bodyH = 4
	}
	body := RenderLayoutState(m.g, w, bodyH, m.noColor, m.selectedName(), m.statusMap())
	chrome := StackChrome(m.g, w, m.noColor, m.env, statusNote)
	return clipBlock(body+"\n"+chrome, w, h)
}

func logsStubBody(g *graph.Graph, nodeName string) string {
	if g == nil || nodeName == "" {
		return "Focus a node with arrow keys, then press l.\n\nEsc back"
	}
	var n *graph.Node
	for i := range g.Nodes {
		if g.Nodes[i].Name == nodeName {
			n = &g.Nodes[i]
			break
		}
	}
	if n == nil {
		return "Unknown node.\n\nEsc back"
	}
	if strings.TrimSpace(n.Logs) != "" {
		return fmt.Sprintf("Node %q — configured logs command:\n\n  %s\n\nStream/follow is not run inside the TUI yet (avoids hanging on tail -f). Run it in your shell, or use perch logs when that command ships.\n\nEsc back", n.Name, n.Logs)
	}
	return fmt.Sprintf("Node %q has no logs command in perch.yaml.\n\nEsc back", n.Name)
}

func envStubBody(nodeName string) string {
	if nodeName == "" {
		return "Focus a node with arrow keys.\n\nRoadmap: perch env list / diff from here.\n\nEsc back"
	}
	return fmt.Sprintf("Node %q — env vars\n\nRoadmap: perch env list --node %s (masked values, reveal flag).\n\nEsc back", nodeName, nodeName)
}
