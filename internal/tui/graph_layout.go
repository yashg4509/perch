package tui

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/yashg4509/perch/internal/config"
	"github.com/yashg4509/perch/internal/graph"
	"github.com/yashg4509/perch/internal/stackstatus"
)

// nodeRect tracks a node's cell on the canvas for ANSI coloring.
type nodeRect struct {
	name            string
	x0, cellW       int
	bulletY, labelY int
}

// renderHorizontalGraph draws nodes as ● + label, left → right by layer, with optional color.
// selected highlights focus (blue ● per spec). status maps node name → last collected row (nil map = unknown health).
func renderHorizontalGraph(g *graph.Graph, width, height int, noColor bool, selected string, status map[string]stackstatus.NodeReport) string {
	if g == nil || len(g.Nodes) == 0 {
		return ""
	}
	nodeByName := make(map[string]graph.Node, len(g.Nodes))
	for _, n := range g.Nodes {
		nodeByName[n.Name] = n
	}

	layer, ok := computeLayers(nodeByName, g.Edges)
	if !ok {
		return ""
	}
	cols := nodesByLayer(nodeByName, layer)
	maxPerCol := 0
	for _, col := range cols {
		if len(col) > maxPerCol {
			maxPerCol = len(col)
		}
	}
	if maxPerCol == 0 {
		return ""
	}

	cellW := minCellW(g.Nodes)
	const nodeH = 2 // bullet row + label row
	const vGap = 1
	const arrowW = 5

	rowStride := nodeH + vGap
	canvasH := maxPerCol*rowStride - vGap
	if canvasH < nodeH {
		canvasH = nodeH
	}
	numCols := len(cols)
	canvasW := numCols*cellW + max(0, numCols-1)*arrowW

	if canvasW > width && numCols > 1 {
		avail := width - max(0, numCols-1)*arrowW
		if avail < numCols*6 {
			return ""
		}
		cellW = avail / numCols
		if cellW < 6 {
			return ""
		}
		canvasW = numCols*cellW + max(0, numCols-1)*arrowW
	}

	canvas := newCanvas(canvasW, canvasH)
	var rects []nodeRect

	pos := make(map[string]struct{ x, y int })
	for ci, names := range cols {
		k := len(names)
		startRow := (maxPerCol - k) / 2
		x := ci * (cellW + arrowW)
		for ri, name := range names {
			row := startRow + ri
			y := row * rowStride
			drawNodeCircle(canvas, x, y, cellW, nodeByName[name], &rects)
			pos[name] = struct{ x, y int }{x, y}
		}
	}

	for _, e := range g.Edges {
		p0, ok0 := pos[e.From]
		p1, ok1 := pos[e.To]
		if !ok0 || !ok1 {
			continue
		}
		y0 := p0.y
		y1 := p1.y
		x0 := p0.x + cellW
		x1 := p1.x
		drawHArrow(canvas, x0, y0, x1, y1)
	}

	s := canvas.renderColored(rects, nodeByName, noColor, selected, status)
	return clipBlock(strings.TrimSuffix(s, "\n"), width, height)
}

func edgeLipglossStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"})
}

func isEdgeRune(ch rune) bool {
	switch ch {
	case '─', '►', '│', '┐', '└', '┌', '┘':
		return true
	default:
		return false
	}
}

func paintLine(r []rune, y int, rects []nodeRect, m map[string]graph.Node, noColor bool, selected string, status map[string]stackstatus.NodeReport) string {
	if noColor {
		return string(r)
	}
	bulletAt := make(map[int]string)
	for _, rect := range rects {
		if rect.bulletY != y {
			continue
		}
		bx := rect.x0 + (rect.cellW-1)/2
		if bx >= 0 && bx < len(r) {
			bulletAt[bx] = rect.name
		}
	}
	var b strings.Builder
	for i, ch := range r {
		if name, ok := bulletAt[i]; ok && ch == '●' {
			n := m[name]
			var rep *stackstatus.NodeReport
			if status != nil {
				if row, ok := status[name]; ok {
					rep = &row
				}
			}
			sel := name == selected
			b.WriteString(bulletStyle(n, sel, rep, noColor).Render("●"))
			continue
		}
		if isEdgeRune(ch) {
			b.WriteString(edgeLipglossStyle().Render(string(ch)))
			continue
		}
		b.WriteRune(ch)
	}
	return b.String()
}

func paintLabelLine(r []rune, y int, rects []nodeRect, m map[string]graph.Node, noColor bool, selected string, status map[string]stackstatus.NodeReport) string {
	if noColor {
		return string(r)
	}
	owner := make([]string, len(r))
	for _, rect := range rects {
		if rect.labelY != y {
			continue
		}
		for j := rect.x0; j < rect.x0+rect.cellW && j < len(r); j++ {
			owner[j] = rect.name
		}
	}
	var b strings.Builder
	i := 0
	for i < len(r) {
		name := owner[i]
		if name != "" {
			start := i
			for i < len(r) && owner[i] == name {
				i++
			}
			cell := string(r[start:i])
			inner := strings.TrimSpace(cell)
			var rep *stackstatus.NodeReport
			if status != nil {
				if row, ok := status[name]; ok {
					rep = &row
				}
			}
			tier := tierFromReport(rep)
			b.WriteString(styleLabelCell(inner, name == selected, tier, noColor, utf8.RuneCountInString(cell)))
			continue
		}
		ch := r[i]
		if isEdgeRune(ch) {
			b.WriteString(edgeLipglossStyle().Render(string(ch)))
		} else {
			b.WriteRune(ch)
		}
		i++
	}
	return b.String()
}

func isLabelRow(y int, rects []nodeRect) bool {
	for _, r := range rects {
		if r.labelY == y {
			return true
		}
	}
	return false
}

func (c *canvas) renderColored(rects []nodeRect, m map[string]graph.Node, noColor bool, selected string, status map[string]stackstatus.NodeReport) string {
	lines := make([]string, c.h)
	for y := 0; y < c.h; y++ {
		row := c.b[y]
		if isLabelRow(y, rects) {
			lines[y] = paintLabelLine(row, y, rects, m, noColor, selected, status)
		} else {
			lines[y] = paintLine(row, y, rects, m, noColor, selected, status)
		}
	}
	return strings.Join(lines, "\n")
}

// comboLabel prefers "name·provider" when it fits; otherwise truncates the name.
func comboLabel(n graph.Node, maxInner int) string {
	if maxInner < 1 {
		return "…"
	}
	combo := n.Name + "·" + n.Provider
	rc := []rune(combo)
	if len(rc) <= maxInner {
		return combo
	}
	rn := []rune(n.Name)
	if len(rn) <= maxInner {
		return n.Name
	}
	if maxInner <= 1 {
		return "…"
	}
	return string(rn[:maxInner-1]) + "…"
}

func minCellW(nodes []graph.Node) int {
	w := 8
	for _, n := range nodes {
		combo := n.Name + "·" + n.Provider
		l := len([]rune(combo)) + 2
		if l > w {
			w = l
		}
	}
	if w > 28 {
		w = 28
	}
	return w
}

// computeLayers assigns integer layers (0 = leftmost). Returns false if a cycle is detected.
func computeLayers(nodeByName map[string]graph.Node, edges []config.Edge) (map[string]int, bool) {
	layer := make(map[string]int, len(nodeByName))
	for name := range nodeByName {
		layer[name] = 0
	}
	n := len(nodeByName)
	for k := 0; k < n+2; k++ {
		changed := false
		for _, e := range edges {
			if _, ok := nodeByName[e.From]; !ok {
				continue
			}
			if _, ok := nodeByName[e.To]; !ok {
				continue
			}
			want := layer[e.From] + 1
			if layer[e.To] < want {
				layer[e.To] = want
				changed = true
			}
		}
		if !changed {
			return layer, true
		}
	}
	return layer, false
}

// nodesByLayer returns columns left-to-right; each column is sorted node names.
func nodesByLayer(nodeByName map[string]graph.Node, layer map[string]int) [][]string {
	maxZ := 0
	for _, z := range layer {
		if z > maxZ {
			maxZ = z
		}
	}
	cols := make([][]string, maxZ+1)
	for name := range nodeByName {
		z := layer[name]
		cols[z] = append(cols[z], name)
	}
	for _, col := range cols {
		sort.Strings(col)
	}
	return cols
}

type canvas struct {
	w, h int
	b    [][]rune
}

func padCenter(s string, w int) string {
	r := []rune(s)
	if len(r) >= w {
		if w <= 0 {
			return ""
		}
		if len(r) > w {
			return string(r[:w])
		}
	}
	pad := w - len(r)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func newCanvas(w, h int) *canvas {
	b := make([][]rune, h)
	for i := range b {
		b[i] = make([]rune, w)
		for j := range b[i] {
			b[i][j] = ' '
		}
	}
	return &canvas{w: w, h: h, b: b}
}

func (c *canvas) set(x, y int, ch rune) {
	if y < 0 || y >= c.h || x < 0 || x >= c.w {
		return
	}
	if ch == 0 {
		return
	}
	if c.b[y][x] != ' ' && ch == ' ' {
		return
	}
	c.b[y][x] = ch
}

func drawNodeCircle(c *canvas, x, y, cellW int, n graph.Node, rects *[]nodeRect) {
	label := comboLabel(n, cellW)
	dotLine := padCenter("●", cellW)
	labelLine := padCenter(label, cellW)
	for dx, ch := range []rune(dotLine) {
		c.set(x+dx, y, ch)
	}
	for dx, ch := range []rune(labelLine) {
		c.set(x+dx, y+1, ch)
	}
	*rects = append(*rects, nodeRect{
		name:    n.Name,
		x0:      x,
		cellW:   cellW,
		bulletY: y,
		labelY:  y + 1,
	})
}

func drawHArrow(c *canvas, x0, y0, x1, y1 int) {
	if x1 <= x0 {
		return
	}
	midX := x0 + (x1-x0)/2
	if midX <= x0 {
		midX = x0 + 1
	}
	if y0 == y1 {
		for x := x0; x < x1-1; x++ {
			c.set(x, y0, '─')
		}
		c.set(x1-1, y0, '►')
		return
	}
	if y1 > y0 {
		for x := x0; x < midX; x++ {
			c.set(x, y0, '─')
		}
		c.set(midX, y0, '┐')
		for y := y0 + 1; y < y1; y++ {
			c.set(midX, y, '│')
		}
		c.set(midX, y1, '└')
		for x := midX + 1; x < x1-1; x++ {
			c.set(x, y1, '─')
		}
		c.set(x1-1, y1, '►')
		return
	}
	for x := x0; x < midX; x++ {
		c.set(x, y0, '─')
	}
	c.set(midX, y0, '┘')
	for y := y0 - 1; y > y1; y-- {
		c.set(midX, y, '│')
	}
	c.set(midX, y1, '┌')
	for x := midX + 1; x < x1-1; x++ {
		c.set(x, y1, '─')
	}
	c.set(x1-1, y1, '►')
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// renderListFallback is the original node/edge list (used on cycle or empty graph draw).
func renderListFallback(g *graph.Graph, width, height int, noColor bool, selected string, status map[string]stackstatus.NodeReport) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	if g == nil {
		return clipBlock("(no graph)", width, height)
	}

	var b strings.Builder
	title := fmt.Sprintf("%s  [%s]", g.AppName, g.Environment)
	b.WriteString(padLinePlain(title, width))
	b.WriteByte('\n')
	b.WriteString(padLinePlain(strings.Repeat("─", min(60, width-2)), width))
	b.WriteByte('\n')
	b.WriteString(padLinePlain("Nodes", width))
	b.WriteByte('\n')
	names := SortedNodeNames(g)
	for _, name := range names {
		n := graphNodeByName(g, name)
		kind := "read-only"
		if n.Deployable {
			kind = "deployable"
		}
		mark := " "
		if n.Name == selected {
			mark = "›"
		}
		health := "unknown"
		var rep *stackstatus.NodeReport
		if status != nil {
			if row, ok := status[n.Name]; ok {
				rep = &row
				if row.Healthy {
					health = "healthy"
				} else {
					health = "error"
				}
			}
		}
		line := fmt.Sprintf("%s • %s  %s  (%s)  %s", mark, n.Name, n.Provider, kind, health)
		if n.Project != "" {
			line += fmt.Sprintf("  project=%s", n.Project)
		}
		ln := line
		if !noColor {
			tier := tierFromReport(rep)
			st := lipgloss.NewStyle()
			if n.Name == selected {
				st = st.Foreground(selectedBulletColor()).Bold(true)
			} else {
				st = st.Foreground(healthBulletColor(tier))
			}
			ln = st.Render(line)
		}
		b.WriteString(padLineVisual(ln, width))
		b.WriteByte('\n')
	}
	if len(g.Edges) > 0 {
		b.WriteString(padLinePlain("Edges", width))
		b.WriteByte('\n')
		for _, e := range g.Edges {
			line := fmt.Sprintf("  %s -> %s", e.From, e.To)
			el := line
			if !noColor {
				el = edgeLipglossStyle().Render(line)
			}
			b.WriteString(padLineVisual(el, width))
			b.WriteByte('\n')
		}
	}

	block := strings.TrimSuffix(b.String(), "\n")
	return clipBlock(block, width, height)
}

func graphNodeByName(g *graph.Graph, name string) graph.Node {
	for _, n := range g.Nodes {
		if n.Name == name {
			return n
		}
	}
	return graph.Node{Name: name, Provider: "?"}
}
