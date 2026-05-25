import { X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { buildErrorPrompt, isErroredNode, openInAIWithFallback } from '../lib/aiHandoff.js'
import { DeployRow } from './DeployRow.jsx'

function providerTitle(provider) {
  const map = {
    vercel: 'Vercel',
    supabase: 'Supabase',
    github: 'GitHub',
    render: 'Render',
    custom: 'Custom',
  }
  return map[provider] ?? provider
}

function metaLookup(meta, key) {
  return meta?.find((m) => m.key === key)?.value ?? '—'
}

/** @param {{ key: string, value: string }[] | undefined} meta */
function metaValue(meta, key) {
  const v = meta?.find((m) => m.key === key)?.value
  if (v == null) {
    return ''
  }
  return String(v)
}

/**
 * @param {{ meta?: { key: string, value: string }[], recentErrors?: string[] } | null} node
 * @param {string} environment
 */
function buildDeploymentRows(node, environment) {
  if (!node) {
    return []
  }
  const sha = metaValue(node.meta, 'last_deploy.sha').trim()
  const ago = metaValue(node.meta, 'last_deploy.ago').trim()
  const hasLastDeploy = sha !== '' || ago !== ''

  const rows = []
  if (hasLastDeploy) {
    rows.push({
      id: sha !== '' ? sha.slice(0, 7) : 'deploy',
      env: environment,
      current: true,
      branch: '—',
      commit: sha !== '' ? sha : '—',
      date: ago !== '' ? ago : '—',
      duration: '—',
    })
  }

  const errs = Array.isArray(node.recentErrors) ? node.recentErrors : []
  errs.forEach((msg, i) => {
    const t = String(msg ?? '').trim()
    if (t === '') {
      return
    }
    rows.push({
      id: `recent-${i}-${t.slice(0, 12)}`,
      env: 'Recent activity',
      current: false,
      branch: '—',
      commit: t,
      date: '—',
      duration: '—',
    })
  })

  return rows
}

const statusLabel = {
  healthy: 'healthy',
  degraded: 'degraded',
  down: 'down',
  source: 'source',
  unknown: 'unknown',
}

/** @param {{ text: string, copied: boolean, onCopy: () => void }} props */
function CopyBlock({ label, text, copied, onCopy }) {
  const trimmed = text.trim()
  if (trimmed === '') {
    return null
  }
  return (
    <div className="mt-3 first:mt-0">
      <div className="text-[11px] font-medium text-gray-600">{label}</div>
      <pre className="mt-1 max-h-[28vh] overflow-auto whitespace-pre-wrap break-all rounded-md border border-gray-200 bg-gray-50 p-2 font-mono text-[11px] leading-relaxed text-gray-900">
        {trimmed}
      </pre>
      <button
        type="button"
        onClick={onCopy}
        className="mt-1.5 rounded-md border border-gray-300 bg-white px-2.5 py-1 text-xs font-medium text-gray-800 hover:bg-gray-50"
      >
        {copied ? 'Copied' : 'Copy'}
      </button>
    </div>
  )
}

async function fetchLogs(env, nodeName) {
  const url = `/api/logs?env=${encodeURIComponent(env)}&node=${encodeURIComponent(nodeName)}`
  const res = await fetch(url, { headers: { Accept: 'application/json' } })
  const text = await res.text()
  let body = null
  try {
    body = text !== '' ? JSON.parse(text) : null
  } catch {
    body = null
  }
  if (!res.ok) {
    const msg = body?.error ?? res.statusText ?? 'request failed'
    throw new Error(typeof msg === 'string' ? msg : JSON.stringify(msg))
  }
  return {
    stdoutLines: Array.isArray(body?.stdout_lines) ? body.stdout_lines.map((s) => String(s ?? '')) : [],
    stderrLines: Array.isArray(body?.stderr_lines) ? body.stderr_lines.map((s) => String(s ?? '')) : [],
    exitCode: Number.isInteger(body?.exit_code) ? body.exit_code : 0,
    truncated: Boolean(body?.truncated),
    timedOut: Boolean(body?.timed_out),
    runError: body?.run_error != null ? String(body.run_error) : '',
  }
}

/** @param {{ node: object | null, environment: string }} props */
export function DetailPanel({ node, environment }) {
  const { stackName, nodeId } = useParams()
  const navigate = useNavigate()
  const [tab, setTab] = useState('deployments')
  const [copyWhich, setCopyWhich] = useState(null)
  const [logsLoading, setLogsLoading] = useState(false)
  const [logsError, setLogsError] = useState('')
  const [logsResult, setLogsResult] = useState(null)
  const [handoffStatus, setHandoffStatus] = useState('')
  const [openInMenuOpen, setOpenInMenuOpen] = useState(false)

  const close = () => {
    navigate(`/stack/${stackName}`)
  }

  useEffect(() => {
    const onKey = (e) => {
      if (e.key === 'Escape') navigate(`/stack/${stackName}`)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [navigate, stackName])

  useEffect(() => {
    setCopyWhich(null)
    setHandoffStatus('')
    setOpenInMenuOpen(false)
  }, [node?.id, tab])

  useEffect(() => {
    setLogsLoading(false)
    setLogsError('')
    setLogsResult(null)
  }, [node?.id, environment])

  const deployments = node ? buildDeploymentRows(node, environment) : []
  const logsCmd = node?.logs != null ? String(node.logs).trim() : ''
  const statusCmd = node?.statusCommand != null ? String(node.statusCommand).trim() : ''
  const hasLogsCommand = logsCmd !== ''
  const hasStatusCommand = statusCmd !== ''
  const showAIActions = node != null && isErroredNode(node)

  const region = node ? metaLookup(node.meta, 'region') : '—'
  const branch = node ? metaLookup(node.meta, 'branch') : '—'
  const st = node ? statusLabel[node.status] ?? node.status : '—'
  const project = node ? metaLookup(node.meta, 'project') : '—'
  const service = node ? metaLookup(node.meta, 'service') : '—'
  const errRate = node ? metaValue(node.meta, 'error_rate').trim() : ''
  const dailyCost = node ? metaValue(node.meta, 'daily_cost_usd').trim() : ''

  const copyText = async (which, text) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopyWhich(which)
      window.setTimeout(() => setCopyWhich(null), 2000)
    } catch {
      setCopyWhich(null)
    }
  }

  const runLogs = async () => {
    if (!node?.id) {
      return
    }
    setLogsLoading(true)
    setLogsError('')
    try {
      const next = await fetchLogs(environment, node.id)
      setLogsResult(next)
    } catch (e) {
      setLogsError(e instanceof Error ? e.message : String(e))
      setLogsResult(null)
    } finally {
      setLogsLoading(false)
    }
  }

  const openAI = async (tool) => {
    if (!node) {
      return
    }
    const prompt = buildErrorPrompt({ node, environment, logsResult })
    try {
      const result = await openInAIWithFallback({ tool, prompt })
      if (result.mode === 'deep_link') {
        setHandoffStatus(`Opened ${tool}`)
      } else {
        setHandoffStatus(`Prompt copied for ${tool}`)
      }
    } catch {
      setHandoffStatus(`Could not open ${tool}; copy failed`)
    }
    setOpenInMenuOpen(false)
    window.setTimeout(() => setHandoffStatus(''), 2500)
  }

  const recentList = node && Array.isArray(node.recentErrors) ? node.recentErrors.filter((s) => String(s).trim() !== '') : []

  return (
    <aside className="flex h-full w-[300px] shrink-0 flex-col border-l border-gray-200 bg-white">
      <div className="flex items-start justify-between gap-2 border-b border-gray-200 px-3 py-3">
        <div className="min-w-0">
          <div className="text-[13px] font-medium leading-tight text-black">
            {node ? node.label : 'Not found'}
          </div>
          <div className="mt-0.5 text-[11px] text-gray-400">
            {node ? (
              <>
                {providerTitle(node.provider)} · {environment}
              </>
            ) : (
              <span className="font-mono text-gray-500">{nodeId}</span>
            )}
          </div>
        </div>
        <button
          type="button"
          onClick={close}
          className="rounded-md p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-900"
          aria-label="Close panel"
        >
          <X className="h-4 w-4" strokeWidth={2} />
        </button>
      </div>

      {node && (
        <div className="grid grid-cols-3 gap-2 border-b border-gray-200 px-3 py-2">
          <div>
            <div className="text-[10px] uppercase tracking-wide text-gray-400">region</div>
            <div className="truncate text-xs font-medium text-gray-900">{region}</div>
          </div>
          <div>
            <div className="text-[10px] uppercase tracking-wide text-gray-400">status</div>
            <div className="truncate text-xs font-medium text-gray-900">{st}</div>
          </div>
          <div>
            <div className="text-[10px] uppercase tracking-wide text-gray-400">branch</div>
            <div className="truncate text-xs font-medium text-gray-900">{branch}</div>
          </div>
        </div>
      )}

      {node && (project !== '—' || service !== '—' || errRate !== '' || dailyCost !== '' || recentList.length > 0) && (
        <div className="space-y-1.5 border-b border-gray-200 px-3 py-2 text-[11px] text-gray-600">
          {(project !== '—' || service !== '—') && (
            <div className="flex flex-wrap gap-x-2 gap-y-0.5">
              {project !== '—' && (
                <span>
                  <span className="text-gray-400">project</span> {project}
                </span>
              )}
              {service !== '—' && (
                <span>
                  <span className="text-gray-400">service</span> {service}
                </span>
              )}
            </div>
          )}
          {(errRate !== '' || dailyCost !== '') && (
            <div className="flex flex-wrap gap-x-2">
              {errRate !== '' && (
                <span>
                  <span className="text-gray-400">error rate</span> {errRate}
                </span>
              )}
              {dailyCost !== '' && (
                <span>
                  <span className="text-gray-400">daily $</span> {dailyCost}
                </span>
              )}
            </div>
          )}
          {recentList.length > 0 && (
            <div>
              <div className="mb-0.5 text-[10px] uppercase tracking-wide text-gray-400">recent errors</div>
              <ul className="list-inside list-disc space-y-0.5 text-gray-800">
                {recentList.slice(0, 8).map((line, i) => (
                  <li key={`${i}-${String(line).slice(0, 24)}`} className="break-words">
                    {String(line)}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {!node && (
        <div className="border-b border-gray-200 px-3 py-3 text-sm text-gray-500">
          No node exists for this id in the stack graph.
        </div>
      )}

      {node && (
        <>
          <div className="flex border-b border-gray-200 text-xs text-gray-500">
            {[
              { id: 'deployments', label: 'deployments' },
              { id: 'logs', label: 'logs' },
              { id: 'env', label: 'env vars' },
            ].map((t) => (
              <button
                key={t.id}
                type="button"
                onClick={() => setTab(t.id)}
                className={`flex-1 border-r border-gray-200 px-2 py-2 capitalize last:border-r-0 hover:bg-gray-50 hover:text-gray-900 ${
                  tab === t.id ? 'bg-gray-50 text-gray-900' : ''
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto">
            {tab === 'deployments' && (
              <div>
                {deployments.length === 0 ? (
                  <div className="space-y-2 p-3 text-sm text-gray-600">
                    <p className="font-medium text-gray-800">No deployment data</p>
                    {(node.provider ?? '').toLowerCase() === 'custom' ? (
                      <p>
                        Custom nodes only expose health from the <span className="font-mono text-xs">status:</span> shell
                        exit code right now. Last deploy and recent errors are not filled for this provider yet.
                      </p>
                    ) : (
                      <p>
                        Deployable providers (Vercel, Render, …) still return a placeholder from{' '}
                        <span className="font-mono text-xs">perch status</span> in this build, so{' '}
                        <span className="font-mono text-xs">last_deploy</span> and errors do not appear here.
                      </p>
                    )}
                    <p className="text-xs text-gray-500">
                      In <span className="font-mono">examples/scenarios/full-stack</span>, switch the header environment
                      to <span className="font-semibold text-gray-700">dev</span> for demo{' '}
                      <span className="font-mono text-[11px]">custom</span> nodes. Production{' '}
                      <span className="font-mono text-[11px]">web</span> is Vercel YAML only (no shell hooks).
                    </p>
                  </div>
                ) : (
                  deployments.map((d, i) => <DeployRow key={`${node.id}-row-${i}`} deployment={d} />)
                )}
              </div>
            )}

            {tab === 'logs' && (
              <div className="p-3 text-sm text-gray-700">
                <p className="text-xs text-gray-500">
                  Commands from perch.yaml (same as the TUI). Run locally in your shell; nothing is executed from the
                  browser.
                </p>
                <div className="mt-2 flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={() => void runLogs()}
                    disabled={!node?.id || logsLoading}
                    className="rounded-md border border-gray-300 bg-white px-2.5 py-1 text-xs font-medium text-gray-800 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {logsLoading ? 'Running...' : 'Run logs'}
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      const prompt = buildErrorPrompt({ node, environment, logsResult })
                      void copyText('prompt', prompt)
                    }}
                    className="rounded-md border border-gray-300 bg-white px-2.5 py-1 text-xs font-medium text-gray-800 hover:bg-gray-50"
                  >
                    {copyWhich === 'prompt' ? 'Prompt copied' : 'Copy prompt'}
                  </button>
                </div>
                {showAIActions && (
                  <div className="relative mt-2 inline-block">
                    <button
                      type="button"
                      onClick={() => setOpenInMenuOpen((v) => !v)}
                      className="rounded-md border border-gray-300 bg-white px-2.5 py-1 text-xs font-medium text-gray-800 hover:bg-gray-50"
                    >
                      Open in
                    </button>
                    {openInMenuOpen && (
                      <div className="absolute z-20 mt-1 min-w-[180px] rounded-md border border-gray-200 bg-white p-1 shadow-lg">
                        <button
                          type="button"
                          onClick={() => void openAI('codex')}
                          className="block w-full rounded px-2 py-1.5 text-left text-xs text-gray-800 hover:bg-gray-50"
                        >
                          Codex
                        </button>
                        <button
                          type="button"
                          onClick={() => void openAI('cursor')}
                          className="block w-full rounded px-2 py-1.5 text-left text-xs text-gray-800 hover:bg-gray-50"
                        >
                          Cursor
                        </button>
                        <button
                          type="button"
                          onClick={() => void openAI('claude')}
                          className="block w-full rounded px-2 py-1.5 text-left text-xs text-gray-800 hover:bg-gray-50"
                        >
                          Claude Code
                        </button>
                      </div>
                    )}
                  </div>
                )}
                {handoffStatus !== '' && <p className="mt-2 text-xs text-gray-500">{handoffStatus}</p>}
                {logsError !== '' && <p className="mt-2 text-xs text-red-600">{logsError}</p>}
                {logsResult && (
                  <div className="mt-3 space-y-2">
                    {logsResult.runError !== '' && (
                      <p className="rounded-md border border-red-200 bg-red-50 px-2 py-1 text-xs text-red-700">
                        {logsResult.runError}
                      </p>
                    )}
                    {(logsResult.timedOut || logsResult.truncated) && (
                      <p className="rounded-md border border-amber-200 bg-amber-50 px-2 py-1 text-xs text-amber-800">
                        {logsResult.timedOut ? 'Command timed out. ' : ''}
                        {logsResult.truncated ? 'Output truncated.' : ''}
                      </p>
                    )}
                    <div className="text-xs text-gray-500">Exit code: {logsResult.exitCode}</div>
                    <CopyBlock
                      label="stderr"
                      text={logsResult.stderrLines.join('\n')}
                      copied={copyWhich === 'stderr-out'}
                      onCopy={() => void copyText('stderr-out', logsResult.stderrLines.join('\n'))}
                    />
                    <CopyBlock
                      label="stdout"
                      text={logsResult.stdoutLines.join('\n')}
                      copied={copyWhich === 'stdout-out'}
                      onCopy={() => void copyText('stdout-out', logsResult.stdoutLines.join('\n'))}
                    />
                    {logsResult.stdoutLines.length === 0 && logsResult.stderrLines.length === 0 && (
                      <p className="text-xs text-gray-500">No output returned for this run.</p>
                    )}
                  </div>
                )}
                {hasStatusCommand && (
                  <CopyBlock
                    label="Health check (status:)"
                    text={statusCmd}
                    copied={copyWhich === 'status'}
                    onCopy={() => void copyText('status', statusCmd)}
                  />
                )}
                {hasLogsCommand && (
                  <CopyBlock
                    label="Tail logs (logs:)"
                    text={logsCmd}
                    copied={copyWhich === 'logs'}
                    onCopy={() => void copyText('logs', logsCmd)}
                  />
                )}
                {!hasStatusCommand && !hasLogsCommand && (
                  <div className="mt-2 space-y-2 text-sm text-gray-600">
                    <p>No <span className="font-mono text-xs">status:</span> or <span className="font-mono text-xs">logs:</span> shell lines for this node in perch.yaml.</p>
                    {(node.provider ?? '').toLowerCase() !== 'custom' && (
                      <p className="text-xs text-gray-500">
                        Platform nodes use provider APIs, not YAML shell hooks. Try environment{' '}
                        <span className="font-semibold text-gray-700">dev</span> in this example stack to see sample{' '}
                        <span className="font-mono text-[11px]">status:</span> / <span className="font-mono text-[11px]">logs:</span>{' '}
                        on <span className="font-mono text-[11px]">web</span>, <span className="font-mono text-[11px]">api</span>, etc.
                      </p>
                    )}
                  </div>
                )}
              </div>
            )}

            {tab === 'env' && (
              <p className="p-3 text-sm italic text-gray-500">
                env vars are redacted — open provider dashboard to reveal
              </p>
            )}
          </div>
        </>
      )}
    </aside>
  )
}
