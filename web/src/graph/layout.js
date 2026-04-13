import dagre from 'dagre'

/** Keep in sync with ServiceCard visual size (padding + content). */
export const NODE_WIDTH = 210
/** Conservative height so Dagre clears tall cards (meta lists vary). */
export const NODE_HEIGHT = 300

/**
 * @param {import('@xyflow/react').Node[]} nodes
 * @param {import('@xyflow/react').Edge[]} edges
 */
export function getLayoutedElements(nodes, edges) {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({
    rankdir: 'LR',
    ranksep: 120,
    nodesep: 88,
    marginx: 32,
    marginy: 32,
  })

  nodes.forEach((node) => {
    g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT })
  })

  edges.forEach((edge) => {
    g.setEdge(edge.source, edge.target)
  })

  dagre.layout(g)

  const layoutedNodes = nodes.map((node) => {
    const pos = g.node(node.id)
    return {
      ...node,
      position: {
        x: pos.x - NODE_WIDTH / 2,
        y: pos.y - NODE_HEIGHT / 2,
      },
    }
  })

  return { nodes: layoutedNodes, edges }
}

const OVERLAP_PAD = 6

function boxesOverlap(ax, ay, bx, by, w, h, pad) {
  return ax < bx + w + pad && bx < ax + w + pad && ay < by + h + pad && by < ay + h + pad
}

/**
 * Push node pairs apart so axis-aligned boxes (width × height) do not overlap.
 * @param {import('@xyflow/react').Node[]} nodes
 * @param {number} width
 * @param {number} height
 */
export function separateOverlappingNodes(nodes, width, height) {
  const out = nodes.map((n) => ({
    ...n,
    position: { x: n.position.x, y: n.position.y },
  }))
  for (let iter = 0; iter < 64; iter++) {
    let moved = false
    for (let i = 0; i < out.length; i++) {
      for (let j = i + 1; j < out.length; j++) {
        const a = out[i].position
        const b = out[j].position
        if (!boxesOverlap(a.x, a.y, b.x, b.y, width, height, OVERLAP_PAD)) continue
        const dx = b.x + width / 2 - (a.x + width / 2)
        const dy = b.y + height / 2 - (a.y + height / 2)
        const len = Math.hypot(dx, dy) || 1
        const step = 8
        const ox = (dx / len) * step
        const oy = (dy / len) * step
        out[j].position = { x: b.x + ox, y: b.y + oy }
        out[i].position = { x: a.x - ox, y: a.y - oy }
        moved = true
      }
    }
    if (!moved) break
  }
  return out
}
