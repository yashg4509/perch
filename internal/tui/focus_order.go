package tui

import (
	"sort"

	"github.com/yashg4509/perch/internal/graph"
)

// NodeFocusOrder returns node names in left-to-right layer order, top-to-bottom within each column.
// Matches horizontal graph layout. Empty if the graph cannot be layered (cycle) or has no nodes.
func NodeFocusOrder(g *graph.Graph) []string {
	if g == nil || len(g.Nodes) == 0 {
		return nil
	}
	nodeByName := make(map[string]graph.Node, len(g.Nodes))
	for _, n := range g.Nodes {
		nodeByName[n.Name] = n
	}
	layer, ok := computeLayers(nodeByName, g.Edges)
	if !ok {
		return nil
	}
	cols := nodesByLayer(nodeByName, layer)
	var out []string
	for _, col := range cols {
		for _, name := range col {
			out = append(out, name)
		}
	}
	return out
}

// SortedNodeNames is lexicographic order (list fallback).
func SortedNodeNames(g *graph.Graph) []string {
	if g == nil {
		return nil
	}
	names := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		names = append(names, n.Name)
	}
	sort.Strings(names)
	return names
}
