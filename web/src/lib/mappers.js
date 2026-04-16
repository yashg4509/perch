/**
 * Maps `perch graph --json` and `perch status --json` payloads to the React UI node model.
 * Field names match Go jsonreport / stackstatus JSON tags.
 */

/**
 * @param {string} provider
 * @returns {string[]}
 */
export function tabsForProvider(provider) {
  const p = (provider ?? '').toLowerCase()
  switch (p) {
    case 'vercel':
      return ['logs', 'env', 'deploy']
    case 'supabase':
      return ['logs', 'tables', 'metrics']
    case 'github':
      return ['PRs', 'actions', 'commits']
    case 'render':
      return ['logs', 'env', 'metrics']
    case 'custom':
      return ['command', 'output']
    default:
      return ['logs', 'env']
  }
}

/**
 * @param {object} graphNode — element of graph JSON `nodes` (name, provider, deployable, project, service, status, logs)
 * @param {object | undefined} statusNode — matching row from status JSON `nodes`, or undefined
 * @returns {'healthy' | 'degraded' | 'down' | 'unknown' | 'source'}
 */
export function deriveStatus(graphNode, statusNode) {
  const prov = (graphNode?.provider ?? '').toLowerCase()
  if (prov === 'github' && graphNode && !graphNode.deployable) {
    return 'source'
  }
  if (!statusNode) {
    return 'unknown'
  }
  if (!statusNode.healthy) {
    return 'down'
  }
  const er = statusNode.error_rate
  if (typeof er === 'number' && er >= 0.01) {
    return 'degraded'
  }
  return 'healthy'
}

/**
 * @param {object} graphNode
 * @param {object | undefined} statusNode
 * @returns {{ key: string, value: string, semantic?: string }[]}
 */
export function buildMeta(graphNode, statusNode) {
  const rows = []
  const g = graphNode ?? {}
  const push = (key, value, semantic) => {
    const v = value == null ? '' : String(value)
    const row = { key, value: v }
    if (semantic) {
      row.semantic = semantic
    }
    rows.push(row)
  }

  push('name', g.name ?? '')
  push('provider', g.provider ?? '')
  push('deployable', g.deployable != null ? String(g.deployable) : '')
  push('project', g.project ?? '')
  push('service', g.service ?? '')
  push('status', g.status ?? '')
  push('logs', g.logs ?? '')

  if (statusNode) {
    push('healthy', statusNode.healthy != null ? String(statusNode.healthy) : '')
    if (statusNode.error_rate != null) {
      push('error_rate', statusNode.error_rate)
    }
    if (statusNode.last_deploy) {
      const ld = statusNode.last_deploy
      push('last_deploy.sha', ld.sha ?? '')
      push('last_deploy.ago', ld.ago ?? '')
    }
    if (statusNode.daily_tokens != null) {
      push('daily_tokens', statusNode.daily_tokens)
    }
    if (statusNode.daily_cost_usd != null) {
      push('daily_cost_usd', statusNode.daily_cost_usd)
    }
    if (Array.isArray(statusNode.recent_errors) && statusNode.recent_errors.length > 0) {
      push('recent_errors', statusNode.recent_errors.join('; '))
    }
  }

  return rows
}

function pickLabel(graphNode) {
  const g = graphNode ?? {}
  if (g.service && String(g.service).trim() !== '') {
    return g.service
  }
  if (g.project && String(g.project).trim() !== '') {
    return g.project
  }
  return g.name ?? ''
}

function pickUrl(graphNode) {
  const g = graphNode ?? {}
  const parts = [g.logs, g.project, g.service].filter(Boolean)
  return parts[0] != null ? String(parts[0]) : ''
}

/**
 * @param {object} graphJson — `perch graph --json` payload
 * @param {object} statusJson — `perch status --json` payload
 */
export function mapGraphToNodes(graphJson, statusJson) {
  const nodes = graphJson?.nodes ?? []
  const byName = new Map()
  for (const row of statusJson?.nodes ?? []) {
    if (row?.name != null) {
      byName.set(row.name, row)
    }
  }

  return nodes.map((gn) => {
    const st = byName.get(gn.name)
    const provider = gn.provider ?? ''
    return {
      id: gn.name,
      provider,
      label: pickLabel(gn),
      url: pickUrl(gn),
      status: deriveStatus(gn, st),
      meta: buildMeta(gn, st),
      logs: gn.logs != null ? String(gn.logs) : '',
      statusCommand: gn.status != null ? String(gn.status) : '',
      recentErrors: Array.isArray(st?.recent_errors) ? st.recent_errors : [],
      tabs: tabsForProvider(provider),
    }
  })
}

/**
 * @param {object} graphJson
 */
export function mapGraphToEdges(graphJson) {
  const edges = graphJson?.edges ?? []
  return edges.map((e, i) => ({
    id: `e-${e.from}-${e.to}-${i}`,
    source: e.from,
    target: e.to,
    animated: true,
  }))
}
