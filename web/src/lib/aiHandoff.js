const MAX_PROMPT_LINES = 40

function trimLines(lines) {
  const normalized = Array.isArray(lines)
    ? lines.map((line) => String(line ?? '').trim()).filter((line) => line !== '')
    : []
  if (normalized.length <= MAX_PROMPT_LINES) {
    return normalized
  }
  return normalized.slice(normalized.length - MAX_PROMPT_LINES)
}

export function isErroredNode(node) {
  const status = String(node?.status ?? '').toLowerCase()
  if (status === 'down' || status === 'degraded') {
    return true
  }
  return Array.isArray(node?.recentErrors) && node.recentErrors.some((line) => String(line ?? '').trim() !== '')
}

export function buildErrorPrompt({ node, environment, logsResult }) {
  const label = String(node?.label ?? node?.id ?? node?.provider ?? 'unknown-node')
  const provider = String(node?.provider ?? 'unknown')
  const env = String(environment ?? node?.environment ?? 'unknown')
  const recentErrors = trimLines(node?.recentErrors ?? [])
  const stderrTail = trimLines(logsResult?.stderrLines ?? [])
  const stdoutTail = trimLines(logsResult?.stdoutLines ?? [])

  const sections = [
    `You are debugging a failing service node in my stack.`,
    ``,
    `Node: ${label}`,
    `Provider: ${provider}`,
    `Environment: ${env}`,
    ``,
    `Please analyze these logs and tell me:`,
    `1) most likely root cause`,
    `2) related systems/dependencies to inspect`,
    `3) concrete next checks and fixes (ordered)`,
    ``,
    `Recent error lines:`,
    recentErrors.length === 0 ? '(none)' : recentErrors.join('\n'),
    ``,
    `stderr tail:`,
    stderrTail.length === 0 ? '(none)' : stderrTail.join('\n'),
    ``,
    `stdout tail:`,
    stdoutTail.length === 0 ? '(none)' : stdoutTail.join('\n'),
  ]
  return sections.join('\n').trim()
}

const deepLinkBuilders = {
  cursor: (prompt) => `cursor://ai/new?prompt=${encodeURIComponent(prompt)}`,
  codex: (prompt) => `codex://new?prompt=${encodeURIComponent(prompt)}`,
  claude: (prompt) => `claude://new?prompt=${encodeURIComponent(prompt)}`,
}

/**
 * Deep-link first; fallback to clipboard.
 * @param {{ tool: 'cursor'|'codex'|'claude', prompt: string, openWindow?: Function, copyText?: Function }} input
 */
export async function openInAIWithFallback({ tool, prompt, openWindow, copyText }) {
  const deepLink = deepLinkBuilders[tool]?.(prompt)
  const openFn = openWindow ?? ((url) => window.open(url, '_blank', 'noopener,noreferrer'))
  const copyFn = copyText ?? ((text) => navigator.clipboard.writeText(text))

  if (deepLink) {
    try {
      const opened = openFn(deepLink)
      if (opened != null) {
        return { mode: 'deep_link', deepLink }
      }
    } catch {
      // fallback below
    }
  }
  await copyFn(prompt)
  return { mode: 'copied', deepLink: deepLink ?? null }
}
