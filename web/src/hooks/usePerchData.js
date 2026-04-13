import { useCallback, useEffect, useRef, useState } from 'react'
import { mockEdges, mockNodes } from '../data/mock.js'
import { getLayoutedElements } from '../graph/layout.js'
import { mapGraphToEdges, mapGraphToNodes } from '../lib/mappers.js'

function buildFlowFromMock() {
  const rawNodes = mockNodes.map((n) => ({
    id: n.id,
    type: 'serviceCard',
    data: { ...n },
    position: { x: 0, y: 0 },
  }))
  const rawEdges = mockEdges.map((e) => ({
    id: e.id,
    source: e.source,
    target: e.target,
    animated: e.animated,
  }))
  return getLayoutedElements(rawNodes, rawEdges)
}

const initial = buildFlowFromMock()

async function fetchJson(url) {
  const res = await fetch(url, { headers: { Accept: 'application/json' } })
  const text = await res.text()
  let body
  try {
    body = text ? JSON.parse(text) : null
  } catch {
    body = null
  }
  if (!res.ok) {
    const msg = body?.error ?? res.statusText ?? 'request failed'
    throw new Error(typeof msg === 'string' ? msg : JSON.stringify(msg))
  }
  return body
}

/**
 * @param {string} env
 */
export function usePerchData(env) {
  const [nodes, setNodes] = useState(initial.nodes)
  const [edges, setEdges] = useState(initial.edges)
  const [appName, setAppName] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const mounted = useRef(true)

  const load = useCallback(async () => {
    const graphUrl = `/api/graph?env=${encodeURIComponent(env)}`
    const statusUrl = `/api/status?env=${encodeURIComponent(env)}`
    try {
      const [graphJson, statusJson] = await Promise.all([fetchJson(graphUrl), fetchJson(statusUrl)])
      const name = graphJson?.appName
      if (typeof name === 'string' && name.trim() !== '') {
        setAppName(name.trim())
      }
      const mappedNodes = mapGraphToNodes(graphJson, statusJson)
      const mappedEdges = mapGraphToEdges(graphJson)
      const rawNodes = mappedNodes.map((n) => ({
        id: n.id,
        type: 'serviceCard',
        data: { ...n },
        position: { x: 0, y: 0 },
      }))
      const rawEdges = mappedEdges.map((e) => ({
        id: e.id,
        source: e.source,
        target: e.target,
        animated: e.animated,
      }))
      const { nodes: nextNodes, edges: nextEdges } = getLayoutedElements(rawNodes, rawEdges)
      if (!mounted.current) {
        return
      }
      setNodes(nextNodes)
      setEdges(nextEdges)
      setError(null)
    } catch (e) {
      if (!mounted.current) {
        return
      }
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      if (mounted.current) {
        setLoading(false)
      }
    }
  }, [env])

  useEffect(() => {
    mounted.current = true
    setLoading(true)
    setError(null)
    void load()
    const id = window.setInterval(() => {
      void load()
    }, 10_000)
    return () => {
      mounted.current = false
      window.clearInterval(id)
    }
  }, [load])

  return { nodes, edges, appName, loading, error, refetch: load }
}
